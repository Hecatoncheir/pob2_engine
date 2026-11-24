package integration_test

import (
	"strings"
	"testing"

	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/maintenance"
	"github.com/Hecatoncheir/pob2_engine/internal/transport/grpcapi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestStandaloneMaintenance(t *testing.T) {
	offlineMaintenance(t)
	ctx := t.Context()
	report := decodeCatalogJSON[maintenance.Report](t, cli(t, ctx, 0, "doctor"))
	checkDoctor(t, report)
	if report.Server != nil {
		t.Fatalf("в локальной диагностике появился сервер: %+v", report.Server)
	}
	licenses := decodeCatalogJSON[map[string]string](t, cli(t, ctx, 0, "doctor", "--licenses"))
	if !strings.Contains(licenses["LICENSE"], "MIT License") || licenses[".licenses/runtime/manifest.json"] == "" {
		t.Fatal("неполные встроенные уведомления")
	}
	cli(t, ctx, 1, "update", "--target", "engine")
	mcp := connectMCP(t, ctx, "--standalone", "--session", "maintenance")
	fromMCP := call(t, ctx, mcp, "doctor", map[string]any{}, false)
	checkDoctor(t, structuredCatalogJSON[maintenance.Report](t, fromMCP.StructuredContent))
	call(t, ctx, mcp, "update", map[string]any{"apply": true, "target": "engine"}, true)
	checkCLIUpdatesOffline(t)
	for _, target := range []string{"engine", "pob"} {
		call(t, ctx, mcp, "update", map[string]any{"target": target}, true)
	}
	for _, args := range []map[string]any{{}, {"target": "all"}} {
		result := call(t, ctx, mcp, "update", args, false)
		checkBothUpdatesOffline(t, structuredCatalogJSON[maintenance.UpdateResult](t, result.StructuredContent))
	}
}

func TestServerMaintenanceAcrossAdapters(t *testing.T) {
	offlineMaintenance(t)
	ctx := t.Context()
	server := startServer(t, ctx, "--ephemeral")
	client, err := grpcapi.Dial(ctx, server.ready.Address, server.ready.Token)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	if _, err := client.Create(ctx, application.CreateSession{SessionID: "main", BuildName: "Сборка", RequestID: "create"}); err != nil {
		t.Fatal(err)
	}
	report, err := client.Doctor(ctx)
	if err != nil {
		t.Fatal(err)
	}
	checkServerDoctor(t, report)
	if _, err := client.Update(ctx, true, "engine"); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("gRPC принял обновление сборки без тега: %v", err)
	}
	checkGRPCUpdatesOffline(t, client)
	connect := []string{"--address", server.ready.Address, "--token", server.ready.Token}
	checkServerDoctor(t, decodeCatalogJSON[maintenance.Report](t, cli(t, ctx, 0, append([]string{"doctor"}, connect...)...)))
	cli(t, ctx, 1, append([]string{"update", "--target", "engine"}, connect...)...)
	mcp := connectMCP(t, ctx, append([]string{"--attach", "main"}, connect...)...)
	result := call(t, ctx, mcp, "doctor", map[string]any{}, false)
	checkServerDoctor(t, structuredCatalogJSON[maintenance.Report](t, result.StructuredContent))
	call(t, ctx, mcp, "update", map[string]any{"apply": true, "target": "engine"}, true)
}

func checkDoctor(t *testing.T, report maintenance.Report) {
	t.Helper()
	if report.Status != "ok" || report.Version != "integration-test" || report.ProtocolVersion != "pob2.v1" ||
		report.Executable == "" || report.DataID == "" || len(report.DataSHA256) != 64 || len(report.Issues) != 0 {
		t.Fatalf("неполная диагностика: %+v", report)
	}
	if report.EngineUpdate == nil || report.EngineUpdate.InstalledVersion != report.Version ||
		report.EngineUpdate.Status != "unavailable" || report.PoB == nil || report.PoB.Status != "unavailable" {
		t.Fatalf("неполная диагностика версий без сети: engine=%+v, pob=%+v", report.EngineUpdate, report.PoB)
	}
}

func checkServerDoctor(t *testing.T, report maintenance.Report) {
	t.Helper()
	checkDoctor(t, report)
	if report.Server == nil || report.Server.State != "ready" || report.Server.StorageMode != "memory" {
		t.Fatalf("неполная диагностика сервера: %+v", report.Server)
	}
}

func offlineMaintenance(t *testing.T) {
	t.Helper()
	base := t.TempDir()
	t.Setenv("HOME", base)
	t.Setenv("XDG_CACHE_HOME", base)
	t.Setenv("LOCALAPPDATA", base)
	t.Setenv("HTTPS_PROXY", "http://127.0.0.1:0")
	t.Setenv("NO_PROXY", "")
	t.Setenv("no_proxy", "")
}

func checkCLIUpdatesOffline(t *testing.T) {
	t.Helper()
	ctx := t.Context()
	for _, target := range []string{"engine", "pob"} {
		cli(t, ctx, 1, "update", "--check", "--target", target)
	}
	for _, args := range [][]string{{"update", "--check"}, {"update", "--check", "--target", "all"}} {
		checkBothUpdatesOffline(t, decodeCatalogJSON[maintenance.UpdateResult](t, cli(t, ctx, 0, args...)))
	}
}

func checkGRPCUpdatesOffline(t *testing.T, client *grpcapi.Client) {
	t.Helper()
	ctx := t.Context()
	for _, target := range []string{"engine", "pob"} {
		if _, err := client.Update(ctx, false, target); status.Code(err) != codes.FailedPrecondition {
			t.Fatalf("gRPC вернул неверный статус %s: %v", target, err)
		}
	}
	for _, targets := range [][]string{nil, {"all"}} {
		result, err := client.Update(ctx, false, targets...)
		if err != nil {
			t.Fatal(err)
		}
		checkBothUpdatesOffline(t, result)
	}
}

func checkBothUpdatesOffline(t *testing.T, result maintenance.UpdateResult) {
	t.Helper()
	if result.Status != "partial" || result.RestartRequired || result.Engine == nil || result.PoB == nil {
		t.Fatalf("нет статуса обеих целей обновления: %+v", result)
	}
	if result.Engine.Status != "unavailable" || result.PoB.Status != "unavailable" ||
		result.Engine.Message == "" || result.PoB.Message == "" {
		t.Fatalf("не показаны причины отказа: engine=%+v, pob=%+v", result.Engine, result.PoB)
	}
}
