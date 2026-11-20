package grpcapi_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
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

func TestServerWatchAnnouncesShutdownAndEndsBeforeBackgroundTask(t *testing.T) {
	s, c := start(t, application.NewService())
	ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
	defer cancel()
	watch, err := c.WatchServer(ctx)
	if err != nil {
		t.Fatal(err)
	}
	initial, err := watch.Recv()
	if err != nil || initial.GetState() != "ready" || initial.GetInstanceId() != c.InstanceID() || initial.GetShutdownTimeoutMs() != 5000 {
		t.Fatalf("начальная фаза: %v %v", initial, err)
	}
	finish, err := s.Lifecycle().BeginTask("")
	if err != nil {
		t.Fatal(err)
	}
	defer finish()
	if err := c.Shutdown(ctx); err != nil {
		t.Fatal(err)
	}
	stopping, err := watch.Recv()
	if err != nil || stopping.GetState() != "draining" || stopping.GetReason() != "shutdown" || stopping.GetSequence() != initial.GetSequence()+1 {
		t.Fatalf("уведомление: %v %v", stopping, err)
	}
	if _, err := watch.Recv(); !errors.Is(err, io.EOF) {
		t.Fatalf("поток не завершён штатно: %v", err)
	}
	if tasks := s.Lifecycle().Stats().Tasks; tasks != 1 {
		t.Fatalf("подписка стала задачей или потеряна фоновая работа: %d", tasks)
	}
	finish()
	if err := s.Close(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestServerWatchAuthorizationAndLeaseRelease(t *testing.T) {
	s, c := start(t, application.NewService())
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	defer cancel()
	raw, err := grpc.NewClient("passthrough:///"+s.Ready().Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	defer raw.Close()
	rpc := pb.NewSessionServiceClient(raw)
	for _, test := range []struct {
		token, lease string
		code         codes.Code
	}{
		{"wrong", c.LeaseID(), codes.Unauthenticated}, {s.Ready().Token, "", codes.FailedPrecondition}, {s.Ready().Token, "expired", codes.FailedPrecondition},
	} {
		md := metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+test.token, "pob2-lease", test.lease)
		watch, err := rpc.WatchServer(md, &pb.WatchServerRequest{})
		if err == nil {
			_, err = watch.Recv()
		}
		if status.Code(err) != test.code {
			t.Fatalf("доступ: %v", err)
		}
	}
	watch, err := c.WatchServer(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := watch.Recv(); err != nil {
		t.Fatal(err)
	}
	auth := metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+s.Ready().Token)
	if _, err := rpc.ReleaseClient(auth, &pb.ReleaseClientRequest{LeaseId: c.LeaseID()}); err != nil {
		t.Fatal(err)
	}
	if _, err := watch.Recv(); status.Code(err) != codes.Canceled {
		t.Fatalf("подписка освобождённого клиента: %v", err)
	}
}

func TestServerWatchAndShutdownRace(t *testing.T) {
	for range 20 {
		s, c := start(t, application.NewService())
		ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
		var group sync.WaitGroup
		group.Go(func() { _ = c.Shutdown(ctx) })
		watch, err := c.WatchServer(ctx)
		if err == nil {
			for {
				message, recvErr := watch.Recv()
				if recvErr != nil {
					break
				}
				if message.GetState() != "ready" && message.GetState() != "draining" {
					t.Errorf("неизвестная фаза: %v", message)
				}
			}
		}
		group.Wait()
		if err := s.Close(ctx); err != nil {
			t.Fatal(err)
		}
		cancel()
	}
}

func TestSlowSessionSubscriberDoesNotBlockForcedShutdown(t *testing.T) {
	service := application.NewService()
	s, c := start(t, service)
	ctx, cancel := context.WithTimeout(t.Context(), 4*time.Second)
	defer cancel()
	if _, err := c.Create(ctx, application.CreateSession{SessionID: "slow", BuildName: "Начальная", RequestID: "create"}); err != nil {
		t.Fatal(err)
	}
	watch, err := c.Watch(ctx, "slow", 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := watch.Header(); err != nil {
		t.Fatal(err)
	}
	for i := 1; i <= 2048; i++ {
		_, err := service.Rename(ctx, application.RenameBuild{SessionID: "slow", BuildName: strings.Repeat("x", 500), ExpectedRevision: uint64(i), RequestID: fmt.Sprint(i)})
		if err != nil {
			t.Fatal(err)
		}
	}
	stopCtx, stopCancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
	defer stopCancel()
	stopped := make(chan error, 1)
	go func() { stopped <- s.Close(stopCtx) }()
	select {
	case err := <-stopped:
		if err != nil && !errors.Is(err, context.DeadlineExceeded) {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("медленный клиент заблокировал остановку")
	}
}
