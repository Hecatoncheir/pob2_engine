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
	"github.com/Hecatoncheir/pob2_engine/internal/lifecycle"
	"github.com/Hecatoncheir/pob2_engine/internal/maintenance"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const gracefulStopTimeout = 5 * time.Second

type Ready struct {
	Address         string `json:"address"`
	Token           string `json:"token"`
	ProtocolVersion string `json:"protocol_version"`
	Mode            string `json:"mode"`
	InstanceID      string `json:"instance_id"`
	StorageMode     string `json:"storage_mode"`
	StoragePath     string `json:"storage_path,omitempty"`
}

// Server обслуживает loopback; способ хранения задаётся общим сервисом команд.
type Server struct {
	pb.UnimplementedSessionServiceServer
	commands    application.Commands
	maintenance maintenance.Commands
	rpc         *grpc.Server
	listener    net.Listener
	token       string
	life        *lifecycle.Manager
	instanceID  string
	stopOnce    sync.Once
	forceOnce   sync.Once
	force       chan struct{}
	done        chan struct{}
	stopped     chan struct{}
	serveErr    error
}

func Start(address, token string, commands application.Commands) (*Server, error) {
	return StartWithConfig(address, token, commands, lifecycle.Config{})
}

func StartWithConfig(address, token string, commands application.Commands, config lifecycle.Config) (*Server, error) {
	if err := ValidateAddress(address); err != nil {
		return nil, err
	}
	listener, err := net.Listen("tcp4", address)
	if err != nil {
		return nil, err
	}
	if token == "" {
		token = rand.Text()
	}
	life, err := lifecycle.NewStarting(config)
	if err != nil {
		listener.Close()
		return nil, err
	}
	s := &Server{
		commands: commands, maintenance: maintenance.Local(), listener: listener, token: token, life: life, instanceID: rand.Text(),
		done: make(chan struct{}), stopped: make(chan struct{}), force: make(chan struct{}),
	}
	s.rpc = grpc.NewServer(grpc.MaxRecvMsgSize(maxMessageSize), grpc.MaxSendMsgSize(maxMessageSize), grpc.UnaryInterceptor(s.authorize), grpc.StreamInterceptor(s.authorizeStream))
	pb.RegisterSessionServiceServer(s.rpc, s)
	if err := life.MarkReady(); err != nil {
		listener.Close()
		life.Close()
		return nil, err
	}
	go func() { s.serveErr = s.rpc.Serve(listener); close(s.done) }()
	go func() { <-life.Draining(); s.beginStop() }()
	return s, nil
}

func ValidateAddress(address string) error {
	host, _, err := net.SplitHostPort(address)
	if err != nil || host != "127.0.0.1" {
		return fmt.Errorf("нужен явный адрес 127.0.0.1:порт")
	}
	return nil
}

func (s *Server) Ready() Ready {
	mode, path := s.storageInfo()
	return Ready{Address: s.listener.Addr().String(), Token: s.token, ProtocolVersion: buildinfo.ProtocolVersion, Mode: s.life.Config().Mode, InstanceID: s.instanceID, StorageMode: mode, StoragePath: path}
}

func (s *Server) storageInfo() (string, string) {
	if source, ok := s.commands.(interface{ StoragePath() string }); ok {
		if path := source.StoragePath(); path != "" {
			return "disk", path
		}
	}
	return "memory", ""
}

func (s *Server) Lifecycle() *lifecycle.Manager { return s.life }

func (s *Server) Done() <-chan struct{}     { return s.done }
func (s *Server) Draining() <-chan struct{} { return s.life.Draining() }

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
	case pb.SessionService_GetServerInfo_FullMethodName, pb.SessionService_Doctor_FullMethodName, pb.SessionService_RegisterClient_FullMethodName,
		pb.SessionService_Heartbeat_FullMethodName, pb.SessionService_ReleaseClient_FullMethodName, pb.SessionService_Shutdown_FullMethodName:
		return true
	default:
		return false
	}
}

