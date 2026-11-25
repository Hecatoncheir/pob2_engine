package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Hecatoncheir/pob2_engine/internal/build"
	"github.com/Hecatoncheir/pob2_engine/internal/calculation"
	"github.com/Hecatoncheir/pob2_engine/internal/domain"
	"github.com/Hecatoncheir/pob2_engine/internal/gamedata"
)

type ApplyBuild struct {
	SessionID        string      `json:"session_id"`
	ExpectedRevision uint64      `json:"expected_revision"`
	RequestID        string      `json:"request_id"`
	XML              string      `json:"xml,omitempty"`
	Edit             *build.Edit `json:"edit,omitempty"`
}
type CalculateBuild struct {
	SessionID        string `json:"session_id"`
	ExpectedRevision uint64 `json:"expected_revision"`
	RequestID        string `json:"request_id"`
}
type CompareBuild struct {
	SessionID        string      `json:"session_id"`
	ExpectedRevision uint64      `json:"expected_revision"`
	XML              string      `json:"xml,omitempty"`
	Edit             *build.Edit `json:"edit,omitempty"`
}
type BuildComparison struct {
	SessionID  string                 `json:"session_id"`
	Revision   uint64                 `json:"revision"`
	Comparison calculation.Comparison `json:"comparison"`
	Passives   *build.PassiveChange   `json:"passive_change,omitempty"`
}
type ExportedBuild struct {
	SessionID   string `json:"session_id"`
	Revision    uint64 `json:"revision"`
	XML         string `json:"xml"`
	InputSHA256 string `json:"input_sha256"`
}

type BuildCommands interface {
	Commands
	Apply(context.Context, ApplyBuild) (domain.Snapshot, error)
	Calculate(context.Context, CalculateBuild) (domain.Snapshot, error)
	Compare(context.Context, CompareBuild) (BuildComparison, error)
	Export(context.Context, string) (ExportedBuild, error)
}

type buildRequest struct {
	fingerprint string
	result      domain.Snapshot
}
type pendingCalculation struct {
	fingerprint string
	done        chan struct{}
	result      domain.Snapshot
	err         error
}

func fingerprint(kind string, command any) (string, error) {
	data, err := json.Marshal(command)
	if err != nil {
		return "", failure(InvalidArgument, err.Error())
	}
	return kind + ":" + string(data), nil
}

func (s *Service) loadData() error {
	s.dataOnce.Do(func() {
		s.data, s.dataError = gamedata.Embedded()
	})
	if s.dataError != nil {
		return failure(CalculationFailure, fmt.Sprintf("не удалось загрузить набор данных: %v", s.dataError))
	}
	return nil
}

func requestResult(value *session, requestID, fp string) (domain.Snapshot, bool, error) {
	if requestID == value.createRequestID {
		return domain.Snapshot{}, false, failure(RequestConflict, "request_id использован для создания сессии")
	}
	if _, ok := value.requests[requestID]; ok {
		return domain.Snapshot{}, false, failure(RequestConflict, "request_id использован для переименования")
	}
	previous, ok := value.buildRequests[requestID]
	if !ok {
		return domain.Snapshot{}, false, nil
	}
	if previous.fingerprint != fp {
		return domain.Snapshot{}, false, failure(RequestConflict, "request_id использован для другой команды")
	}
	return previous.result, true, nil
}

func checkRevision(ctx context.Context, value *session, expected uint64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if value.snapshot.Revision != expected {
		return failure(RevisionConflict, fmt.Sprintf("ожидалась ревизия %d, текущая %d", expected, value.snapshot.Revision))
	}
	if len(value.requests)+len(value.buildRequests) >= maxRequests {
		return failure(ResourceExhausted, "достигнут лимит команд сессии")
	}
	return nil
}

const retainedLimit = 128 << 20

func (s *Service) reserve(size int64) bool {
	for {
		current := s.retained.Load()
		if size < 0 || size > retainedLimit-current {
			return false
		}
		if s.retained.CompareAndSwap(current, current+size) {
			return true
		}
	}
}

func (s *Service) commitBuild(ctx context.Context, value *session, requestID, fp, kind, fields string, change Change, newBytes int64) error {
	if !s.reserve(newBytes) {
		return failure(ResourceExhausted, "достигнут лимит истории сборок в памяти")
	}
	if err := s.persist(ctx, change); err != nil {
		s.retained.Add(-newBytes)
		return err
	}
	value.snapshot = change.Result
	value.buildRequests[requestID] = buildRequest{fingerprint: fp, result: change.Result}
	value.events = append(value.events, Event{Snapshot: change.Result, Kind: kind, RequestID: requestID, ChangedFields: fields})
	close(value.changed)
	value.changed = make(chan struct{})
	return nil
}

func (s *Service) Apply(ctx context.Context, cmd ApplyBuild) (domain.Snapshot, error) {
	if err := validateApply(ctx, cmd); err != nil {
		return domain.Snapshot{}, err
	}
	fp, err := fingerprint("apply", cmd)
	if err != nil {
		return domain.Snapshot{}, err
	}
	value, err := s.find(cmd.SessionID)
	if err != nil {
		return domain.Snapshot{}, err
	}
	value.mu.Lock()
	before, replay, err := applySnapshot(ctx, value, cmd, fp)
	value.mu.Unlock()
	if replay || err != nil {
		return before, err
	}
	result, fields, prepareErr := s.prepareApply(ctx, cmd, before)
	value.mu.Lock()
	defer value.mu.Unlock()
	// Другая попытка могла подтвердить этот запрос, пока готовилась правка.
	if previous, replay, err := applySnapshot(ctx, value, cmd, fp); replay || err != nil {
		return previous, err
	}
	if prepareErr != nil {
		return domain.Snapshot{}, prepareErr
	}
	change := Change{Apply: &cmd, Result: result}
	if err := s.commitBuild(ctx, value, cmd.RequestID, fp, "build_changed", fields, change, int64(len(fp)+len(result.BuildXML)+1024)); err != nil {
		return domain.Snapshot{}, err
	}
	return result, nil
}

