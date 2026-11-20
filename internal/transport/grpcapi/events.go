package grpcapi

import (
	"context"
	"crypto/subtle"
	"time"

	pb "github.com/Hecatoncheir/pob2_engine/gen/go/pob2/v1"
	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func (s *Server) authorizeStream(srv any, stream grpc.ServerStream, _ *grpc.StreamServerInfo, next grpc.StreamHandler) error {
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
	go s.watchLease(ctx, cancel, md.Get("pob2-lease")[0])
	for {
		event, err := sub.Next(ctx)
		if err != nil {
			return rpcError(err)
		}
		message := &pb.WatchSessionResponse{Snapshot: snapshot(event.Snapshot), Kind: event.Kind, RequestId: event.RequestID, PreviousBuildName: event.PreviousBuildName}
		if event.Kind != "snapshot" {
			message.ChangedFields = []string{"build_name"}
		}
		if err := stream.Send(message); err != nil {
			return err
		}
	}
}

func (s *Server) watchLease(ctx context.Context, cancel context.CancelFunc, leaseID string) {
	ticker := time.NewTicker(min(s.life.Config().LeaseTimeout/4, 100*time.Millisecond))
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.life.Draining():
			cancel()
			return
		case <-ticker.C:
			if s.life.CheckLease(leaseID) != nil {
				cancel()
				return
			}
		}
	}
}
