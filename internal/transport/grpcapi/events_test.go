package grpcapi_test

import (
	"context"
	"testing"
	"time"

	pb "github.com/Hecatoncheir/pob2_engine/gen/go/pob2/v1"
	"github.com/Hecatoncheir/pob2_engine/internal/application"
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
