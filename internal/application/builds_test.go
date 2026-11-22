package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Hecatoncheir/pob2_engine/internal/build"
	"github.com/Hecatoncheir/pob2_engine/internal/calculation"
	"github.com/Hecatoncheir/pob2_engine/internal/domain"
)

func importedService(t *testing.T, journal Journal) (*Service, domain.Snapshot) {
	t.Helper()
	s := NewService()
	if journal != nil {
		var err error
		s, err = OpenService(t.Context(), journal)
		if err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.Create(t.Context(), CreateSession{SessionID: "main", BuildName: "Сборка", RequestID: "create"}); err != nil {
		t.Fatal(err)
	}
	xml, err := os.ReadFile("../../testdata/reference/fireball.xml")
	if err != nil {
		t.Fatal(err)
	}
	initial, err := s.Apply(t.Context(), ApplyBuild{SessionID: "main", ExpectedRevision: 1, RequestID: "import", XML: string(xml)})
	if err != nil {
		t.Fatal(err)
	}
	return s, initial
}

func TestFullBuildRecoveryRetriesAndStaleCalculation(t *testing.T) {
	journal := &testJournal{}
	s, imported := importedService(t, journal)
	cmd := CalculateBuild{SessionID: "main", ExpectedRevision: 2, RequestID: "calculate"}
	calculated, err := s.Calculate(t.Context(), cmd)
	if err != nil {
		t.Fatal(err)
	}
	view, err := ViewSnapshot(calculated)
	if err != nil || view.Build.Calculation.Stats["TotalDPS"] <= 0 || calculated.CalculatedFromRevision != imported.BuildRevision {
		t.Fatalf("расчёт: %+v %v", view, err)
	}
	changed, err := s.Apply(t.Context(), ApplyBuild{SessionID: "main", ExpectedRevision: 3, RequestID: "config", Edit: &build.Edit{Config: &build.ConfigEdit{Name: "customMods", Value: &build.ConfigValue{Type: "string", Value: "100% increased Spell Damage"}}}})
	if err != nil || changed.CalculationJSON != "" || changed.BuildRevision != 4 {
		t.Fatalf("устаревший расчёт не сброшен: %+v %v", changed, err)
	}
	restored, err := OpenService(t.Context(), journal)
	if err != nil {
		t.Fatal(err)
	}
	if got, err := restored.Snapshot(t.Context(), "main"); err != nil || got != changed {
		t.Fatal("потеряна сборка после восстановления")
	}
	if got, err := restored.Calculate(t.Context(), cmd); err != nil || got != calculated {
		t.Fatal("повтор не вернул исходный результат")
	}
	rename := RenameBuild{SessionID: "main", ExpectedRevision: 4, RequestID: "rename", BuildName: "Новое имя"}
	renamed, err := restored.Rename(t.Context(), rename)
	if err != nil || renamed.BuildXML != changed.BuildXML || renamed.InputSHA256 != changed.InputSHA256 {
		t.Fatal("переименование потеряло сборку")
	}
	_, err = restored.Rename(t.Context(), RenameBuild{SessionID: "main", BuildName: "Конфликт", ExpectedRevision: 5, RequestID: "calculate"})
	expectCode(t, err, RequestConflict)
}

func TestRecoveryUsesSavedNumbersWithoutRecalculation(t *testing.T) {
	journal := &testJournal{}
	s, _ := importedService(t, journal)
	cmd := CalculateBuild{SessionID: "main", ExpectedRevision: 2, RequestID: "old-engine-result"}
	if _, err := s.Calculate(t.Context(), cmd); err != nil {
		t.Fatal(err)
	}
	// Имитируем иной результат ранее выпущенной версии, сохраняя привязки входа и данных.
	change := &journal.changes[len(journal.changes)-1]
	var saved calculation.Result
	if err := json.Unmarshal([]byte(change.Result.CalculationJSON), &saved); err != nil {
		t.Fatal(err)
	}
	saved.Stats["TotalDPS"] = 123.456
	data, err := json.Marshal(saved)
	if err != nil {
		t.Fatal(err)
	}
	change.Result.CalculationJSON = string(data)
	restored, err := OpenService(t.Context(), journal)
	if err != nil {
		t.Fatal(err)
	}
	got, err := restored.Calculate(t.Context(), cmd)
	if err != nil || got.CalculationJSON != string(data) {
		t.Fatal("сохранённые числа пересчитаны вместо восстановления")
	}
}

func TestBuildStorageFailureDoesNotPublish(t *testing.T) {
	journal := &testJournal{}
	s, initial := importedService(t, journal)
	journal.writeError = errors.New("диск недоступен")
	_, err := s.Calculate(t.Context(), CalculateBuild{SessionID: "main", ExpectedRevision: 2, RequestID: "calculate"})
	expectCode(t, err, StorageFailure)
	if got, _ := s.Snapshot(t.Context(), "main"); got != initial {
		t.Fatal("расчёт опубликован до устойчивой записи")
	}
	journal.writeError = nil
	if _, err := s.Calculate(t.Context(), CalculateBuild{SessionID: "main", ExpectedRevision: 2, RequestID: "calculate"}); err != nil {
		t.Fatal(err)
	}
	if len(journal.changes) != 3 {
		t.Fatal("ошибка или повтор записаны как новая команда")
	}
}

func TestQueuedMultigroupCalculationCannotOverwriteGroupSelection(t *testing.T) {
	journal := &testJournal{}
	s, initial := importedService(t, journal)
	xml, err := os.ReadFile("../../testdata/reference/0.0.5/caster-main.xml")
	if err != nil {
		t.Fatal(err)
	}
	imported, err := s.Apply(t.Context(), ApplyBuild{SessionID: "main", ExpectedRevision: initial.Revision, RequestID: "caster", XML: string(xml)})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
	defer cancel()
	release := holdWorkers(s)
	completed := make(chan error, 1)
	go func() {
		_, err := s.Calculate(ctx, CalculateBuild{SessionID: "main", ExpectedRevision: imported.Revision, RequestID: "queued"})
		completed <- err
	}()
	waitAdmission(t, s, 1)
	changed, err := s.Apply(t.Context(), ApplyBuild{SessionID: "main", ExpectedRevision: imported.Revision, RequestID: "select-group",
		Edit: &build.Edit{Selection: &build.SelectionEdit{MainGroup: 1}}})
	release()
	if err != nil {
		t.Fatal(err)
	}
	expectCode(t, <-completed, RevisionConflict)
	if got, err := s.Snapshot(t.Context(), "main"); err != nil || got != changed || len(journal.changes) != int(changed.Revision) {
		t.Fatal("устаревший расчёт изменил сессию или журнал")
	}
}

func waitAdmission(t *testing.T, s *Service, count int) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if len(s.admission) == count {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("в очереди %d, ожидалось %d", len(s.admission), count)
}

func holdWorkers(s *Service) func() {
	for range cap(s.workers) {
		s.workers <- struct{}{}
	}
	return func() {
		for range cap(s.workers) {
			<-s.workers
		}
	}
}

func TestConcurrentCalculationRetrySharesWork(t *testing.T) {
	s, _ := importedService(t, nil)
	release := holdWorkers(s)
	cmd := CalculateBuild{SessionID: "main", ExpectedRevision: 2, RequestID: "shared"}
	type answer struct {
		snapshot domain.Snapshot
		err      error
	}
	answers := make(chan answer, 2)
	go func() { snapshot, err := s.Calculate(t.Context(), cmd); answers <- answer{snapshot, err} }()
	waitAdmission(t, s, 1)
	ctx, updates := WithProgress(t.Context())
	go func() { snapshot, err := s.Calculate(ctx, cmd); answers <- answer{snapshot, err} }()
	select {
	case update := <-updates:
		if update.Stage != "waiting" {
			t.Fatal(update)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("повтор не дождался текущей задачи")
	}
	if len(s.admission) != 1 {
		t.Fatal("повтор добавил вторую работу")
	}
	release()
	first, second := <-answers, <-answers
	if first.err != nil || second.err != nil || first.snapshot != second.snapshot || first.snapshot.Revision != 3 {
		t.Fatal("повторы завершились разными результатами")
	}
}

func TestQueuedCalculationCannotOverwriteNewInput(t *testing.T) {
	s, _ := importedService(t, nil)
	release := holdWorkers(s)
	done := make(chan error, 1)
	go func() {
		_, err := s.Calculate(t.Context(), CalculateBuild{SessionID: "main", ExpectedRevision: 2, RequestID: "queued"})
		done <- err
	}()
	waitAdmission(t, s, 1)
	changed, err := s.Apply(t.Context(), ApplyBuild{SessionID: "main", ExpectedRevision: 2, RequestID: "level", Edit: &build.Edit{Level: &build.LevelEdit{Level: 20}}})
	if err != nil {
		t.Fatal(err)
	}
	release()
	expectCode(t, <-done, RevisionConflict)
	current, _ := s.Snapshot(t.Context(), "main")
	if current != changed || current.CalculationJSON != "" {
		t.Fatal("старый расчёт подменил новый вход")
	}
}

func TestQueueCancellationAndCapacity(t *testing.T) {
	s, _ := importedService(t, nil)
	release := holdWorkers(s)
	defer release()
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	answers := make(chan error, cap(s.admission))
	for i := 0; i < cap(s.admission); i++ {
		go func() {
			_, err := s.Calculate(ctx, CalculateBuild{SessionID: "main", ExpectedRevision: 2, RequestID: fmt.Sprintf("queued-%d", i)})
			answers <- err
		}()
	}
	waitAdmission(t, s, cap(s.admission))
	_, err := s.Calculate(t.Context(), CalculateBuild{SessionID: "main", ExpectedRevision: 2, RequestID: "overflow"})
	expectCode(t, err, ResourceExhausted)
	cancel()
	for range cap(s.admission) {
		if err := <-answers; !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	}
	waitAdmission(t, s, 0)
}
