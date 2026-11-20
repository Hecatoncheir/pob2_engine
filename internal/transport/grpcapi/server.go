package grpcapi

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"fmt"
	"net"
	"sync"
	"time"

	pb "github.com/Hecatoncheir/pob2_engine/gen/go/pob2/v1"
	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/buildinfo"
	"github.com/Hecatoncheir/pob2_engine/internal/domain"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Ready struct {
	Address         string `json:"address"`
	Token           string `json:"token"`
	ProtocolVersion string `json:"protocol_version"`
	Mode            string `json:"mode"`
}

// Server обслуживает только loopback; состояние пока сохраняется лишь в памяти.
type Server struct {
	pb.UnimplementedSessionServiceServer
	commands application.Commands
	rpc      *grpc.Server
	listener net.Listener
	token    string
	mu       sync.RWMutex
	state    string
	stopOnce sync.Once
	done     chan struct{}
	stopped  chan struct{}
	serveErr error
}

func Start(address, token string, commands application.Commands) (*Server, error) {
	if err := validateAddress(address); err != nil {
		return nil, err
	}
	listener, err := net.Listen("tcp4", address)
	if err != nil {
		return nil, err
	}
	if token == "" {
		token = rand.Text()
	}
	s := &Server{commands: commands, listener: listener, token: token, state: "ready", done: make(chan struct{}), stopped: make(chan struct{})}
	s.rpc = grpc.NewServer(grpc.UnaryInterceptor(s.authorize))
	pb.RegisterSessionServiceServer(s.rpc, s)
	go func() { s.serveErr = s.rpc.Serve(listener); close(s.done) }()
	return s, nil
}

func validateAddress(address string) error {
	host, _, err := net.SplitHostPort(address)
	if err != nil || host != "127.0.0.1" {
		return fmt.Errorf("нужен явный адрес 127.0.0.1:порт")
	}
	return nil
}

func (s *Server) Ready() Ready {
	return Ready{Address: s.listener.Addr().String(), Token: s.token, ProtocolVersion: buildinfo.ProtocolVersion, Mode: "persistent"}
}

func (s *Server) Done() <-chan struct{} { return s.done }

func (s *Server) Err() error { <-s.done; return s.serveErr }

func (s *Server) authorize(ctx context.Context, req any, info *grpc.UnaryServerInfo, next grpc.UnaryHandler) (any, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	values := md.Get("authorization")
	if len(values) != 1 || subtle.ConstantTimeCompare([]byte(values[0]), []byte("Bearer "+s.token)) != 1 {
		return nil, status.Error(codes.Unauthenticated, "неверный токен")
	}
	s.mu.RLock()
	ready := s.state == "ready"
	s.mu.RUnlock()
	if !ready && info.FullMethod != pb.SessionService_Shutdown_FullMethodName {
		return nil, status.Error(codes.Unavailable, "сервер завершает работу")
	}
	return next(ctx, req)
}

func (s *Server) GetServerInfo(context.Context, *pb.GetServerInfoRequest) (*pb.GetServerInfoResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return &pb.GetServerInfoResponse{Version: buildinfo.Version, ProtocolVersion: buildinfo.ProtocolVersion, Capabilities: buildinfo.Capabilities(), State: s.state}, nil
}

func snapshot(value domain.Snapshot) *pb.SessionSnapshot {
	return &pb.SessionSnapshot{SessionId: value.SessionID, BuildName: value.BuildName, Revision: value.Revision}
}

func (s *Server) CreateSession(ctx context.Context, req *pb.CreateSessionRequest) (*pb.CreateSessionResponse, error) {
	value, err := s.commands.Create(ctx, application.CreateSession{SessionID: req.GetSessionId(), BuildName: req.GetBuildName(), RequestID: req.GetRequestId()})
	if err != nil {
		return nil, rpcError(err)
	}
	return &pb.CreateSessionResponse{Snapshot: snapshot(value)}, nil
}

func (s *Server) GetSnapshot(ctx context.Context, req *pb.GetSnapshotRequest) (*pb.GetSnapshotResponse, error) {
	value, err := s.commands.Snapshot(ctx, req.GetSessionId())
	if err != nil {
		return nil, rpcError(err)
	}
	return &pb.GetSnapshotResponse{Snapshot: snapshot(value)}, nil
}

func (s *Server) RenameBuild(ctx context.Context, req *pb.RenameBuildRequest) (*pb.RenameBuildResponse, error) {
	value, err := s.commands.Rename(ctx, application.RenameBuild{SessionID: req.GetSessionId(), BuildName: req.GetBuildName(), ExpectedRevision: req.GetExpectedRevision(), RequestID: req.GetRequestId()})
	if err != nil {
		return nil, rpcError(err)
	}
	return &pb.RenameBuildResponse{Snapshot: snapshot(value)}, nil
}

func (s *Server) Shutdown(context.Context, *pb.ShutdownRequest) (*pb.ShutdownResponse, error) {
	s.beginStop()
	return &pb.ShutdownResponse{}, nil
}

func (s *Server) beginStop() {
	s.stopOnce.Do(func() {
		s.mu.Lock()
		s.state = "draining"
		s.mu.Unlock()
		go func() {
			timer := time.AfterFunc(5*time.Second, s.rpc.Stop)
			defer timer.Stop()
			s.rpc.GracefulStop()
			close(s.stopped)
		}()
	})
}

// Close ограничивает ожидание штатного завершения контекстом вызывающего кода.
func (s *Server) Close(ctx context.Context) error {
	s.beginStop()
	select {
	case <-s.stopped:
		return s.Err()
	case <-ctx.Done():
		s.rpc.Stop()
		<-s.stopped
		return ctx.Err()
	}
}
