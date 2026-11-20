package application

import (
	"context"
	"errors"
	"testing"

	"github.com/Hecatoncheir/pob2_engine/internal/domain"
)

type testJournal struct {
	changes    []Change
	writeError error
}

func (j *testJournal) Path() string { return "test.db" }
func (j *testJournal) Load(ctx context.Context, apply func(Change) error) error {
	for _, change := range j.changes {
		if err := apply(change); err != nil {
			return err
		}
	}
	return ctx.Err()
}
func (j *testJournal) Append(ctx context.Context, change Change) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if j.writeError != nil {
		return j.writeError
	}
	j.changes = append(j.changes, change)
	return nil
}

func TestStorageFailureDoesNotPublishOrReserveCommand(t *testing.T) {
	j := &testJournal{writeError: errors.New("ошибка записи")}
	s, err := OpenService(t.Context(), j)
	if err != nil {
		t.Fatal(err)
	}
	create := CreateSession{SessionID: "main", BuildName: "Начальная", RequestID: "create"}
	_, err = s.Create(t.Context(), create)
	expectCode(t, err, StorageFailure)
	_, err = s.Snapshot(t.Context(), "main")
	expectCode(t, err, NotFound)
	j.writeError = nil
	initial, err := s.Create(t.Context(), create)
	if err != nil {
		t.Fatal(err)
	}
	rename := RenameBuild{SessionID: "main", BuildName: "Новая", ExpectedRevision: 1, RequestID: "rename"}
	j.writeError = errors.New("ошибка записи")
	_, err = s.Rename(t.Context(), rename)
	expectCode(t, err, StorageFailure)
	if got, err := s.Snapshot(t.Context(), "main"); err != nil || got != initial {
		t.Fatalf("неподтверждённая правка опубликована: %v, %v", got, err)
	}
	j.writeError = nil
	changed, err := s.Rename(t.Context(), rename)
	if err != nil || changed.Revision != 2 {
		t.Fatalf("повтор после ошибки: %v, %v", changed, err)
	}
	if len(j.changes) != 2 {
		t.Fatalf("сохранены неуспешные команды: %d", len(j.changes))
	}
	restored, err := OpenService(t.Context(), j)
	if err != nil {
		t.Fatal(err)
	}
	if got, err := restored.Rename(t.Context(), rename); err != nil || got != changed {
		t.Fatalf("повтор после восстановления: %v, %v", got, err)
	}
	if len(j.changes) != 2 {
		t.Fatal("повтор команды повторно записан в журнал")
	}
}

func TestRecoveryRejectsInvalidJournal(t *testing.T) {
	create := CreateSession{SessionID: "main", BuildName: "Начальная", RequestID: "create"}
	initial := Change{Create: &create, Result: domain.Snapshot{SessionID: "main", BuildName: "Начальная", Revision: 1}}
	wrongResult := initial
	wrongResult.Result.Revision = 10
	rename := RenameBuild{SessionID: "missing", BuildName: "Правка", ExpectedRevision: 1, RequestID: "rename"}
	for _, changes := range [][]Change{{{}}, {wrongResult}, {initial, initial}, {{Rename: &rename}}} {
		service, err := OpenService(t.Context(), &testJournal{changes: changes})
		if err == nil || service != nil {
			t.Fatal("повреждённый журнал принят")
		}
	}
}
