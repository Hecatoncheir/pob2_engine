package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/Hecatoncheir/pob2_engine/internal/domain"
)

// Change хранит принятую команду и её исходный результат для безопасного повтора.
type Change struct {
	Create *CreateSession  `json:"create,omitempty"`
	Rename *RenameBuild    `json:"rename,omitempty"`
	Result domain.Snapshot `json:"result"`
}

// Journal подтверждает Append только после устойчивой записи транзакции.
type Journal interface {
	Load(context.Context, func(Change) error) error
	Append(context.Context, Change) error
	Path() string
}

func OpenService(ctx context.Context, journal Journal) (*Service, error) {
	s := NewService()
	if err := journal.Load(ctx, func(change Change) error { return s.restore(ctx, change) }); err != nil {
		return nil, fmt.Errorf("не удалось восстановить сессии: %w", err)
	}
	s.journal = journal
	return s, nil
}

func (s *Service) StoragePath() string {
	if s.journal == nil {
		return ""
	}
	return s.journal.Path()
}

func (s *Service) persist(ctx context.Context, change Change) error {
	if s.journal == nil {
		return nil
	}
	err := s.journal.Append(ctx, change)
	if err == nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	return failure(StorageFailure, fmt.Sprintf("изменение не подтверждено хранилищем: %v", err))
}

func (s *Service) restore(ctx context.Context, change Change) error {
	var result domain.Snapshot
	var err error
	switch {
	case change.Create != nil && change.Rename == nil:
		if _, exists := s.sessions[change.Create.SessionID]; exists {
			return fmt.Errorf("повтор создания сессии в журнале")
		}
		result, err = s.Create(ctx, *change.Create)
	case change.Rename != nil && change.Create == nil:
		value, exists := s.sessions[change.Rename.SessionID]
		if !exists {
			return fmt.Errorf("правка неизвестной сессии в журнале")
		}
		if _, exists := value.requests[change.Rename.RequestID]; exists {
			return fmt.Errorf("повтор правки в журнале")
		}
		result, err = s.Rename(ctx, *change.Rename)
	default:
		return fmt.Errorf("неизвестная команда в журнале")
	}
	if err != nil {
		return err
	}
	if result != change.Result {
		return fmt.Errorf("снимок команды не соответствует восстановленному состоянию")
	}
	return nil
}
