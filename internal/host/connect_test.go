package host_test

import (
	"context"
	"errors"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	pb "github.com/Hecatoncheir/pob2_engine/gen/go/pob2/v1"
	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/discovery"
	"github.com/Hecatoncheir/pob2_engine/internal/host"
	"github.com/Hecatoncheir/pob2_engine/internal/lifecycle"
	"google.golang.org/grpc"
)

type infoServer struct {
	pb.UnimplementedSessionServiceServer
	version  string
	instance string
}

type changingInstance struct {
	pb.UnimplementedSessionServiceServer
	calls atomic.Int32
}

func (s *changingInstance) GetServerInfo(context.Context, *pb.GetServerInfoRequest) (*pb.GetServerInfoResponse, error) {
	id := "old"
	if s.calls.Add(1) > 1 {
		id = "new"
	}
	return &pb.GetServerInfoResponse{ProtocolVersion: "pob2.v1", InstanceId: id, State: "ready"}, nil
}

func TestDiscoveryRechecksInstanceOnActualConnection(t *testing.T) {
	base := t.TempDir()
	profile, err := discovery.Open(base, "main")
	if err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := grpc.NewServer()
	pb.RegisterSessionServiceServer(server, &changingInstance{})
	go func() { _ = server.Serve(listener) }()
	defer server.Stop()
	if err := profile.Publish(discovery.Record{InstanceID: "old", Address: listener.Addr().String(), Token: "token", ProtocolVersion: "pob2.v1"}); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	client, err := host.Connect(ctx, host.ConnectOptions{BaseDir: base, Profile: "main"})
	if client != nil {
		client.Close()
	}
	if err == nil || !strings.Contains(err.Error(), "идентификатор экземпляра") {
		t.Fatalf("смена экземпляра не обнаружена: %v", err)
	}
}

func TestCanceledConnectDoesNotCreateProfile(t *testing.T) {
	base := t.TempDir()
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err := host.Connect(ctx, host.ConnectOptions{BaseDir: base, Profile: "main", StartServer: true})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("отменённое подключение: %v", err)
	}
	if _, err := os.Stat(filepath.Join(base, "pob2-engine")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("отменённый запрос создал профиль: %v", err)
	}
}

func (s infoServer) GetServerInfo(context.Context, *pb.GetServerInfoRequest) (*pb.GetServerInfoResponse, error) {
	return &pb.GetServerInfoResponse{ProtocolVersion: s.version, InstanceId: s.instance, State: "ready"}, nil
}

func TestDiscoveryRejectsIncompatibleOrWrongInstance(t *testing.T) {
	for _, item := range []struct{ name, version, instance, message string }{
		{"protocol", "pob2.v999", "expected", "несовместимый протокол"},
		{"instance", "pob2.v1", "other", "идентификатор экземпляра"},
	} {
		t.Run(item.name, func(t *testing.T) {
			base := t.TempDir()
			profile, err := discovery.Open(base, "main")
			if err != nil {
				t.Fatal(err)
			}
			lock, err := profile.TryLock("server")
			if err != nil {
				t.Fatal(err)
			}
			defer lock.Close()
			listener, err := net.Listen("tcp4", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			server := grpc.NewServer()
			pb.RegisterSessionServiceServer(server, infoServer{version: item.version, instance: item.instance})
			go func() { _ = server.Serve(listener) }()
			defer server.Stop()
			record := discovery.Record{InstanceID: "expected", Address: listener.Addr().String(), Token: "token", ProtocolVersion: "pob2.v1"}
			if err := profile.Publish(record); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(t.Context(), time.Second)
			defer cancel()
			_, err = host.Connect(ctx, host.ConnectOptions{BaseDir: base, Profile: "main", StartServer: true})
			if err == nil || !strings.Contains(err.Error(), item.message) {
				t.Fatalf("подключение: %v", err)
			}
			if got, err := profile.Read(); err != nil || got != record {
				t.Fatalf("активный профиль перезаписан: %v, %v", got, err)
			}
		})
	}
}

func TestLockedProfileWaitsInsteadOfStartingSecondProcess(t *testing.T) {
	base := t.TempDir()
	profile, err := discovery.Open(base, "main")
	if err != nil {
		t.Fatal(err)
	}
	lock, err := profile.TryLock("server")
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Close()
	ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
	defer cancel()
	_, err = host.Connect(ctx, host.ConnectOptions{BaseDir: base, Profile: "main", StartServer: true, Executable: "несуществующий-движок"})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("блокировка не ограничила запуск: %v", err)
	}
}

func TestServerCleanupPreservesAnotherInstanceRecord(t *testing.T) {
	profile, err := discovery.Open(t.TempDir(), "main")
	if err != nil {
		t.Fatal(err)
	}
	server, err := host.Start(profile, "127.0.0.1:0", "", application.NewService(), lifecycle.Config{})
	if err != nil {
		t.Fatal(err)
	}
	record, err := profile.Read()
	if err != nil {
		t.Fatal(err)
	}
	record.InstanceID = "another-instance"
	if err := profile.Publish(record); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	if err := server.Close(ctx); err != nil {
		t.Fatal(err)
	}
	if got, err := profile.Read(); err != nil || got.InstanceID != "another-instance" {
		t.Fatalf("чужой файл удалён: %v, %v", got, err)
	}
	lock, err := profile.TryLock("server")
	if err != nil {
		t.Fatal(err)
	}
	lock.Close()
	if err := os.Remove(profile.RecordPath()); err != nil {
		t.Fatal(err)
	}
}