func (s *Server) GetServerInfo(context.Context, *pb.GetServerInfoRequest) (*pb.GetServerInfoResponse, error) {
	stats := s.life.Stats()
	storageMode, storagePath := s.storageInfo()
	capabilities := buildinfo.Capabilities()
	capabilities = append(capabilities, "server.events", maintenance.UpdateCapability, maintenance.DoctorCapability)
	if _, ok := s.commands.(application.Events); ok {
		capabilities = append(capabilities, "session.events")
	}
	if storageMode == "disk" {
		capabilities = append(capabilities, "session.persistence")
	}
	if _, ok := s.commands.(application.BuildCommands); ok {
		capabilities = append(capabilities, "build.import", "build.edit", "build.add_support", "build.calculate", "build.compare", "build.export", "calculation.progress")
	}
	if _, ok := s.commands.(application.CatalogCommands); ok {
		capabilities = append(capabilities, "catalog.search", "build.preview_support")
	}
	if preview, ok := s.commands.(interface{ SupportsEquipmentPreview() bool }); ok && preview.SupportsEquipmentPreview() {
		capabilities = append(capabilities, "build.equipment_preview")
	}
	if preview, ok := s.commands.(interface{ SupportsTreePreview() bool }); ok && preview.SupportsTreePreview() {
		capabilities = append(capabilities, "build.tree_preview", "build.tree_attributes", "build.tree_allocation_modes", "build.tree_sockets")
	}
	if preview, ok := s.commands.(interface{ SupportsJewelPreview() bool }); ok && preview.SupportsJewelPreview() {
		capabilities = append(capabilities, "build.tree_jewels")
	}
	return &pb.GetServerInfoResponse{Version: buildinfo.Version, ProtocolVersion: buildinfo.ProtocolVersion, Capabilities: capabilities, State: stats.State,
		InstanceId: s.instanceID, Mode: s.life.Config().Mode, ClientCount: stats.Clients, ActiveTasks: stats.Tasks, LeaseTimeoutMs: uint64(s.life.Config().LeaseTimeout.Milliseconds()), StorageMode: storageMode, StoragePath: storagePath}, nil
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

func (s *Server) CreateSession(ctx context.Context, req *pb.CreateSessionRequest) (*pb.CreateSessionResponse, error) {
	message, err := s.createSession(ctx, application.CreateSession{SessionID: req.GetSessionId(), BuildName: req.GetBuildName(), RequestID: req.GetRequestId()})
	return &pb.CreateSessionResponse{Snapshot: message}, err
}

func (s *Server) CreateSessionFromXml(ctx context.Context, req *pb.CreateSessionFromXmlRequest) (*pb.CreateSessionFromXmlResponse, error) {
	if req.GetXml() == "" {
		return nil, status.Error(codes.InvalidArgument, "нужен XML сборки")
	}
	message, err := s.createSession(ctx, application.CreateSession{SessionID: req.GetSessionId(), BuildName: req.GetBuildName(), RequestID: req.GetRequestId(), XML: req.GetXml()})
	return &pb.CreateSessionFromXmlResponse{Snapshot: message}, err
}

func (s *Server) createSession(ctx context.Context, cmd application.CreateSession) (*pb.SessionSnapshot, error) {
	value, err := s.commands.Create(ctx, cmd)
	if err != nil {
		return nil, rpcError(err)
	}
	message, err := snapshot(value)
	return message, rpcError(err)
}

func (s *Server) GetSnapshot(ctx context.Context, req *pb.GetSnapshotRequest) (*pb.GetSnapshotResponse, error) {
	value, err := s.commands.Snapshot(ctx, req.GetSessionId())
	if err != nil {
		return nil, rpcError(err)
	}
	message, err := snapshot(value)
	return &pb.GetSnapshotResponse{Snapshot: message}, rpcError(err)
}

func (s *Server) RenameBuild(ctx context.Context, req *pb.RenameBuildRequest) (*pb.RenameBuildResponse, error) {
	value, err := s.commands.Rename(ctx, application.RenameBuild{SessionID: req.GetSessionId(), BuildName: req.GetBuildName(), ExpectedRevision: req.GetExpectedRevision(), RequestID: req.GetRequestId()})
	if err != nil {
		return nil, rpcError(err)
	}
	message, err := snapshot(value)
	return &pb.RenameBuildResponse{Snapshot: message}, rpcError(err)
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
	timer := time.NewTimer(gracefulStopTimeout)
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
