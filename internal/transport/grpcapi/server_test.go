package grpcapi_test

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	pb "github.com/Hecatoncheir/pob2_engine/gen/go/pob2/v1"
	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/domain"
	"github.com/Hecatoncheir/pob2_engine/internal/lifecycle"
	"github.com/Hecatoncheir/pob2_engine/internal/transport/grpcapi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func start(t *testing.T, commands application.Commands) (*grpcapi.Server, *grpcapi.Client) {
	t.Helper()
	s, err := grpcapi.Start("127.0.0.1:0", "test-token", commands)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = s.Close(ctx)
	})
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	c, err := grpcapi.Dial(ctx, s.Ready().Address, s.Ready().Token)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	return s, c
}

func TestRegisteredClientsAndPassiveProbe(t *testing.T) {
	s, first := start(t, application.NewService())
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	defer cancel()
	second, err := grpcapi.DialKind(ctx, s.Ready().Address, s.Ready().Token, "mcp")
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	info, err := grpcapi.Inspect(ctx, s.Ready().Address, s.Ready().Token)
	if err != nil || info.GetClientCount() != 2 || info.GetActiveTasks() != 0 {
		t.Fatalf("учёт клиентов: %v, %v", info, err)
	}
	first.Close()
	info, err = second.Info(ctx)
	if err != nil || info.GetClientCount() != 1 {
		t.Fatalf("освобождение аренды: %v, %v", info, err)
	}
}

type delayed struct {
	application.Commands
	started chan struct{}
	finish  chan struct{}
}

func (d delayed) Snapshot(context.Context, string) (domain.Snapshot, error) {
	close(d.started)
	<-d.finish
	return domain.Snapshot{SessionID: "main"}, nil
}

func TestAutomaticServerWaitsForActiveRPC(t *testing.T) {
	d := delayed{Commands: application.NewService(), started: make(chan struct{}), finish: make(chan struct{})}
	s, err := grpcapi.StartWithConfig("127.0.0.1:0", "token", d, lifecycle.Config{Mode: lifecycle.Automatic, IdleTimeout: 100 * time.Millisecond, LeaseTimeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = s.Close(ctx)
	}()
	ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
	defer cancel()
	c, err := grpcapi.Dial(ctx, s.Ready().Address, s.Ready().Token)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	result := make(chan error, 1)
	go func() { _, err := c.Snapshot(ctx, "main"); result <- err }()
	select {
	case <-d.started:
	case <-ctx.Done():
		t.Fatal("задача не началась")
	}
	c.Close()
	select {
	case <-s.Done():
		t.Fatal("сервер завершился до задачи")
	case <-time.After(200 * time.Millisecond):
	}
	stats := s.Lifecycle().Stats()
	if stats.Clients != 0 || stats.Tasks != 1 {
		t.Fatalf("учёт задачи: %+v", stats)
	}
	close(d.finish)
	select {
	case <-s.Done():
	case <-ctx.Done():
		t.Fatal("сервер не завершился после задачи")
	}
	<-result
}

func TestAuthorizationAndErrors(t *testing.T) {
	s, c := start(t, application.NewService())
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	_, err := grpcapi.Dial(ctx, s.Ready().Address, "wrong-token")
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("неверный токен: %v", err)
	}
	_, err = c.Snapshot(ctx, "missing")
	if status.Code(err) != codes.NotFound {
		t.Fatalf("отсутствующая сессия: %v", err)
	}
	if _, err = c.Create(ctx, application.CreateSession{SessionID: "main", BuildName: "Имя", RequestID: "create"}); err != nil {
		t.Fatal(err)
	}
	_, err = c.Rename(ctx, application.RenameBuild{SessionID: "main", BuildName: "Имя", ExpectedRevision: 0, RequestID: "rename"})
	if status.Code(err) != codes.Aborted {
		t.Fatalf("конфликт ревизии: %v", err)
	}
	_, err = grpcapi.Start("0.0.0.0:0", "token", application.NewService())
	if err == nil {
		t.Fatal("разрешён внешний интерфейс")
	}
}

type incompatible struct {
	pb.UnimplementedSessionServiceServer
}

func (incompatible) GetServerInfo(context.Context, *pb.GetServerInfoRequest) (*pb.GetServerInfoResponse, error) {
	return &pb.GetServerInfoResponse{ProtocolVersion: "pob2.v999"}, nil
}

func TestRejectIncompatibleProtocol(t *testing.T) {
	l, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	s := grpc.NewServer()
	pb.RegisterSessionServiceServer(s, incompatible{})
	go func() { _ = s.Serve(l) }()
	defer s.Stop()
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	if client, err := grpcapi.Dial(ctx, l.Addr().String(), "token"); err == nil {
		client.Close()
		t.Fatal("несовместимый сервер принят")
	}
}

type blocked struct {
	application.Commands
	started chan struct{}
}

func (b blocked) Snapshot(ctx context.Context, _ string) (domain.Snapshot, error) {
	close(b.started)
	<-ctx.Done()
	return domain.Snapshot{}, ctx.Err()
}

func TestShutdownDeadlineCancelsActiveRPC(t *testing.T) {
	b := blocked{Commands: application.NewService(), started: make(chan struct{})}
	s, c := start(t, b)
	result := make(chan error, 1)
	go func() { _, err := c.Snapshot(t.Context(), "main"); result <- err }()
	select {
	case <-b.started:
	case <-time.After(time.Second):
		t.Fatal("RPC не запустился")
	}
	ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
	defer cancel()
	if err := s.Close(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("ожидался таймаут: %v", err)
	}
	select {
	case err := <-result:
		if err == nil {
			t.Fatal("прерванный RPC успешен")
		}
	case <-time.After(time.Second):
		t.Fatal("RPC не отменён")
	}
}

func TestShutdownDeadlineDoesNotWaitForUncooperativeHandler(t *testing.T) {
	d := delayed{Commands: application.NewService(), started: make(chan struct{}), finish: make(chan struct{})}
	s, c := start(t, d)
	defer close(d.finish)
	result := make(chan error, 1)
	go func() { _, err := c.Snapshot(t.Context(), "main"); result <- err }()
	select {
	case <-d.started:
	case <-time.After(time.Second):
		t.Fatal("RPC не начался")
	}
	ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- s.Close(ctx) }()
	select {
	case err := <-done:
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("ошибка остановки: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("остановка зависла на обработчике")
	}
	select {
	case err := <-result:
		if err == nil {
			t.Fatal("прерванный RPC вернул успех")
		}
	case <-time.After(time.Second):
		t.Fatal("клиент не получил прерывание")
	}
}
