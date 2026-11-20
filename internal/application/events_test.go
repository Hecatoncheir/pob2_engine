package application

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestWatchHistorySurvivesRecoveryAndIgnoresRejectedCommands(t *testing.T) {
	j := &testJournal{}
	s, err := OpenService(t.Context(), j)
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.Create(t.Context(), CreateSession{SessionID: "main", BuildName: "До", RequestID: "create"})
	if err != nil {
		t.Fatal(err)
	}
	sub, err := s.Watch(t.Context(), "main", 0)
	if err != nil {
		t.Fatal(err)
	}
	rename := RenameBuild{SessionID: "main", BuildName: "После", ExpectedRevision: 1, RequestID: "rename"}
	j.writeError = errors.New("диск недоступен")
	_, err = s.Rename(t.Context(), rename)
	expectCode(t, err, StorageFailure)
	j.writeError = nil
	for range 2 {
		if _, err := s.Rename(t.Context(), rename); err != nil {
			t.Fatal(err)
		}
	}
	initial, err := sub.Next(t.Context())
	if err != nil || initial.Kind != "snapshot" || initial.Snapshot.Revision != 1 {
		t.Fatalf("снимок: %+v %v", initial, err)
	}
	changed, err := sub.Next(t.Context())
	if err != nil || changed.RequestID != "rename" || changed.PreviousBuildName != "До" || changed.Snapshot.Revision != 2 {
		t.Fatalf("правка: %+v %v", changed, err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Millisecond)
	defer cancel()
	if _, err := sub.Next(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("лишнее событие: %v", err)
	}
	restored, err := OpenService(t.Context(), j)
	if err != nil {
		t.Fatal(err)
	}
	replay, err := restored.Watch(t.Context(), "main", 1)
	if err != nil {
		t.Fatal(err)
	}
	if got, err := replay.Next(t.Context()); err != nil || got != changed {
		t.Fatalf("восстановление: %+v %v", got, err)
	}
	_, err = restored.Watch(t.Context(), "main", 3)
	expectCode(t, err, RevisionConflict)
}

func TestWatchConcurrentSubscriptionHasNoGapsAndDoesNotBlockWriter(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
	defer cancel()
	s := NewService()
	if _, err := s.Create(ctx, CreateSession{SessionID: "main", BuildName: "0", RequestID: "create"}); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() {
		for i := 1; i <= 200; i++ {
			_, err := s.Rename(ctx, RenameBuild{SessionID: "main", BuildName: fmt.Sprint(i), ExpectedRevision: uint64(i), RequestID: fmt.Sprint(i)})
			if err != nil {
				done <- err
				return
			}
		}
		done <- nil
	}()
	sub, err := s.Watch(ctx, "main", 0)
	if err != nil {
		t.Fatal(err)
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	event, err := sub.Next(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for revision := event.Snapshot.Revision + 1; revision <= 201; revision++ {
		event, err = sub.Next(ctx)
		if err != nil || event.Snapshot.Revision != revision {
			t.Fatalf("пропуск ревизии %d: %+v %v", revision, event, err)
		}
	}
}
