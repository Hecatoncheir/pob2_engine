package lifecycle

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func manager(t *testing.T, mode string) *Manager {
	t.Helper()
	m, err := New(Config{Mode: mode, IdleTimeout: 60 * time.Millisecond, LeaseTimeout: 300 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(m.Close)
	return m
}

func alive(t *testing.T, m *Manager, duration time.Duration) {
	t.Helper()
	select {
	case <-m.Draining():
		t.Fatal("сервер завершился преждевременно")
	case <-time.After(duration):
	}
}

func drained(t *testing.T, m *Manager) {
	t.Helper()
	select {
	case <-m.Draining():
	case <-time.After(2 * time.Second):
		t.Fatal("сервер не перешёл в draining")
	}
}

func TestAutomaticWaitsForLastClient(t *testing.T) {
	m := manager(t, Automatic)
	a, err := m.Register("a", "ui")
	if err != nil {
		t.Fatal(err)
	}
	b, err := m.Register("b", "mcp")
	if err != nil {
		t.Fatal(err)
	}
	m.Release(a.ID)
	alive(t, m, 100*time.Millisecond)
	if got := m.Stats(); got.Clients != 1 {
		t.Fatalf("клиентов: %+v", got)
	}
	m.Release(b.ID)
	drained(t, m)
	if _, err := m.Register("late", "cli"); !errors.Is(err, ErrDraining) {
		t.Fatalf("поздняя регистрация: %v", err)
	}
}

func TestTaskOutlivesClientAndFinishIsIdempotent(t *testing.T) {
	m := manager(t, Automatic)
	lease, err := m.Register("a", "cli")
	if err != nil {
		t.Fatal(err)
	}
	finish, err := m.BeginTask(lease.ID)
	if err != nil {
		t.Fatal(err)
	}
	m.Release(lease.ID)
	alive(t, m, 100*time.Millisecond)
	if got := m.Stats(); got.Clients != 0 || got.Tasks != 1 {
		t.Fatalf("учёт задач: %+v", got)
	}
	finish()
	finish()
	drained(t, m)
	if _, err := m.BeginTask(""); !errors.Is(err, ErrDraining) {
		t.Fatalf("задача после draining: %v", err)
	}
}

func TestHeartbeatAndExpiry(t *testing.T) {
	m := manager(t, Automatic)
	lease, err := m.Register("client", "ui")
	if err != nil {
		t.Fatal(err)
	}
	for range 5 {
		alive(t, m, 80*time.Millisecond)
		if err := m.Heartbeat(lease.ID); err != nil {
			t.Fatal(err)
		}
	}
	drained(t, m)
	if got := m.Stats(); got.Clients != 0 {
		t.Fatalf("истёкшая аренда сохранена: %+v", got)
	}
}

func TestPersistentDoesNotStopWithoutClients(t *testing.T) {
	m := manager(t, Persistent)
	lease, err := m.Register("client", "mcp")
	if err != nil {
		t.Fatal(err)
	}
	m.Release(lease.ID)
	alive(t, m, 100*time.Millisecond)
	m.Drain()
	drained(t, m)
}

func TestRegistrationReplayAndReconnect(t *testing.T) {
	m := manager(t, Automatic)
	first, err := m.Register("client", "ui")
	if err != nil {
		t.Fatal(err)
	}
	second, err := m.Register("client", "ui")
	if err != nil || first != second || m.Stats().Clients != 1 {
		t.Fatalf("повтор регистрации: %v, %v", second, err)
	}
	if _, err := m.Register("client", "cli"); !errors.Is(err, ErrClientConflict) {
		t.Fatalf("смена типа: %v", err)
	}
	m.Release(first.ID)
	m.Release(first.ID)
	if _, err := m.BeginTask(first.ID); !errors.Is(err, ErrLeaseExpired) {
		t.Fatalf("задача со старой арендой: %v", err)
	}
	third, err := m.Register("client", "ui")
	if err != nil || third.ID == first.ID {
		t.Fatalf("переподключение: %v, %v", third, err)
	}
	alive(t, m, 100*time.Millisecond)
}

func TestRegistrationAndIdleDecisionAreAtomic(t *testing.T) {
	for range 50 {
		m := manager(t, Automatic)
		var registered error
		var group sync.WaitGroup
		group.Go(func() { _, registered = m.Register("client", "ui") })
		group.Go(func() { m.mu.Lock(); m.maintain(time.Now().Add(100 * time.Millisecond)); m.mu.Unlock() })
		group.Wait()
		stats := m.Stats()
		if registered == nil && (stats.State != "ready" || stats.Clients != 1) {
			t.Fatalf("принятая регистрация потеряна: %+v", stats)
		}
		if registered != nil && !errors.Is(registered, ErrDraining) {
			t.Fatal(registered)
		}
		m.Close()
	}
}
