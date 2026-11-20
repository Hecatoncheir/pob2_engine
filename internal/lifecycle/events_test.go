package lifecycle

import (
	"context"
	"errors"
	"io"
	"sync"
	"testing"
	"time"
)

func TestStartingRejectsWorkAndPublishesEveryTransition(t *testing.T) {
	m, err := NewStarting(Config{Mode: Persistent})
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	sub := m.Watch()
	if _, err := m.Register("early", "cli"); !errors.Is(err, ErrDraining) {
		t.Fatalf("регистрация при запуске: %v", err)
	}
	if _, err := m.BeginTask(""); !errors.Is(err, ErrDraining) {
		t.Fatalf("задача при запуске: %v", err)
	}
	if err := m.MarkReady(); err != nil {
		t.Fatal(err)
	}
	if err := m.MarkReady(); err != nil {
		t.Fatal(err)
	}
	m.Drain()
	m.Drain()
	m.Close()
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	for index, state := range []string{"starting", "ready", "draining", "stopped"} {
		event, err := sub.Next(ctx)
		if err != nil || event.State != state || event.Sequence != uint64(index+1) {
			t.Fatalf("переход %s: %+v %v", state, event, err)
		}
	}
	if err := m.MarkReady(); !errors.Is(err, ErrDraining) {
		t.Fatalf("повторный запуск остановленного: %v", err)
	}
	if len(m.events) != 4 {
		t.Fatal("повторное действие создало лишний переход")
	}
	if _, err := sub.Next(ctx); !errors.Is(err, io.EOF) {
		t.Fatalf("нет EOF после остановки: %v", err)
	}
}

func TestStartingMayStopWithoutReady(t *testing.T) {
	m, err := NewStarting(Config{Mode: Automatic, IdleTimeout: 50 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	m.mu.Lock()
	m.maintain(time.Now().Add(time.Hour))
	m.mu.Unlock()
	if m.Stats().State != "starting" {
		t.Fatal("таймер остановил незавершённый запуск")
	}
	m.Close()
	if m.Stats().State != "stopped" {
		t.Fatal("запускающийся сервер не закрыт")
	}
}

func TestWatchAndDrainHaveNoGap(t *testing.T) {
	for range 50 {
		m := manager(t, Persistent)
		var sub *Subscription
		var group sync.WaitGroup
		group.Go(func() { sub = m.Watch() })
		group.Go(m.Drain)
		group.Wait()
		ctx, cancel := context.WithTimeout(t.Context(), time.Second)
		event, err := sub.Next(ctx)
		if err == nil && event.State == "ready" {
			event, err = sub.Next(ctx)
		}
		cancel()
		if err != nil || event.State != "draining" || event.Reason != "shutdown" {
			t.Fatalf("пропущена остановка: %+v %v", event, err)
		}
		m.Close()
	}
}

func TestAutomaticDrainHasItsOwnReason(t *testing.T) {
	m := manager(t, Automatic)
	sub := m.Watch()
	m.mu.Lock()
	m.maintain(time.Now().Add(time.Second))
	m.mu.Unlock()
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	if _, err := sub.Next(ctx); err != nil {
		t.Fatal(err)
	}
	event, err := sub.Next(ctx)
	if err != nil || event.State != "draining" || event.Reason != "idle_timeout" {
		t.Fatalf("автоматическая остановка: %+v %v", event, err)
	}
}
