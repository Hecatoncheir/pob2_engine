package grpcapi

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/build"
	"github.com/Hecatoncheir/pob2_engine/internal/calculation"
	"github.com/Hecatoncheir/pob2_engine/internal/lifecycle"
	"github.com/Hecatoncheir/pob2_engine/internal/pob"
	"github.com/Hecatoncheir/pob2_engine/internal/transport/mcpapi"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/grpc/metadata"
)

func TestMain(m *testing.M) {
	if len(os.Args) > 1 && os.Args[1] == "_pob-lua" {
		if err := pob.RunWorker(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func TestLocalLuaJITCalculationAcrossGRPCAndMCP(t *testing.T) {
	service, grpc, ctx := luaAcceptanceService(t)
	for _, spell := range []struct {
		name, fixture string
		dps           float64
	}{
		{"Fireball", "fireball.xml", 8.9166666666667},
		{"Frostbolt", "0.0.2/cases/frostbolt-minimal.xml", 9.1},
		{"Spark", "0.0.2/cases/spark-minimal.xml", 8.5642857142857},
	} {
		t.Run(spell.name, func(t *testing.T) {
			snapshot := calculateThroughAdapters(t, ctx, grpc, service, spell.fixture)
			report := snapshot.Build.Calculation
			if report.DataID != calculation.LocalPoBDataID || report.Status != "complete" ||
				report.Skill != spell.name || math.Abs(report.Stats["TotalDPS"]-2*spell.dps) > 1e-8 {
				t.Fatalf("неожиданный расчёт: %s %s %s DPS %.12g", report.DataID, report.Status, report.Skill, report.Stats["TotalDPS"])
			}
		})
	}
	t.Run("Выбор наборов", func(t *testing.T) {
		checkLuaSetSelection(t, ctx, grpc, service)
	})
}

func luaAcceptanceService(t *testing.T) (*application.Service, *Client, context.Context) {
	t.Helper()
	if os.Getenv("POB2_ENGINE_POB_DIR") == "" {
		t.Skip("для проверки LuaJIT укажите POB2_ENGINE_POB_DIR")
	}
	manager, err := pob.Local()
	if err != nil {
		t.Fatal(err)
	}
	installation, err := manager.Installed()
	if err != nil {
		t.Fatal(err)
	}
	engine, err := pob.NewCalculator(installation)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = engine.Close() })
	service := application.NewServiceWithCalculator(engine)
	life, err := lifecycle.NewStarting(lifecycle.Config{LeaseTimeout: time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(life.Close)
	if err := life.MarkReady(); err != nil {
		t.Fatal(err)
	}
	lease, err := life.Register("lua-acceptance", "cli")
	if err != nil {
		t.Fatal(err)
	}
	client := bufferedSessionClient(t, &Server{commands: service, life: life, token: "lua-acceptance"})
	ctx, cancel := context.WithTimeout(t.Context(), 45*time.Second)
	t.Cleanup(cancel)
	return service, client, metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer lua-acceptance", "pob2-lease", lease.ID)
}

func calculateThroughAdapters(t *testing.T, ctx context.Context, grpc *Client, service application.Commands, fixture string) application.SnapshotView {
	t.Helper()
	id := filepath.Base(fixture)
	client := openLuaAcceptanceSession(t, ctx, grpc, service, id, fixture)
	edit := json.RawMessage(`{"expected_revision":2,"request_id":"edit","edit":{"config":{"name":"customMods","value":{"type":"string","value":"100% increased Spell Damage"}}}}`)
	callLuaTool(t, ctx, client, "edit_build", edit)
	request := application.CalculateBuild{SessionID: id, ExpectedRevision: 3, RequestID: "calculate"}
	calculated := callLuaTool(t, ctx, client, "calculate_build", map[string]any{"expected_revision": 3, "request_id": request.RequestID})
	replay, err := grpc.Calculate(ctx, request)
	if err != nil {
		t.Fatal(err)
	}
	view, err := application.ViewSnapshot(replay)
	if err != nil || !reflect.DeepEqual(view, calculated) {
		t.Fatalf("gRPC не вернул снимок, сохранённый MCP: %v", err)
	}
	checkLuaPreview(t, ctx, grpc, calculated)
	return calculated
}

func openLuaAcceptanceSession(t *testing.T, ctx context.Context, grpc *Client, service application.Commands, id, fixture string) *mcp.ClientSession {
	t.Helper()
	if _, err := grpc.Create(ctx, application.CreateSession{SessionID: id, BuildName: id, RequestID: "create-" + id}); err != nil {
		t.Fatal(err)
	}
	xml, err := os.ReadFile(filepath.Join("..", "..", "..", "testdata", "reference", fixture))
	if err != nil {
		t.Fatal(err)
	}
	client := connectLocalMCP(t, ctx, service, id)
	callLuaTool(t, ctx, client, "import_build", map[string]any{"expected_revision": 1, "request_id": "import", "xml": string(xml)})
	return client
}

func checkLuaSetSelection(t *testing.T, ctx context.Context, grpc *Client, service application.Commands) {
	t.Helper()
	const id = "druid-sets"
	client := openLuaAcceptanceSession(t, ctx, grpc, service, id, "0.0.2/groups/cases/druid-group-1.xml")
	edit := &build.Edit{Selection: &build.SelectionEdit{MainGroup: 1, Sets: &build.ActiveSets{Skills: 2, Config: 2}}}
	if _, err := grpc.Apply(ctx, application.ApplyBuild{SessionID: id, ExpectedRevision: 2, RequestID: "sets", Edit: edit}); err != nil {
		t.Fatal(err)
	}
	snapshot := callLuaTool(t, ctx, client, "calculate_build", map[string]any{"expected_revision": 3, "request_id": "calculate"})
	active, report := snapshot.Build.Model.ActiveSets, snapshot.Build.Calculation
	if active.Skills != 2 || active.Config != 2 || report.Skill != "Spark" || math.Abs(report.Stats["TotalDPS"]-17.1285714285714) > 1e-8 {
		t.Fatalf("неожиданный результат для второго набора: %+v, %s, DPS %.12g", active, report.Skill, report.Stats["TotalDPS"])
	}
}

func connectLocalMCP(t *testing.T, ctx context.Context, commands application.Commands, id string) *mcp.ClientSession {
	t.Helper()
	clientTransport, serverTransport := mcp.NewInMemoryTransports()
	server, err := mcpapi.New(commands, id, 20*time.Second).Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = server.Close() })
	client := mcp.NewClient(&mcp.Implementation{Name: "lua-acceptance", Version: "1"}, nil)
	session, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = session.Close() })
	return session
}