func validateApply(ctx context.Context, cmd ApplyBuild) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := validate(cmd.SessionID, "build", cmd.RequestID); err != nil {
		return err
	}
	if (cmd.XML == "") == (cmd.Edit == nil) {
		return failure(InvalidArgument, "нужен XML либо одна правка")
	}
	return nil
}

// applySnapshot проверяет команду и читает снимок под уже взятой блокировкой сессии.
func applySnapshot(ctx context.Context, value *session, cmd ApplyBuild, fp string) (domain.Snapshot, bool, error) {
	if err := ctx.Err(); err != nil {
		return domain.Snapshot{}, false, err
	}
	if previous, replay, err := requestResult(value, cmd.RequestID, fp); replay || err != nil {
		return previous, replay, err
	}
	if _, ok := value.pending[cmd.RequestID]; ok {
		return domain.Snapshot{}, false, failure(RequestConflict, "request_id занят расчётом")
	}
	if err := checkRevision(ctx, value, cmd.ExpectedRevision); err != nil {
		return domain.Snapshot{}, false, err
	}
	return value.snapshot, false, nil
}

func (s *Service) prepareApply(ctx context.Context, cmd ApplyBuild, before domain.Snapshot) (domain.Snapshot, string, error) {
	if err := s.loadData(); err != nil {
		return domain.Snapshot{}, "", err
	}
	doc, err := s.applyDocument(ctx, cmd, before.BuildXML)
	if err != nil {
		return domain.Snapshot{}, "", err
	}
	result := before
	result.Revision++
	result, err = withBuildDocument(result, doc)
	if err != nil {
		return domain.Snapshot{}, "", err
	}
	fields, err := changedBuildFields(before.BuildXML, doc)
	if err != nil {
		return domain.Snapshot{}, "", failure(InvalidArgument, err.Error())
	}
	return result, fields, nil
}

func (s *Service) applyDocument(ctx context.Context, cmd ApplyBuild, previousXML string) (*build.Document, error) {
	xml := cmd.XML
	if cmd.Edit != nil {
		xml = previousXML
	}
	doc, err := build.Parse([]byte(xml))
	if err != nil {
		return nil, failure(InvalidArgument, err.Error())
	}
	if cmd.Edit == nil {
		return doc, nil
	}
	doc, _, err = s.editDocument(ctx, doc, *cmd.Edit)
	var app *Error
	switch {
	case err == nil:
		return doc, nil
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded), errors.As(err, &app):
		return nil, err
	default:
		return nil, failure(InvalidArgument, err.Error())
	}
}

func withBuildDocument(snapshot domain.Snapshot, doc *build.Document) (domain.Snapshot, error) {
	xml, err := doc.Export()
	if err != nil {
		return domain.Snapshot{}, failure(InvalidArgument, err.Error())
	}
	digest, err := doc.InputSHA256()
	if err != nil {
		return domain.Snapshot{}, failure(InvalidArgument, err.Error())
	}
	snapshot.BuildRevision = snapshot.Revision
	snapshot.BuildXML = string(xml)
	snapshot.InputSHA256 = digest
	snapshot.CalculationJSON = ""
	snapshot.CalculatedFromRevision = 0
	return snapshot, nil
}

func (s *Service) Export(ctx context.Context, id string) (ExportedBuild, error) {
	snapshot, err := s.Snapshot(ctx, id)
	if err != nil {
		return ExportedBuild{}, err
	}
	if snapshot.BuildXML == "" {
		return ExportedBuild{}, failure(InvalidArgument, "сначала импортируйте XML сборки")
	}
	return ExportedBuild{SessionID: id, Revision: snapshot.Revision, XML: snapshot.BuildXML, InputSHA256: snapshot.InputSHA256}, nil
}

func (s *Service) Compare(ctx context.Context, cmd CompareBuild) (BuildComparison, error) {
	if (cmd.XML == "") == (cmd.Edit == nil) {
		return BuildComparison{}, failure(InvalidArgument, "нужен XML варианта либо одна правка")
	}
	snapshot, err := s.Snapshot(ctx, cmd.SessionID)
	if err != nil {
		return BuildComparison{}, err
	}
	if snapshot.Revision != cmd.ExpectedRevision {
		return BuildComparison{}, failure(RevisionConflict, "снимок для сравнения устарел")
	}
	if err := s.loadData(); err != nil {
		return BuildComparison{}, err
	}
	release, err := s.acquire(ctx)
	if err != nil {
		return BuildComparison{}, err
	}
	defer release()
	before, err := build.Parse([]byte(snapshot.BuildXML))
	if err != nil {
		return BuildComparison{}, failure(InvalidArgument, err.Error())
	}
	var after *build.Document
	var passives *build.PassiveChange
	if cmd.Edit != nil {
		after, passives, err = s.editDocument(ctx, before, *cmd.Edit)
	} else {
		after, err = build.Parse([]byte(cmd.XML))
	}
	if err != nil {
		return BuildComparison{}, failure(InvalidArgument, err.Error())
	}
	if s.calculator == nil {
		return BuildComparison{}, failure(CalculationFailure, "модуль расчёта не настроен")
	}
	result, err := s.calculator.Compare(ctx, before, after)
	if err != nil {
		return BuildComparison{}, err
	}
	return BuildComparison{SessionID: cmd.SessionID, Revision: snapshot.Revision, Comparison: result, Passives: passives}, nil
}
