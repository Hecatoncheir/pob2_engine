package grpcapi

import (
	"context"
	"crypto/subtle"
	"strings"
	"time"

	pb "github.com/Hecatoncheir/pob2_engine/gen/go/pob2/v1"
	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/lifecycle"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func (s *Server) authorizeStream(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, next grpc.StreamHandler) error {
	md, _ := metadata.FromIncomingContext(stream.Context())
	tokens := md.Get("authorization")
	if len(tokens) != 1 || subtle.ConstantTimeCompare([]byte(tokens[0]), []byte("Bearer "+s.token)) != 1 {
		return status.Error(codes.Unauthenticated, "неверный токен")
	}
	leases := md.Get("pob2-lease")
	if len(leases) != 1 || leases[0] == "" {
		return status.Error(codes.FailedPrecondition, "для подписки нужна аренда клиента")
	}
	if err := s.life.CheckLease(leases[0]); err != nil {
		return lifecycleError(err)
	}
	if info.FullMethod == pb.SessionService_CalculateBuild_FullMethodName || info.FullMethod == pb.SessionService_WatchTreePower_FullMethodName ||
		info.FullMethod == pb.SessionService_GetTreeTextures_FullMethodName {
		finish, err := s.life.BeginTask(leases[0])
		if err != nil {
			return lifecycleError(err)
		}
		defer finish()
	}
	return next(srv, stream)
}

func (s *Server) WatchSession(req *pb.WatchSessionRequest, stream grpc.ServerStreamingServer[pb.WatchSessionResponse]) error {
	source, ok := s.commands.(application.Events)
	if !ok {
		return status.Error(codes.Unimplemented, "сервис не поддерживает события")
	}
	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()
	sub, err := source.Watch(ctx, req.GetSessionId(), req.GetAfterRevision())
	if err != nil {
		return rpcError(err)
	}
	md, _ := metadata.FromIncomingContext(ctx)
	go s.watchLease(ctx, cancel, md.Get("pob2-lease")[0], true)
	for {
		event, err := sub.Next(ctx)
		if err != nil {
			return rpcError(err)
		}
		value, err := snapshot(event.Snapshot)
		if err != nil {
			return rpcError(err)
		}
		message := &pb.WatchSessionResponse{Snapshot: value, Kind: event.Kind, RequestId: event.RequestID, PreviousBuildName: event.PreviousBuildName}
		if event.ChangedFields != "" {
			message.ChangedFields = strings.Split(event.ChangedFields, ",")
		} else if event.Kind != "snapshot" {
			message.ChangedFields = []string{"build_name"}
		}
		if err := stream.Send(message); err != nil {
			return err
		}
	}
}

func (s *Server) WatchServer(_ *pb.WatchServerRequest, stream grpc.ServerStreamingServer[pb.WatchServerResponse]) error {
	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()
	md, _ := metadata.FromIncomingContext(ctx)
	go s.watchLease(ctx, cancel, md.Get("pob2-lease")[0], false)
	sub := s.life.Watch()
	for {
		event, err := sub.Next(ctx)
		if err != nil {
			return rpcError(err)
		}
		if err := stream.Send(&pb.WatchServerResponse{InstanceId: s.instanceID, Sequence: event.Sequence, State: event.State, Reason: event.Reason, ShutdownTimeoutMs: uint64(gracefulStopTimeout.Milliseconds())}); err != nil {
			return err
		}
		if event.State == "draining" || event.State == "stopped" {
			return nil
		}
	}
}

func (s *Server) watchLease(ctx context.Context, cancel context.CancelFunc, leaseID string, stopOnDrain bool) {
	// Поток сервера отправляет последнюю фазу сам; поток сессии отменяется сразу.
	ticker := time.NewTicker(min(s.life.Config().LeaseTimeout/4, 100*time.Millisecond))
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.life.Draining():
			if stopOnDrain {
				cancel()
			}
			return
		case <-ticker.C:
			if err := s.life.CheckLease(leaseID); err != nil {
				if err != lifecycle.ErrDraining || stopOnDrain {
					cancel()
				}
				return
			}
		}
	}
}
