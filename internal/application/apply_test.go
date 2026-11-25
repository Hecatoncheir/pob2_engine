package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Hecatoncheir/pob2_engine/internal/build"
	"github.com/Hecatoncheir/pob2_engine/internal/calculation"
	"github.com/Hecatoncheir/pob2_engine/internal/domain"
)

type blockedTreeEditor struct {
	calculation.Calculator
	started chan struct{}
	replied chan struct{}
	replies chan error
}

func blockTreeEdits(service *Service) *blockedTreeEditor {
	editor := &blockedTreeEditor{
		Calculator: service.calculator,
		started:    make(chan struct{}, 2),
		replied:    make(chan struct{}, 2),
		replies:    make(chan error, 2),
	}
	service.calculator = editor
	return editor
}

func (e *blockedTreeEditor) ResolveTreeEdit(ctx context.Context, _ *build.Document, edit build.TreeNodeEdit) (build.PassiveChange, error) {
	e.started <- struct{}{}
	select {
	case err := <-e.replies:
		e.replied <- struct{}{}
		return build.PassiveChange{Added: []int{edit.NodeID}}, err
	case <-ctx.Done():
		return build.PassiveChange{}, ctx.Err()
	}
}

func allocateNode(snapshot domain.Snapshot) ApplyBuild {
	return ApplyBuild{SessionID: snapshot.SessionID, ExpectedRevision: snapshot.Revision, RequestID: "tree-edit",
		Edit: &build.Edit{TreeNode: &build.TreeNodeEdit{NodeID: 123, Action: "allocate"}}}
}

type applyAnswer struct {
	snapshot domain.Snapshot
	err      error
}

func startApply(ctx context.Context, service *Service, cmd ApplyBuild, answers chan<- applyAnswer) {
	go func() {
		snapshot, err := service.Apply(ctx, cmd)
		answers <- applyAnswer{snapshot: snapshot, err: err}
	}()
}

func awaitApply[T any](t *testing.T, result <-chan T) T {
	t.Helper()
	select {
	case answer := <-result:
		return answer
	case <-time.After(3 * time.Second):
		t.Fatal("операция не завершилась за отведённое время")
		var zero T
		return zero
	}
}

func TestApplyAllowsReadingWhileTreeEditWaitsAndCanBeCanceled(t *testing.T) {
	journal := &testJournal{}
	service, initial := importedService(t, journal)
	editor := blockTreeEdits(service)
	command := allocateNode(initial)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	done := make(chan applyAnswer, 1)
	startApply(ctx, service, command, done)
	awaitApply(t, editor.started)
	read := make(chan applyAnswer, 1)
	go func() {
		snapshot, err := service.Snapshot(t.Context(), initial.SessionID)
		read <- applyAnswer{snapshot: snapshot, err: err}
	}()
	if answer := awaitApply(t, read); answer.err != nil || answer.snapshot != initial {
		t.Fatalf("пока готовится правка, ожидался прежний снимок: ревизия %d, %v", answer.snapshot.Revision, answer.err)
	}
	value, err := service.find(initial.SessionID)
	if err != nil {
		t.Fatal(err)
	}
	func() {
		value.mu.Lock()
		defer value.mu.Unlock()
		editor.replies <- nil
		awaitApply(t, editor.replied)
		cancel()
	}()
	if answer := awaitApply(t, done); !errors.Is(answer.err, context.Canceled) || len(journal.changes) != int(initial.Revision) {
		t.Fatalf("отмена правки: %v, записей в журнале: %d", answer.err, len(journal.changes))
	}
	editor.replies <- nil
	if snapshot, err := service.Apply(t.Context(), command); err != nil || snapshot.Revision != initial.Revision+1 {
		t.Fatalf("повтор после отмены: ревизия %d, %v", snapshot.Revision, err)
	}
}

func TestApplyRechecksRevisionAfterTreeEdit(t *testing.T) {
	journal := &testJournal{}
	service, initial := importedService(t, journal)
	editor := blockTreeEdits(service)
	done := make(chan applyAnswer, 1)
	startApply(t.Context(), service, allocateNode(initial), done)
	awaitApply(t, editor.started)
	rename := make(chan applyAnswer, 1)
	go func() {
		snapshot, err := service.Rename(t.Context(), RenameBuild{SessionID: initial.SessionID, ExpectedRevision: initial.Revision,
			RequestID: "rename", BuildName: "Новое имя"})
		rename <- applyAnswer{snapshot: snapshot, err: err}
	}()
	changed := awaitApply(t, rename)
	if changed.err != nil {
		t.Fatal(changed.err)
	}
	editor.replies <- nil
	expectCode(t, awaitApply(t, done).err, RevisionConflict)
	if snapshot, err := service.Snapshot(t.Context(), initial.SessionID); err != nil || snapshot != changed.snapshot || len(journal.changes) != int(changed.snapshot.Revision) {
		t.Fatal("подготовленная ранее правка изменила сессию или журнал", err)
	}
}

func TestApplyPreservesTypedTreeEditErrors(t *testing.T) {
	service, initial := importedService(t, nil)
	editor := blockTreeEdits(service)
	command := allocateNode(initial)
	for _, want := range []error{context.Canceled, context.DeadlineExceeded, failure(ResourceExhausted, "очередь заполнена")} {
		editor.replies <- want
		if _, err := service.Apply(t.Context(), command); !errors.Is(err, want) {
			t.Fatalf("ошибка правки: %v, ожидалась %v", err, want)
		}
		awaitApply(t, editor.started)
		awaitApply(t, editor.replied)
	}
}

func TestConcurrentApplyReplayPublishesOnce(t *testing.T) {
	journal := &testJournal{}
	service, initial := importedService(t, journal)
	editor := blockTreeEdits(service)
	command := allocateNode(initial)
	answers := make(chan applyAnswer, 2)
	startApply(t.Context(), service, command, answers)
	startApply(t.Context(), service, command, answers)
	awaitApply(t, editor.started)
	awaitApply(t, editor.started)
	editor.replies <- nil
	a := awaitApply(t, answers)
	editor.replies <- errors.New("процесс Lua завершился")
	b := awaitApply(t, answers)
	if a.err != nil || b.err != nil || a.snapshot != b.snapshot || a.snapshot.Revision != initial.Revision+1 {
		t.Fatalf("повторы должны вернуть один снимок: ревизии %d и %d, ошибки %v и %v", a.snapshot.Revision, b.snapshot.Revision, a.err, b.err)
	}
	if len(journal.changes) != int(a.snapshot.Revision) {
		t.Fatalf("повтор правки записан в журнал: %d", len(journal.changes))
	}
	value, err := service.find(initial.SessionID)
	if err != nil || len(value.events) != int(a.snapshot.Revision) {
		t.Fatal("повтор создал лишнее событие", err)
	}
}
