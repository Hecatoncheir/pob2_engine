package application

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/Hecatoncheir/pob2_engine/internal/domain"
)

const (
	maxSessions = 1024
	maxRequests = 4096
)

var identifier = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$`)

type CreateSession struct {
	SessionID string `json:"session_id"`
	BuildName string `json:"build_name"`
	RequestID string `json:"request_id"`
}

type RenameBuild struct {
	SessionID        string `json:"session_id"`
	BuildName        string `json:"build_name"`
	ExpectedRevision uint64 `json:"expected_revision"`
	RequestID        string `json:"request_id"`
}

// Commands — общий контракт для локального сервиса и удалённого клиента.
type Commands interface {
	Create(context.Context, CreateSession) (domain.Snapshot, error)
	Snapshot(context.Context, string) (domain.Snapshot, error)
	Rename(context.Context, RenameBuild) (domain.Snapshot, error)
}

type creation struct {
	command CreateSession
	result  domain.Snapshot
}

type rename struct {
	command RenameBuild
	result  domain.Snapshot
}

type session struct {
	mu              sync.Mutex
	snapshot        domain.Snapshot
	createRequestID string
	requests        map[string]rename
	events          []Event
	changed         chan struct{}
}

// Service публикует изменения в памяти после подтверждения журнала, если он задан.
type Service struct {
	mu       sync.RWMutex
	sessions map[string]*session
	requests map[string]creation
	journal  Journal
}

func NewService() *Service {
	return &Service{sessions: make(map[string]*session), requests: make(map[string]creation)}
}

func validate(sessionID, name, requestID string) error {
	if !identifier.MatchString(sessionID) || !identifier.MatchString(requestID) {
		return failure(InvalidArgument, "session_id и request_id: 1–128 символов ASCII, буквы, цифры, точка, дефис или подчёркивание; первый символ — буква или цифра")
	}
	if !utf8.ValidString(name) || strings.TrimSpace(name) == "" || len(name) > 512 {
		return failure(InvalidArgument, "название сборки должно содержать 1–512 байт корректного UTF-8")
	}
	return nil
}

func (s *Service) Create(ctx context.Context, cmd CreateSession) (domain.Snapshot, error) {
	if err := validate(cmd.SessionID, cmd.BuildName, cmd.RequestID); err != nil {
		return domain.Snapshot{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return domain.Snapshot{}, err
	}
	if previous, ok := s.requests[cmd.RequestID]; ok {
		if previous.command != cmd {
			return domain.Snapshot{}, failure(RequestConflict, "request_id уже использован другой командой создания")
		}
		return previous.result, nil
	}
	if _, ok := s.sessions[cmd.SessionID]; ok {
		return domain.Snapshot{}, failure(AlreadyExists, "сессия уже существует")
	}
	if len(s.sessions) >= maxSessions || len(s.requests) >= maxRequests {
		return domain.Snapshot{}, failure(ResourceExhausted, "достигнут лимит сессий или команд")
	}
	result := domain.Snapshot{SessionID: cmd.SessionID, BuildName: cmd.BuildName, Revision: 1}
	if err := s.persist(ctx, Change{Create: &cmd, Result: result}); err != nil {
		return domain.Snapshot{}, err
	}
	s.sessions[cmd.SessionID] = &session{snapshot: result, createRequestID: cmd.RequestID, requests: make(map[string]rename),
		events: []Event{{Snapshot: result, Kind: "created", RequestID: cmd.RequestID}}, changed: make(chan struct{})}
	s.requests[cmd.RequestID] = creation{command: cmd, result: result}
	return result, nil
}

func (s *Service) find(id string) (*session, error) {
	if !identifier.MatchString(id) {
		return nil, failure(InvalidArgument, "некорректный session_id")
	}
	s.mu.RLock()
	value, ok := s.sessions[id]
	s.mu.RUnlock()
	if !ok {
		return nil, failure(NotFound, "сессия не найдена")
	}
	return value, nil
}

func (s *Service) Snapshot(ctx context.Context, id string) (domain.Snapshot, error) {
	value, err := s.find(id)
	if err != nil {
		return domain.Snapshot{}, err
	}
	value.mu.Lock()
	defer value.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return domain.Snapshot{}, err
	}
	return value.snapshot, nil
}

func (s *Service) Rename(ctx context.Context, cmd RenameBuild) (domain.Snapshot, error) {
	if err := validate(cmd.SessionID, cmd.BuildName, cmd.RequestID); err != nil {
		return domain.Snapshot{}, err
	}
	value, err := s.find(cmd.SessionID)
	if err != nil {
		return domain.Snapshot{}, err
	}
	value.mu.Lock()
	defer value.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return domain.Snapshot{}, err
	}
	if cmd.RequestID == value.createRequestID {
		return domain.Snapshot{}, failure(RequestConflict, "request_id уже использован для создания сессии")
	}
	if previous, ok := value.requests[cmd.RequestID]; ok {
		if previous.command != cmd {
			return domain.Snapshot{}, failure(RequestConflict, "request_id уже использован другой правкой")
		}
		return previous.result, nil
	}
	if cmd.ExpectedRevision != value.snapshot.Revision {
		return domain.Snapshot{}, failure(RevisionConflict, fmt.Sprintf("ожидалась ревизия %d, текущая %d", cmd.ExpectedRevision, value.snapshot.Revision))
	}
	if len(value.requests) >= maxRequests {
		return domain.Snapshot{}, failure(ResourceExhausted, "достигнут лимит команд сессии")
	}
	result := domain.Snapshot{SessionID: cmd.SessionID, BuildName: cmd.BuildName, Revision: value.snapshot.Revision + 1}
	if err := s.persist(ctx, Change{Rename: &cmd, Result: result}); err != nil {
		return domain.Snapshot{}, err
	}
	previousName := value.snapshot.BuildName
	value.snapshot = result
	value.requests[cmd.RequestID] = rename{command: cmd, result: value.snapshot}
	value.events = append(value.events, Event{Snapshot: result, Kind: "renamed", RequestID: cmd.RequestID, PreviousBuildName: previousName})
	close(value.changed)
	value.changed = make(chan struct{})
	return value.snapshot, nil
}
