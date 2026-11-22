package application

import (
	"context"
	"fmt"

	"github.com/Hecatoncheir/pob2_engine/internal/build"
	"github.com/Hecatoncheir/pob2_engine/internal/domain"
)

func (s *Service) restoreApply(ctx context.Context, cmd ApplyBuild, stored domain.Snapshot) (domain.Snapshot, error) {
	if err := validate(cmd.SessionID, "build", cmd.RequestID); err != nil {
		return domain.Snapshot{}, err
	}
	if err := validateStoredApply(cmd, stored); err != nil {
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
	defer value.mu.Unlock()
	if _, ok, err := requestResult(value, cmd.RequestID, fp); ok || err != nil {
		return domain.Snapshot{}, fmt.Errorf("повтор правки сборки в журнале")
	}
	if err := checkRevision(ctx, value, cmd.ExpectedRevision); err != nil {
		return domain.Snapshot{}, err
	}
	expected := value.snapshot
	expected.Revision++
	expected.BuildRevision = expected.Revision
	expected.BuildXML, expected.InputSHA256 = stored.BuildXML, stored.InputSHA256
	expected.CalculationJSON, expected.CalculatedFromRevision = "", 0
	if expected != stored {
		return domain.Snapshot{}, fmt.Errorf("правка меняет посторонние поля или неверную ревизию")
	}
	return s.commitRestoredApply(ctx, value, cmd, fp, stored)
}

func (s *Service) commitRestoredApply(ctx context.Context, value *session, cmd ApplyBuild, fp string, stored domain.Snapshot) (domain.Snapshot, error) {
	doc, err := build.Parse([]byte(stored.BuildXML))
	if err != nil {
		return domain.Snapshot{}, err
	}
	fields, err := changedBuildFields(value.snapshot.BuildXML, doc)
	if err != nil {
		return domain.Snapshot{}, err
	}
	change := Change{Apply: &cmd, Result: stored}
	if err := s.commitBuild(ctx, value, cmd.RequestID, fp, "build_changed", fields, change, int64(len(fp)+len(stored.BuildXML)+1024)); err != nil {
		return domain.Snapshot{}, err
	}
	return stored, nil
}

func validateStoredApply(cmd ApplyBuild, stored domain.Snapshot) error {
	if (cmd.XML == "") == (cmd.Edit == nil) || stored.BuildXML == "" {
		return fmt.Errorf("в журнале правки нужен XML либо одна правка и сохранённый результат")
	}
	if _, err := ViewSnapshot(stored); err != nil {
		return err
	}
	if cmd.Edit != nil {
		return build.ValidateEdit(*cmd.Edit)
	}
	doc, err := build.Parse([]byte(cmd.XML))
	if err != nil {
		return err
	}
	xml, err := doc.Export()
	if err != nil {
		return err
	}
	if string(xml) != stored.BuildXML {
		return fmt.Errorf("сохранённый XML не соответствует команде импорта")
	}
	return nil
}
