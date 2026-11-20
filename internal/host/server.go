package host

import (
	"context"
	"errors"
	"os"
	"sync"
	"time"

	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/discovery"
	"github.com/Hecatoncheir/pob2_engine/internal/lifecycle"
	"github.com/Hecatoncheir/pob2_engine/internal/transport/grpcapi"
)

// Server владеет блокировкой профиля до завершения RPC и удаления своего файла.
type Server struct {
	API     *grpcapi.Server
	profile discovery.Profile
	lock    *discovery.Lock
	once    sync.Once
	done    chan struct{}
	err     error
}

func Start(profile discovery.Profile, address, token string, commands application.Commands, config lifecycle.Config) (*Server, error) {
	lock, err := profile.TryLock("server")
	if err != nil {
		return nil, err
	}
	api, err := grpcapi.StartWithConfig(address, token, commands, config)
	if err != nil {
		lock.Close()
		return nil, err
	}
	ready := api.Ready()
	record := discovery.Record{InstanceID: ready.InstanceID, Address: ready.Address, Token: ready.Token, ProtocolVersion: ready.ProtocolVersion, Mode: ready.Mode, PID: os.Getpid()}
	if err := profile.Publish(record); err != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
		defer cancel()
		api.Close(ctx)
		lock.Close()
		return nil, err
	}
	s := &Server{API: api, profile: profile, lock: lock, done: make(chan struct{})}
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
		s.err = errors.Join(s.API.Close(ctx), s.profile.RemoveOwn(s.Ready().InstanceID), s.lock.Close())
		close(s.done)
	})
	return s.err
}
