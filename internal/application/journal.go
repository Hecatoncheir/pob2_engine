package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/Hecatoncheir/pob2_engine/internal/calculation"
	"github.com/Hecatoncheir/pob2_engine/internal/domain"
)

// Change хранит принятую команду и её исходный результат для безопасного повтора.
type Change struct {
	Create    *CreateSession  `json:"create,omitempty"`
	Rename    *RenameBuild    `json:"rename,omitempty"`
	Apply     *ApplyBuild     `json:"apply,omitempty"`
	Calculate *CalculateBuild `json:"calculate,omitempty"`
	Result    domain.Snapshot `json:"result"`
}

// Journal подтверждает Append только после устойчивой записи транзакции.
type Journal interface {
	Load(context.Context, func(Change) error) error
	Append(context.Context, Change) error
	Path() string
}

func OpenService(ctx context.Context, journal Journal) (*Service, error) {
	return OpenServiceWithCalculator(ctx, journal, nil)
}

func OpenServiceWithCalculator(ctx context.Context, journal Journal, calculator calculation.Calculator) (*Service, error) {
	s := NewServiceWithCalculator(calculator)
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
	var commandError *Error
	if errors.As(err, &commandError) {
		return err
	}
	if err == nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	return failure(StorageFailure, fmt.Sprintf("изменение не подтверждено хранилищем: %v", err))
}

func (s *Service) restore(ctx context.Context, change Change) error {
	if err := validateChange(change); err != nil {
		return err
	}
	result, err := s.restoreCommand(ctx, change)
	if err != nil {
		return err
	}
	if result != change.Result {
		return fmt.Errorf("снимок команды не соответствует восстановленному состоянию")
	}
	return nil
}

func validateChange(change Change) error {
	count := 0
	for _, present := range []bool{change.Create != nil, change.Rename != nil, change.Apply != nil, change.Calculate != nil} {
		if present {
			count++
		}
	}
	if count != 1 {
		return fmt.Errorf("нужна ровно одна команда в журнале")
	}
	return nil
}

func (s *Service) restoreCommand(ctx context.Context, change Change) (domain.Snapshot, error) {
	switch {
	case change.Create != nil:
		if _, exists := s.sessions[change.Create.SessionID]; exists {
			return domain.Snapshot{}, fmt.Errorf("повтор создания сессии в журнале")
		}
		return s.Create(ctx, *change.Create)
	case change.Rename != nil:
		value, exists := s.sessions[change.Rename.SessionID]
		if !exists {
			return domain.Snapshot{}, fmt.Errorf("правка неизвестной сессии в журнале")
		}
		if _, exists := value.requests[change.Rename.RequestID]; exists {
			return domain.Snapshot{}, fmt.Errorf("повтор правки в журнале")
		}
		return s.Rename(ctx, *change.Rename)
	case change.Apply != nil:
		return s.restoreApply(ctx, *change.Apply, change.Result)
	case change.Calculate != nil:
		return s.restoreCalculation(ctx, *change.Calculate, change.Result)
	default:
		return domain.Snapshot{}, fmt.Errorf("неизвестная команда в журнале")
	}
}
