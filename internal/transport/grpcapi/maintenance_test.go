package grpcapi

import (
	"context"
	"net"
	"testing"

	pb "github.com/Hecatoncheir/pob2_engine/gen/go/pob2/v1"
	"github.com/Hecatoncheir/pob2_engine/internal/lifecycle"
	"github.com/Hecatoncheir/pob2_engine/internal/maintenance"
	"github.com/Hecatoncheir/pob2_engine/internal/pob"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

type testMaintenance struct{}

func (testMaintenance) Update(_ context.Context, apply bool, targets ...string) (maintenance.UpdateResult, error) {
	result := maintenance.UpdateResult{Status: "available", LatestVersion: "v1.2.3"}
	if targets[0] != "engine" {
		result.PoB = &pob.State{Status: "available", InstalledVersion: "v0.23.1", LatestVersion: "v0.23.2"}
	}
	if targets[0] == "" || targets[0] == "all" {
		result.Engine = &pob.State{Status: "available", InstalledVersion: "v1.2.2", LatestVersion: "v1.2.3"}
	}
	if apply {
		result.Status, result.RestartRequired = "updated", true
	}
	return result, nil
}

func (testMaintenance) Doctor(context.Context) (maintenance.Report, error) {
	return maintenance.Report{Status: "ok", DataID: "test", PoB: &pob.State{InstalledVersion: "v0.23.1"}}, nil
}

func TestMaintenanceRPC(t *testing.T) {
	client := startMaintenanceRPC(t)
	ctx := metadata.AppendToOutgoingContext(t.Context(), "authorization", "Bearer test-maintenance")
	report, err := client.Doctor(ctx)
	if err != nil || report.DataID != "test" || report.Server == nil || report.Server.State != "ready" ||
		report.PoB == nil || report.PoB.InstalledVersion != "v0.23.1" {
		t.Fatalf("Doctor: %+v, %v", report, err)
	}
	check, err := client.Update(ctx, false, "pob")
	if err != nil || check.Status != "available" || check.RestartRequired || check.PoB == nil || check.PoB.LatestVersion != "v0.23.2" {
		t.Fatalf("Update check: %+v, %v", check, err)
	}
	applied, err := client.Update(ctx, true)
	if err != nil || applied.Status != "updated" || !applied.RestartRequired {
		t.Fatalf("Update apply: %+v, %v", applied, err)
	}
	if _, err := client.Update(t.Context(), true); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("Update без токена: %v", err)
	}
}

func TestMaintenanceRPCTargets(t *testing.T) {
	client := startMaintenanceRPC(t)
	ctx := metadata.AppendToOutgoingContext(t.Context(), "authorization", "Bearer test-maintenance")
	engine, err := client.Update(ctx, false, "engine")
	if err != nil || engine.LatestVersion != "v1.2.3" || engine.PoB != nil {
		t.Fatalf("неверный ответ update для engine: %+v, %v", engine, err)
	}
	for _, targets := range [][]string{nil, {"all"}} {
		result, err := client.Update(ctx, false, targets...)
		if err != nil || result.Engine == nil || result.PoB == nil {
			t.Fatalf("неполный ответ update для %v: %+v, %v", targets, result, err)
		}
	}
}

func startMaintenanceRPC(t *testing.T) *Client {
	t.Helper()
	life, err := lifecycle.NewStarting(lifecycle.Config{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(life.Close)
	if err := life.MarkReady(); err != nil {
		t.Fatal(err)
	}
	server := &Server{maintenance: testMaintenance{}, life: life, token: "test-maintenance"}
	return bufferedSessionClient(t, server)
}

func bufferedSessionClient(t *testing.T, server *Server) *Client {
	t.Helper()
	rpc := grpc.NewServer(grpc.UnaryInterceptor(server.authorize), grpc.StreamInterceptor(server.authorizeStream))
	pb.RegisterSessionServiceServer(rpc, server)
	listener := bufconn.Listen(1 << 20)
	go func() { _ = rpc.Serve(listener) }()
	t.Cleanup(rpc.Stop)
	conn, err := grpc.NewClient("passthrough:///maintenance", grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return listener.DialContext(ctx)
		}))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return &Client{rpc: pb.NewSessionServiceClient(conn)}
}
