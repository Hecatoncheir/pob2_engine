package grpcapi_test

import (
	"context"
	"testing"
	"time"

	pb "github.com/Hecatoncheir/pob2_engine/gen/go/pob2/v1"
	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/lifecycle"
	"github.com/Hecatoncheir/pob2_engine/internal/transport/grpcapi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestWatchAuthorizationReplayAndShutdown(t *testing.T) {
	s, c := start(t, application.NewService())
	ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
	defer cancel()
	if _, err := c.Create(ctx, application.CreateSession{SessionID: "main", BuildName: "До", RequestID: "create"}); err != nil {
		t.Fatal(err)
	}
	raw, err := grpc.NewClient("passthrough:///"+s.Ready().Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	defer raw.Close()
	for _, test := range []struct {
		token, lease string
		code         codes.Code
	}{
		{"wrong", c.LeaseID(), codes.Unauthenticated}, {s.Ready().Token, "", codes.FailedPrecondition}, {s.Ready().Token, "missing", codes.FailedPrecondition},
	} {
		callCtx := metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+test.token, "pob2-lease", test.lease)
		stream, err := pb.NewSessionServiceClient(raw).WatchSession(callCtx, &pb.WatchSessionRequest{SessionId: "main"})
		if err == nil {
			_, err = stream.Recv()
		}
		if status.Code(err) != test.code {
			t.Fatalf("доступ: %v", err)
		}
	}
	stream, err := c.Watch(ctx, "main", 0)
	if err != nil {
		t.Fatal(err)
	}
	initial, err := stream.Recv()
	if err != nil || initial.GetKind() != "snapshot" {
		t.Fatalf("снимок: %v %v", initial, err)
	}
	cmd := application.RenameBuild{SessionID: "main", BuildName: "После", ExpectedRevision: 1, RequestID: "rename"}
	for range 2 {
		if _, err := c.Rename(ctx, cmd); err != nil {
			t.Fatal(err)
		}
	}
	changed, err := stream.Recv()
	if err != nil || changed.GetSnapshot().GetRevision() != 2 || changed.GetPreviousBuildName() != "До" || changed.GetChangedFields()[0] != "build_name" {
		t.Fatalf("правка: %v %v", changed, err)
	}
	if stats := s.Lifecycle().Stats(); stats.Tasks != 0 {
		t.Fatalf("подписка удерживает задачу: %+v", stats)
	}
	if err := s.Close(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := stream.Recv(); err == nil {
		t.Fatal("подписка не закрыта при остановке")
	}
}

func TestPersistentWatchEndsWhenLeaseExpires(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	defer cancel()
	service := application.NewService()
	if _, err := service.Create(ctx, application.CreateSession{SessionID: "main", BuildName: "Начальная", RequestID: "create"}); err != nil {
		t.Fatal(err)
	}
	s, err := grpcapi.StartWithConfig("127.0.0.1:0", "token", service, lifecycle.Config{LeaseTimeout: 100 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close(ctx)
	raw, err := grpc.NewClient("passthrough:///"+s.Ready().Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	defer raw.Close()
	rpc := pb.NewSessionServiceClient(raw)
	auth := metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer token")
	lease, err := rpc.RegisterClient(auth, &pb.RegisterClientRequest{ClientId: "ui", Kind: "ui"})
	if err != nil {
		t.Fatal(err)
	}
	watch, err := rpc.WatchSession(metadata.AppendToOutgoingContext(auth, "pob2-lease", lease.GetLeaseId()), &pb.WatchSessionRequest{SessionId: "main"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := watch.Recv(); err != nil {
		t.Fatal(err)
	}
	if _, err := watch.Recv(); status.Code(err) != codes.Canceled {
		t.Fatalf("истечение аренды: %v", err)
	}
	if stats := s.Lifecycle().Stats(); stats.State != "ready" || stats.Tasks != 0 || stats.Clients != 0 {
		t.Fatalf("подписка удерживает сервер: %+v", stats)
	}
}