func callLuaTool(t *testing.T, ctx context.Context, client *mcp.ClientSession, name string, args any) application.SnapshotView {
	t.Helper()
	response, err := client.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil || response.IsError {
		t.Fatalf("MCP %s: %+v, %v", name, response, err)
	}
	data, err := json.Marshal(response.StructuredContent)
	if err != nil {
		t.Fatal(err)
	}
	var result application.SnapshotView
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}
	return result
}

func checkLuaPreview(t *testing.T, ctx context.Context, grpc *Client, saved application.SnapshotView) {
	t.Helper()
	preview, err := grpc.PreviewSupport(ctx, application.PreviewSupport{SessionID: saved.SessionID,
		ExpectedRevision: saved.Revision, SupportID: "Rapid Casting I"})
	if err != nil || !preview.Compatible || preview.Status != "complete" {
		t.Fatalf("предпросмотр LuaJIT: %+v, %v", preview, err)
	}
	before, after := preview.Comparison.Before, preview.Comparison.After
	if !reflect.DeepEqual(before.Stats, saved.Build.Calculation.Stats) || before.InputSHA256 != saved.Build.InputSHA256 ||
		after.InputSHA256 == before.InputSHA256 {
		t.Fatal("предпросмотр не привязан к сохранённому входу и новому варианту")
	}
	if after.Stats["Speed"] <= before.Stats["Speed"] {
		t.Fatal("Rapid Casting I не увеличил скорость")
	}
	current, err := grpc.Snapshot(ctx, saved.SessionID)
	if err != nil {
		t.Fatal(err)
	}
	view, err := application.ViewSnapshot(current)
	if err != nil || !reflect.DeepEqual(view, saved) {
		t.Fatalf("предпросмотр изменил сохранённый снимок: %v", err)
	}
}
