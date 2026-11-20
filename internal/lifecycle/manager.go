package lifecycle

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"sync"
	"time"
)

const (
	Automatic           = "automatic"
	Persistent          = "persistent"
	DefaultIdleTimeout  = 5 * time.Second
	DefaultLeaseTimeout = 15 * time.Second
)

var (
	ErrDraining       = errors.New("сервер завершает работу")
	ErrLeaseExpired   = errors.New("аренда клиента отсутствует или истекла")
	ErrClientConflict = errors.New("client_id уже зарегистрирован с другим типом")
	ErrClientLimit    = errors.New("достигнут лимит клиентов")
)

type Config struct {
	Mode         string
	IdleTimeout  time.Duration
	LeaseTimeout time.Duration
}

func (c Config) Normalize() (Config, error) {
	if c.Mode == "" {
		c.Mode = Persistent
	}
	if c.IdleTimeout == 0 {
		c.IdleTimeout = DefaultIdleTimeout
	}
	if c.LeaseTimeout == 0 {
		c.LeaseTimeout = DefaultLeaseTimeout
	}
	if c.Mode != Automatic && c.Mode != Persistent {
		return c, fmt.Errorf("режим должен быть automatic или persistent")
	}
	if c.IdleTimeout < 50*time.Millisecond || c.LeaseTimeout < 100*time.Millisecond || c.LeaseTimeout > 24*time.Hour {
		return c, fmt.Errorf("idle-timeout должен быть не меньше 50ms; lease-timeout — от 100ms до 24h")
	}
	return c, nil
}

type Lease struct {
	ID      string
	Timeout time.Duration
}

type client struct {
	id      string
	kind    string
	expires time.Time
}

type Stats struct {
	State   string
	Clients uint32
	Tasks   uint32
}

// Manager принимает регистрацию, задачи и решение об остановке под одним mutex.
type Manager struct {
	mu        sync.Mutex
	config    Config
	state     string
	clients   map[string]client
	tasks     uint32
	idleSince time.Time
	draining  chan struct{}
	tasksDone chan struct{}
	tasksOnce sync.Once
	context   context.Context
	cancel    context.CancelFunc
	closed    chan struct{}
	done      chan struct{}
	closeOnce sync.Once
	events    []Event
	changed   chan struct{}
}

func New(config Config) (*Manager, error) {
	m, err := NewStarting(config)
	if err != nil {
		return nil, err
	}
	_ = m.MarkReady()
	return m, nil
}

// NewStarting запрещает клиентов и задачи до явного подтверждения готовности.
func NewStarting(config Config) (*Manager, error) {
	config, err := config.Normalize()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	m := &Manager{
		config: config, state: "starting", clients: make(map[string]client),
		draining: make(chan struct{}), tasksDone: make(chan struct{}),
		context: ctx, cancel: cancel, closed: make(chan struct{}), done: make(chan struct{}),
		events: []Event{{Sequence: 1, State: "starting"}}, changed: make(chan struct{}),
	}
	go m.monitor()
	return m, nil
}

func (m *Manager) MarkReady() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.state == "ready" {
		return nil
	}
	if m.state != "starting" {
		return ErrDraining
	}
	m.transition("ready", "")
	m.idleSince = time.Now()
	return nil
}

func (m *Manager) Config() Config             { return m.config }
func (m *Manager) Draining() <-chan struct{}  { return m.draining }
func (m *Manager) TasksDone() <-chan struct{} { return m.tasksDone }
func (m *Manager) Context() context.Context   { return m.context }

func (m *Manager) monitor() {
	defer close(m.done)
	interval := min(100*time.Millisecond, m.config.IdleTimeout/4, m.config.LeaseTimeout/4)
	timer := time.NewTicker(interval)
	defer timer.Stop()
	for {
		select {
		case now := <-timer.C:
			m.mu.Lock()
			m.maintain(now)
			m.mu.Unlock()
		case <-m.closed:
			return
		}
	}
}

