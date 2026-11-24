package application

import (
	"context"
	"encoding/json"

	"github.com/Hecatoncheir/pob2_engine/internal/build"
	"github.com/Hecatoncheir/pob2_engine/internal/domain"
)

type Progress struct {
	Stage string `json:"stage"`
}
type progressKey struct{}

// WithProgress предоставляет последнее состояние; медленный читатель не задерживает расчёт.
// Канал не закрывается: окончательный результат возвращается самой командой.
func WithProgress(ctx context.Context) (context.Context, <-chan Progress) {
	updates := make(chan Progress, 1)
	return context.WithValue(ctx, progressKey{}, updates), updates
}

func progress(ctx context.Context, stage string) {
	updates, ok := ctx.Value(progressKey{}).(chan Progress)
	if !ok {
		return
	}
	select {
	case updates <- Progress{Stage: stage}:
		return
	default:
	}
	select {
	case <-updates:
	default:
	}
	select {
	case updates <- Progress{Stage: stage}:
	default:
	}
}

// ReportProgress передаёт состояние удалённой задачи наблюдателю того же контракта.
func ReportProgress(ctx context.Context, stage string) { progress(ctx, stage) }

func (s *Service) acquire(ctx context.Context) (func(), error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	select {
	case s.admission <- struct{}{}:
	default:
		return nil, failure(ResourceExhausted, "очередь расчётов заполнена")
	}
	progress(ctx, "queued")
	select {
	case s.workers <- struct{}{}:
		progress(ctx, "running")
		return func() { <-s.workers; <-s.admission }, nil
	case <-ctx.Done():
		<-s.admission
		return nil, ctx.Err()
	}
}

func (s *Service) Calculate(ctx context.Context, cmd CalculateBuild) (result domain.Snapshot, err error) {
	if err = ctx.Err(); err != nil {
		return result, err
	}
	if err = validate(cmd.SessionID, "build", cmd.RequestID); err != nil {
		return result, err
	}
	fp, err := fingerprint("calculate", cmd)
	if err != nil {
		return result, err
	}
	value, err := s.find(cmd.SessionID)
	if err != nil {
		return result, err
	}
	value.mu.Lock()
	previous, ok, err := requestResult(value, cmd.RequestID, fp)
	if ok || err != nil {
		value.mu.Unlock()
		return previous, err
	}
	if pending, ok := value.pending[cmd.RequestID]; ok {
		value.mu.Unlock()
		if pending.fingerprint != fp {
			return result, failure(RequestConflict, "request_id занят другим расчётом")
		}
		progress(ctx, "waiting")
		select {
		case <-pending.done:
			return pending.result, pending.err
		case <-ctx.Done():
			return result, ctx.Err()
		}
	}
	if err = checkRevision(ctx, value, cmd.ExpectedRevision); err != nil {
		value.mu.Unlock()
		return result, err
	}
	before := value.snapshot
	if before.BuildXML == "" {
		value.mu.Unlock()
		return result, failure(InvalidArgument, "сначала импортируйте XML сборки")
	}
	pending := &pendingCalculation{fingerprint: fp, done: make(chan struct{})}
	value.pending[cmd.RequestID] = pending
	value.mu.Unlock()
	defer func() {
		value.mu.Lock()
		pending.result, pending.err = result, err
		delete(value.pending, cmd.RequestID)
		close(pending.done)
		value.mu.Unlock()
	}()
	if err = s.loadData(); err != nil {
		return result, err
	}
	release, err := s.acquire(ctx)
	if err != nil {
		return result, err
	}
	defer release()
	doc, err := build.Parse([]byte(before.BuildXML))
	if err != nil {
		return result, failure(CalculationFailure, err.Error())
	}
	if s.calculator == nil {
		return result, failure(CalculationFailure, "модуль расчёта не настроен")
	}
	report, err := s.calculator.Calculate(ctx, doc)
	if err != nil {
		return result, err
	}
	serialized, err := json.Marshal(report)
	if err != nil {
		return result, failure(CalculationFailure, err.Error())
	}
	value.mu.Lock()
	defer value.mu.Unlock()
	if err = checkRevision(ctx, value, cmd.ExpectedRevision); err != nil {
		return result, err
	}
	result = value.snapshot
	result.Revision++
	result.CalculationJSON = string(serialized)
	result.CalculatedFromRevision = result.BuildRevision
	progress(ctx, "saving")
	if err = s.commitBuild(ctx, value, cmd.RequestID, fp, "calculated", "calculation", Change{Calculate: &cmd, Result: result}, int64(len(fp)+len(serialized)+1024)); err != nil {
		return domain.Snapshot{}, err
	}
	progress(ctx, "completed")
	return result, nil
}
