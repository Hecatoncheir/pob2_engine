package host

import (
	"context"
	"errors"
	"io"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/discovery"
	"github.com/Hecatoncheir/pob2_engine/internal/lifecycle"
)

type blockedStore struct {
	entered chan struct{}
	finish  chan struct{}
	err     error
}

func (s *blockedStore) Close() error {
	close(s.entered)
	<-s.finish
	return s.err
}

func TestCloseDeadlineKeepsProfileLockedUntilStorageIsClosed(t *testing.T) {
	profile, err := discovery.Open(t.TempDir(), "closing")
	if err != nil {
		t.Fatal(err)
	}
	closeError := errors.New("ошибка закрытия хранилища")
	store := &blockedStore{entered: make(chan struct{}), finish: make(chan struct{}), err: closeError}
	var release sync.Once
	unblock := func() { release.Do(func() { close(store.finish) }) }
	defer unblock()
	s, err := start(profile, "127.0.0.1:0", "token", lifecycle.Config{}, func() (application.Commands, io.Closer, error) { return application.NewService(), store, nil })
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		unblock()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = s.Close(ctx)
	})
	ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
	defer cancel()
	finished := make(chan error, 1)
	go func() { finished <- s.Close(ctx) }()
	select {
	case <-store.entered:
	case <-time.After(time.Second):
		t.Fatal("хранилище не начало закрываться")
	}
	select {
	case err := <-finished:
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("ожидался таймаут: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Close завис на хранилище")
	}
	if lock, err := profile.TryLock("server"); !errors.Is(err, discovery.ErrLocked) {
		if lock != nil {
			lock.Close()
		}
		t.Fatalf("блокировка освобождена раньше базы: %v", err)
	}
	if _, err := profile.Read(); err != nil {
		t.Fatalf("запись владельца удалена до закрытия базы: %v", err)
	}
	select {
	case <-s.Done():
		t.Fatal("зависшая очистка объявлена завершённой")
	default:
	}
	unblock()
	select {
	case <-s.Done():
	case <-time.After(time.Second):
		t.Fatal("очистка не завершилась")
	}
	if !errors.Is(s.Err(), closeError) {
		t.Fatalf("ошибка закрытия потеряна: %v", s.Err())
	}
	if _, err := profile.Read(); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("запись владельца не удалена: %v", err)
	}
	lock, err := profile.TryLock("server")
	if err != nil {
		t.Fatal(err)
	}
	lock.Close()
}
