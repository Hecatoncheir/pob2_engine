package application

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"regexp"

	"github.com/Hecatoncheir/pob2_engine/internal/build"
	"github.com/Hecatoncheir/pob2_engine/internal/calculation"
	"github.com/Hecatoncheir/pob2_engine/internal/domain"
)

type BuildView struct {
	XML                    string              `json:"xml"`
	InputSHA256            string              `json:"input_sha256"`
	InputRevision          uint64              `json:"input_revision"`
	Model                  build.Model         `json:"model"`
	Calculation            *calculation.Result `json:"calculation,omitempty"`
	CalculatedFromRevision uint64              `json:"calculated_from_revision,omitempty"`
}
type SnapshotView struct {
	SessionID string     `json:"session_id"`
	BuildName string     `json:"build_name"`
	Revision  uint64     `json:"revision"`
	Build     *BuildView `json:"build,omitempty"`
}

var inputHash = regexp.MustCompile(`^[a-f0-9]{64}$`)
var sourceCommit = regexp.MustCompile(`^[a-f0-9]{40}$`)

func DecodeCalculation(snapshot domain.Snapshot) (*calculation.Result, error) {
	if snapshot.CalculationJSON == "" {
		if snapshot.CalculatedFromRevision != 0 {
			return nil, fmt.Errorf("нет сохранённого расчёта")
		}
		return nil, nil
	}
	var result calculation.Result
	decoder := json.NewDecoder(bytes.NewBufferString(snapshot.CalculationJSON))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&result); err != nil {
		return nil, err
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return nil, fmt.Errorf("лишние данные расчёта")
	}
	if result.InputSHA256 != snapshot.InputSHA256 || !inputHash.MatchString(result.InputSHA256) ||
		!hasCalculationSource(result) ||
		snapshot.BuildRevision == 0 || snapshot.CalculatedFromRevision != snapshot.BuildRevision {
		return nil, fmt.Errorf("результат не соответствует входу или данным")
	}
	if err := validateCalculationStats(result); err != nil {
		return nil, err
	}
	return &result, nil
}

func validateCalculationStats(result calculation.Result) error {
	switch result.Status {
	case "complete", "partial":
		if len(result.Stats) == 0 {
			return fmt.Errorf("нет числовых результатов")
		}
	case "unsupported":
		if result.Stats != nil {
			return fmt.Errorf("неподдержанный расчёт содержит числа")
		}
	default:
		return fmt.Errorf("неизвестный статус расчёта")
	}
	for _, value := range result.Stats {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return fmt.Errorf("неверное число расчёта")
		}
	}
	return nil
}

func hasCalculationSource(result calculation.Result) bool {
	if result.DataID == calculation.LocalPoBDataID {
		return result.DataSHA256 == "" && result.PoBCommit == ""
	}
	return result.DataID != "" && inputHash.MatchString(result.DataSHA256) && sourceCommit.MatchString(result.PoBCommit)
}

func ViewSnapshot(snapshot domain.Snapshot) (SnapshotView, error) {
	view := SnapshotView{SessionID: snapshot.SessionID, BuildName: snapshot.BuildName, Revision: snapshot.Revision}
	if snapshot.BuildXML == "" {
		if snapshot.InputSHA256 != "" || snapshot.BuildRevision != 0 || snapshot.CalculationJSON != "" || snapshot.CalculatedFromRevision != 0 {
			return view, fmt.Errorf("нет XML для сохранённого состояния сборки")
		}
		return view, nil
	}
	doc, err := build.Parse([]byte(snapshot.BuildXML))
	if err != nil {
		return view, err
	}
	model, err := doc.Model()
	if err != nil {
		return view, err
	}
	digest, err := doc.InputSHA256()
	if err != nil {
		return view, err
	}
	if digest != snapshot.InputSHA256 || snapshot.BuildRevision == 0 || snapshot.BuildRevision > snapshot.Revision {
		return view, fmt.Errorf("неверная привязка XML к ревизии")
	}
	calculated, err := DecodeCalculation(snapshot)
	if err != nil {
		return view, err
	}
	view.Build = &BuildView{XML: snapshot.BuildXML, InputSHA256: snapshot.InputSHA256, InputRevision: snapshot.BuildRevision,
		Model: model, Calculation: calculated, CalculatedFromRevision: snapshot.CalculatedFromRevision}
	return view, nil
}

func (s *Service) restoreCalculation(ctx context.Context, cmd CalculateBuild, stored domain.Snapshot) (domain.Snapshot, error) {
	if err := validate(cmd.SessionID, "build", cmd.RequestID); err != nil {
		return domain.Snapshot{}, err
	}
	fp, err := fingerprint("calculate", cmd)
	if err != nil {
		return domain.Snapshot{}, err
	}
	value, err := s.find(cmd.SessionID)
	if err != nil {
		return domain.Snapshot{}, err
	}
	value.mu.Lock()
	defer value.mu.Unlock()
	if _, ok, err := requestResult(value, cmd.RequestID, fp); ok || err != nil {
		return domain.Snapshot{}, fmt.Errorf("повтор расчёта в журнале")
	}
	if err := checkRevision(ctx, value, cmd.ExpectedRevision); err != nil {
		return domain.Snapshot{}, err
	}
	expected := value.snapshot
	expected.Revision++
	expected.CalculationJSON = stored.CalculationJSON
	expected.CalculatedFromRevision = expected.BuildRevision
	if expected != stored {
		return domain.Snapshot{}, fmt.Errorf("расчёт изменяет вход или неверную ревизию")
	}
	if _, err := DecodeCalculation(stored); err != nil {
		return domain.Snapshot{}, err
	}
	if stored.CalculationJSON == "" {
		return domain.Snapshot{}, fmt.Errorf("пустой результат расчёта в журнале")
	}
	if err := s.commitBuild(ctx, value, cmd.RequestID, fp, "calculated", "calculation", Change{Calculate: &cmd, Result: stored}, int64(len(fp)+len(stored.CalculationJSON)+1024)); err != nil {
		return domain.Snapshot{}, err
	}
	return stored, nil
}
