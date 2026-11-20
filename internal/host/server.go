package host

import (
	"context"
	"errors"
	"io"
	"os"
	"sync"
	"time"

	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/discovery"
	"github.com/Hecatoncheir/pob2_engine/internal/lifecycle"
	"github.com/Hecatoncheir/pob2_engine/internal/storage"
	"github.com/Hecatoncheir/pob2_engine/internal/transport/grpcapi"
)

// Server владеет блокировкой профиля до завершения RPC и удаления своего файла.
type Server struct {
	API     *grpcapi.Server
	profile discovery.Profile
	lock    *discovery.Lock
	store   io.Closer
	once    sync.Once
	done    chan struct{}
	err     error
}

func Start(profile discovery.Profile, address, token string, commands application.Commands, config lifecycle.Config) (*Server, error) {
	return start(profile, address, token, config, func() (application.Commands, io.Closer, error) { return commands, nil, nil })
}

func StartDurable(ctx context.Context, profile discovery.Profile, address, token, dataDir string, config lifecycle.Config) (*Server, error) {
	return start(profile, address, token, config, func() (application.Commands, io.Closer, error) {
		journal, err := storage.Open(ctx, dataDir, profile.Name)
		if err != nil {
			return nil, nil, err
		}
		service, err := application.OpenService(ctx, journal)
		if err != nil {
			journal.Close()
			return nil, nil, err
		}
		return service, journal, nil
	})
}

type commandFactory func() (application.Commands, io.Closer, error)

func start(profile discovery.Profile, address, token string, config lifecycle.Config, factory commandFactory) (*Server, error) {
	if err := grpcapi.ValidateAddress(address); err != nil {
		return nil, err
	}
	if _, err := config.Normalize(); err != nil {
		return nil, err
	}
	lock, err := profile.TryLock("server")
	if err != nil {
		return nil, err
	}
	commands, store, err := factory()
	if err != nil {
		lock.Close()
		return nil, err
	}
	closeStore := func() {
		if store != nil {
			_ = store.Close()
		}
	}
	api, err := grpcapi.StartWithConfig(address, token, commands, config)
	if err != nil {
		closeStore()
		lock.Close()
		return nil, err
	}
	ready := api.Ready()
	record := discovery.Record{InstanceID: ready.InstanceID, Address: ready.Address, Token: ready.Token, ProtocolVersion: ready.ProtocolVersion, Mode: ready.Mode, PID: os.Getpid()}
	if err := profile.Publish(record); err != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
		defer cancel()
		api.Close(ctx)
		closeStore()
		lock.Close()
		return nil, err
	}
	s := &Server{API: api, profile: profile, lock: lock, store: store, done: make(chan struct{})}
	go func() {
		<-api.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
		defer cancel()
		_ = s.Close(ctx)
	}()
	return s, nil
}

func (s *Server) Ready() grpcapi.Ready  { return s.API.Ready() }
func (s *Server) Done() <-chan struct{} { return s.done }
func (s *Server) Err() error            { <-s.done; return s.err }

func (s *Server) Close(ctx context.Context) error {
	s.once.Do(func() {
		apiErr := s.API.Close(ctx)
		var storeErr error
		if s.store != nil {
			storeErr = s.store.Close()
		}
		s.err = errors.Join(apiErr, storeErr, s.profile.RemoveOwn(s.Ready().InstanceID), s.lock.Close())
		close(s.done)
	})
	return s.err
}
