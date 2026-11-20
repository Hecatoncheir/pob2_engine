package application

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
)

func expectCode(t *testing.T, err error, code Code) {
	t.Helper()
	var app *Error
	if !errors.As(err, &app) || app.Code != code {
		t.Fatalf("ошибка = %v, нужен код %s", err, code)
	}
}

func TestCommandReplayKeepsOriginalResult(t *testing.T) {
	ctx := t.Context()
	s := NewService()
	create := CreateSession{SessionID: "main", BuildName: "Исходная", RequestID: "create-1"}
	initial, err := s.Create(ctx, create)
	if err != nil {
		t.Fatal(err)
	}
	first := RenameBuild{SessionID: "main", BuildName: "Первая", ExpectedRevision: 1, RequestID: "rename-1"}
	changed, err := s.Rename(ctx, first)
	if err != nil {
		t.Fatal(err)
	}
	latest, err := s.Rename(ctx, RenameBuild{SessionID: "main", BuildName: "Вторая", ExpectedRevision: 2, RequestID: "rename-2"})
	if err != nil {
		t.Fatal(err)
	}
	replayed, err := s.Rename(ctx, first)
	if err != nil || replayed != changed {
		t.Fatalf("повтор правки: %v, %v", replayed, err)
	}
	recreated, err := s.Create(ctx, create)
	if err != nil || recreated != initial {
		t.Fatalf("повтор создания: %v, %v", recreated, err)
	}
	current, err := s.Snapshot(ctx, "main")
	if err != nil || current != latest || current.Revision != 3 {
		t.Fatalf("повтор изменил состояние: %v, %v", current, err)
	}
}

func TestCommandConflictsDoNotMutateSession(t *testing.T) {
	s := NewService()
	ctx := t.Context()
	create := CreateSession{SessionID: "main", BuildName: "Исходная", RequestID: "create-1"}
	if _, err := s.Create(ctx, create); err != nil {
		t.Fatal(err)
	}
	create.BuildName = "Другая"
	_, err := s.Create(ctx, create)
	expectCode(t, err, RequestConflict)
	create.RequestID = "create-2"
	_, err = s.Create(ctx, create)
	expectCode(t, err, AlreadyExists)
	cmd := RenameBuild{SessionID: "main", BuildName: "Новая", ExpectedRevision: 1, RequestID: "rename-1"}
	if _, err = s.Rename(ctx, cmd); err != nil {
		t.Fatal(err)
	}
	cmd.BuildName = "Подмена"
	_, err = s.Rename(ctx, cmd)
	expectCode(t, err, RequestConflict)
	cmd.RequestID = "stale"
	_, err = s.Rename(ctx, cmd)
	expectCode(t, err, RevisionConflict)
	cmd.ExpectedRevision, cmd.RequestID = 2, "create-1"
	_, err = s.Rename(ctx, cmd)
	expectCode(t, err, RequestConflict)
	current, err := s.Snapshot(ctx, "main")
	if err != nil || current.BuildName != "Новая" || current.Revision != 2 {
		t.Fatalf("конфликт изменил состояние: %v, %v", current, err)
	}
}

func TestConcurrentRenamesAcceptOneRevision(t *testing.T) {
	s := NewService()
	ctx := t.Context()
	if _, err := s.Create(ctx, CreateSession{SessionID: "main", BuildName: "Начало", RequestID: "create"}); err != nil {
		t.Fatal(err)
	}
	const count = 20
	results := make(chan error, count)
	var group sync.WaitGroup
	for i := range count {
		group.Go(func() {
			_, err := s.Rename(ctx, RenameBuild{SessionID: "main", BuildName: fmt.Sprint(i), ExpectedRevision: 1, RequestID: fmt.Sprintf("rename-%d", i)})
			results <- err
		})
	}
	group.Wait()
	close(results)
	success := 0
	for err := range results {
		if err == nil {
			success++
		} else {
			expectCode(t, err, RevisionConflict)
		}
	}
	current, err := s.Snapshot(ctx, "main")
	if err != nil || success != 1 || current.Revision != 2 {
		t.Fatalf("успехов %d, снимок %v, ошибка %v", success, current, err)
	}
}

func TestInvalidOrCanceledCommandDoesNotCreateSession(t *testing.T) {
	s := NewService()
	for _, cmd := range []CreateSession{
		{SessionID: "../main", BuildName: "Имя", RequestID: "create"},
		{SessionID: "main", BuildName: "  ", RequestID: "create"},
		{SessionID: "main", BuildName: "Имя", RequestID: ""},
	} {
		_, err := s.Create(t.Context(), cmd)
		expectCode(t, err, InvalidArgument)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err := s.Create(ctx, CreateSession{SessionID: "main", BuildName: "Имя", RequestID: "create"})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("ошибка = %v", err)
	}
	_, err = s.Snapshot(t.Context(), "main")
	expectCode(t, err, NotFound)
}

func TestSessionsAreIsolated(t *testing.T) {
	s := NewService()
	for _, id := range []string{"a", "b"} {
		if _, err := s.Create(t.Context(), CreateSession{SessionID: id, BuildName: id, RequestID: "create-" + id}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.Rename(t.Context(), RenameBuild{SessionID: "a", BuildName: "Изменена", ExpectedRevision: 1, RequestID: "rename"}); err != nil {
		t.Fatal(err)
	}
	other, err := s.Snapshot(t.Context(), "b")
	if err != nil || other.BuildName != "b" || other.Revision != 1 {
		t.Fatalf("изменена другая сессия: %v, %v", other, err)
	}
}
