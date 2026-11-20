package application

import (
	"context"

	"github.com/Hecatoncheir/pob2_engine/internal/domain"
)

// Event описывает подтверждённую правку или начальный снимок подписки.
type Event struct {
	Snapshot          domain.Snapshot
	Kind              string
	RequestID         string
	PreviousBuildName string
	ChangedFields     string
}

type Events interface {
	Watch(context.Context, string, uint64) (*Subscription, error)
}

// Subscription читается одним потребителем. История общая для всех подписчиков
// и ограничена лимитом команд сессии; медленный клиент не задерживает запись.
type Subscription struct {
	session *session
	cursor  uint64
	initial *Event
}

func (s *Service) Watch(ctx context.Context, id string, after uint64) (*Subscription, error) {
	value, err := s.find(id)
	if err != nil {
		return nil, err
	}
	value.mu.Lock()
	defer value.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if after > value.snapshot.Revision {
		return nil, failure(RevisionConflict, "ревизия подписки опережает сессию; запросите начальный снимок")
	}
	sub := &Subscription{session: value, cursor: after}
	if after == 0 {
		sub.initial = &Event{Snapshot: value.snapshot, Kind: "snapshot"}
		sub.cursor = value.snapshot.Revision
	}
	return sub, nil
}

func (s *Subscription) Next(ctx context.Context) (Event, error) {
	for {
		if err := ctx.Err(); err != nil {
			return Event{}, err
		}
		if s.initial != nil {
			event := *s.initial
			s.initial = nil
			return event, nil
		}
		s.session.mu.Lock()
		if s.cursor < uint64(len(s.session.events)) {
			event := s.session.events[s.cursor]
			s.cursor++
			s.session.mu.Unlock()
			return event, nil
		}
		changed := s.session.changed
		s.session.mu.Unlock()
		select {
		case <-changed:
		case <-ctx.Done():
			return Event{}, ctx.Err()
		}
	}
}
