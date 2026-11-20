package lifecycle

import (
	"context"
	"io"
)

type Event struct {
	Sequence uint64
	State    string
	Reason   string
}

// transition вызывается только под mutex; переходов не больше четырёх.
func (m *Manager) transition(state, reason string) {
	m.state = state
	m.events = append(m.events, Event{Sequence: uint64(len(m.events) + 1), State: state, Reason: reason})
	close(m.changed)
	m.changed = make(chan struct{})
}

// Subscription возвращает текущую фазу и все последующие переходы без разрыва.
// Next вызывается одним читателем. История принадлежит одному запуску сервера.
type Subscription struct {
	manager *Manager
	cursor  int
}

func (m *Manager) Watch() *Subscription {
	m.mu.Lock()
	defer m.mu.Unlock()
	return &Subscription{manager: m, cursor: len(m.events) - 1}
}

func (s *Subscription) Next(ctx context.Context) (Event, error) {
	for {
		if err := ctx.Err(); err != nil {
			return Event{}, err
		}
		s.manager.mu.Lock()
		if s.cursor < len(s.manager.events) {
			event := s.manager.events[s.cursor]
			s.cursor++
			s.manager.mu.Unlock()
			return event, nil
		}
		if s.manager.state == "stopped" {
			s.manager.mu.Unlock()
			return Event{}, io.EOF
		}
		changed := s.manager.changed
		s.manager.mu.Unlock()
		select {
		case <-changed:
		case <-ctx.Done():
			return Event{}, ctx.Err()
		}
	}
}
