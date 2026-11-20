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
	"github.com/Hecatoncheir/pob2_engine/internal/lifecycle"
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
	InstanceID      string `json:"instance_id"`
}

// Server обслуживает только loopback; состояние пока сохраняется лишь в памяти.
type Server struct {
	pb.UnimplementedSessionServiceServer
	commands   application.Commands
	rpc        *grpc.Server
	listener   net.Listener
	token      string
	life       *lifecycle.Manager
	instanceID string
	stopOnce   sync.Once
	forceOnce  sync.Once
	force      chan struct{}
	done       chan struct{}
	stopped    chan struct{}
	serveErr   error
}

func Start(address, token string, commands application.Commands) (*Server, error) {
	return StartWithConfig(address, token, commands, lifecycle.Config{})
}

func StartWithConfig(address, token string, commands application.Commands, config lifecycle.Config) (*Server, error) {
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
	life, err := lifecycle.New(config)
	if err != nil {
		listener.Close()
		return nil, err
	}
	s := &Server{
		commands: commands, listener: listener, token: token, life: life, instanceID: rand.Text(),
		done: make(chan struct{}), stopped: make(chan struct{}), force: make(chan struct{}),
	}
	s.rpc = grpc.NewServer(grpc.UnaryInterceptor(s.authorize))
	pb.RegisterSessionServiceServer(s.rpc, s)
	go func() { s.serveErr = s.rpc.Serve(listener); close(s.done) }()
	go func() { <-life.Draining(); s.beginStop() }()
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
	return Ready{Address: s.listener.Addr().String(), Token: s.token, ProtocolVersion: buildinfo.ProtocolVersion, Mode: s.life.Config().Mode, InstanceID: s.instanceID}
}

func (s *Server) Lifecycle() *lifecycle.Manager { return s.life }

func (s *Server) Done() <-chan struct{} { return s.done }

func (s *Server) Err() error { <-s.done; return s.serveErr }

func (s *Server) authorize(ctx context.Context, req any, info *grpc.UnaryServerInfo, next grpc.UnaryHandler) (any, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	values := md.Get("authorization")
	if len(values) != 1 || subtle.ConstantTimeCompare([]byte(values[0]), []byte("Bearer "+s.token)) != 1 {
		return nil, status.Error(codes.Unauthenticated, "неверный токен")
	}
	if controlMethod(info.FullMethod) {
		return next(ctx, req)
	}
	leases := md.Get("pob2-lease")
	if len(leases) > 1 {
		return nil, status.Error(codes.InvalidArgument, "передано несколько аренд")
	}
	leaseID := ""
	if len(leases) == 1 {
		leaseID = leases[0]
	}
	finish, err := s.life.BeginTask(leaseID)
	if err != nil {
		return nil, lifecycleError(err)
	}
	defer finish()
	return next(ctx, req)
}

func controlMethod(method string) bool {
	switch method {
	case pb.SessionService_GetServerInfo_FullMethodName, pb.SessionService_RegisterClient_FullMethodName,
		pb.SessionService_Heartbeat_FullMethodName, pb.SessionService_ReleaseClient_FullMethodName, pb.SessionService_Shutdown_FullMethodName:
		return true
	default:
		return false
	}
}

func (s *Server) GetServerInfo(context.Context, *pb.GetServerInfoRequest) (*pb.GetServerInfoResponse, error) {
	stats := s.life.Stats()
	return &pb.GetServerInfoResponse{Version: buildinfo.Version, ProtocolVersion: buildinfo.ProtocolVersion, Capabilities: buildinfo.Capabilities(), State: stats.State,
		InstanceId: s.instanceID, Mode: s.life.Config().Mode, ClientCount: stats.Clients, ActiveTasks: stats.Tasks, LeaseTimeoutMs: uint64(s.life.Config().LeaseTimeout.Milliseconds())}, nil
}

func (s *Server) RegisterClient(_ context.Context, req *pb.RegisterClientRequest) (*pb.RegisterClientResponse, error) {
	lease, err := s.life.Register(req.GetClientId(), req.GetKind())
	if err != nil {
		return nil, lifecycleError(err)
	}
	return &pb.RegisterClientResponse{LeaseId: lease.ID, LeaseTimeoutMs: uint64(lease.Timeout.Milliseconds())}, nil
}

func (s *Server) Heartbeat(_ context.Context, req *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	if err := s.life.Heartbeat(req.GetLeaseId()); err != nil {
		return nil, lifecycleError(err)
	}
	return &pb.HeartbeatResponse{}, nil
}

func (s *Server) ReleaseClient(_ context.Context, req *pb.ReleaseClientRequest) (*pb.ReleaseClientResponse, error) {
	s.life.Release(req.GetLeaseId())
	return &pb.ReleaseClientResponse{}, nil
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
		s.life.Drain()
		go func() {
			s.waitForStop()
			s.life.Close()
			close(s.stopped)
		}()
	})
}

func (s *Server) waitForStop() {
	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()
	graceful := make(chan struct{})
	go func() { s.rpc.GracefulStop(); close(graceful) }()
	var rpcDone <-chan struct{} = graceful
	tasksDone := s.life.TasksDone()
	for rpcDone != nil || tasksDone != nil {
		select {
		case <-rpcDone:
			rpcDone = nil
		case <-tasksDone:
			tasksDone = nil
		case <-timer.C:
			s.rpc.Stop()
			return
		case <-s.force:
			s.rpc.Stop()
			return
		}
	}
}

// Close ограничивает ожидание штатного завершения контекстом вызывающего кода.
func (s *Server) Close(ctx context.Context) error {
	s.beginStop()
	select {
	case <-s.stopped:
		return s.Err()
	case <-ctx.Done():
		s.forceOnce.Do(func() { close(s.force) })
		<-s.stopped
		return ctx.Err()
	}
}