func (m *Manager) maintain(now time.Time) {
	for id, value := range m.clients {
		if !now.Before(value.expires) {
			delete(m.clients, id)
		}
	}
	m.updateIdle(now)
	if m.state == "ready" && m.config.Mode == Automatic && !m.idleSince.IsZero() && now.Sub(m.idleSince) >= m.config.IdleTimeout {
		m.drain("idle_timeout")
	}
}

func (m *Manager) updateIdle(now time.Time) {
	if len(m.clients) != 0 || m.tasks != 0 {
		m.idleSince = time.Time{}
		return
	}
	if m.idleSince.IsZero() {
		m.idleSince = now
	}
}

func (m *Manager) drain(reason string) {
	if m.state != "ready" && m.state != "starting" {
		return
	}
	m.transition("draining", reason)
	close(m.draining)
	if m.tasks == 0 {
		m.tasksOnce.Do(func() { close(m.tasksDone) })
	}
}

func (m *Manager) Drain() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.drain("shutdown")
}

func (m *Manager) Stats() Stats {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.maintain(time.Now())
	return Stats{State: m.state, Clients: uint32(len(m.clients)), Tasks: m.tasks}
}

func (m *Manager) Register(id, kind string) (Lease, error) {
	if id == "" || len(id) > 128 {
		return Lease{}, fmt.Errorf("client_id должен содержать 1–128 байт")
	}
	if kind != "cli" && kind != "mcp" && kind != "ui" {
		return Lease{}, fmt.Errorf("тип клиента должен быть cli, mcp или ui")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	m.maintain(now)
	if m.state != "ready" {
		return Lease{}, ErrDraining
	}
	for leaseID, value := range m.clients {
		if value.id != id {
			continue
		}
		if value.kind != kind {
			return Lease{}, ErrClientConflict
		}
		value.expires = now.Add(m.config.LeaseTimeout)
		m.clients[leaseID] = value
		return Lease{ID: leaseID, Timeout: m.config.LeaseTimeout}, nil
	}
	if len(m.clients) >= 4096 {
		return Lease{}, ErrClientLimit
	}
	leaseID := rand.Text()
	m.clients[leaseID] = client{id: id, kind: kind, expires: now.Add(m.config.LeaseTimeout)}
	m.updateIdle(now)
	return Lease{ID: leaseID, Timeout: m.config.LeaseTimeout}, nil
}

func (m *Manager) Heartbeat(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	m.maintain(now)
	if m.state != "ready" {
		return ErrDraining
	}
	value, ok := m.clients[id]
	if !ok {
		return ErrLeaseExpired
	}
	value.expires = now.Add(m.config.LeaseTimeout)
	m.clients[id] = value
	return nil
}

func (m *Manager) Release(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.clients, id)
	m.updateIdle(time.Now())
}

// CheckLease проверяет подписчика, не создавая задачу и не продлевая аренду.
func (m *Manager) CheckLease(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.maintain(time.Now())
	if m.state != "ready" {
		return ErrDraining
	}
	if _, ok := m.clients[id]; !ok {
		return ErrLeaseExpired
	}
	return nil
}

// BeginTask удерживает сервер до вызова finish, в том числе после ухода клиента.
// Пустой leaseID предназначен для внутренних задач и старых однократных RPC.
func (m *Manager) BeginTask(leaseID string) (finish func(), err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.maintain(time.Now())
	if m.state != "ready" {
		return nil, ErrDraining
	}
	if leaseID != "" {
		if _, ok := m.clients[leaseID]; !ok {
			return nil, ErrLeaseExpired
		}
	}
	m.tasks++
	m.updateIdle(time.Now())
	var once sync.Once
	return func() { once.Do(m.finishTask) }, nil
}

func (m *Manager) finishTask() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tasks--
	if m.tasks == 0 && m.state != "ready" {
		m.tasksOnce.Do(func() { close(m.tasksDone) })
	}
	m.updateIdle(time.Now())
}

func (m *Manager) Close() {
	m.closeOnce.Do(func() {
		m.mu.Lock()
		m.drain("shutdown")
		m.transition("stopped", m.events[len(m.events)-1].Reason)
		m.mu.Unlock()
		m.cancel()
		close(m.closed)
	})
	<-m.done
}
