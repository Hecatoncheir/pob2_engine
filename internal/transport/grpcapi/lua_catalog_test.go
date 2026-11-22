package grpcapi

import (
	"context"
	"math"
	"reflect"
	"testing"

	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/build"
	"github.com/Hecatoncheir/pob2_engine/internal/calculation"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestLuaFullCatalogAcrossGRPCAndMCP(t *testing.T) {
	service, grpc, ctx := luaAcceptanceService(t)
	const id = "full-catalog"
	client := openLuaAcceptanceSession(t, ctx, grpc, service, id, "fireball.xml")
	checkFullLuaCatalogPages(t, ctx, grpc, client)
	edit := &build.Edit{Gems: &build.GemsEdit{GroupID: 1, Enabled: true,
		Gems: []build.GemInput{{ID: "Ice Nova", Level: 1, Enabled: true}}}}
	if _, err := grpc.Apply(ctx, application.ApplyBuild{SessionID: id, ExpectedRevision: 2, RequestID: "new-gem", Edit: edit}); err != nil {
		t.Fatal(err)
	}
	snapshot := callLuaTool(t, ctx, client, "calculate_build", map[string]any{"expected_revision": 3, "request_id": "new-calculation"})
	if result := snapshot.Build.Calculation; result.Status != "complete" || result.Skill != "Ice Nova" || math.Abs(result.Stats["TotalDPS"]-7.84) > 1e-8 {
		t.Fatalf("Ice Nova: %+v", result)
	}
	command := application.PreviewSupport{SessionID: id, ExpectedRevision: snapshot.Revision,
		SupportID: "Concentrated Area", Level: 1, Quality: 20}
	preview, err := grpc.PreviewSupport(ctx, command)
	if err != nil || !preview.Compatible || preview.Status != "complete" || math.Abs(preview.Comparison.After.Stats["TotalDPS"]-10.08) > 1e-8 {
		t.Fatalf("новая поддержка: %+v %v", preview, err)
	}
	fromMCP := luaToolResult[application.SupportPreview](t, ctx, client, "preview_support", map[string]any{
		"expected_revision": command.ExpectedRevision, "support_id": command.SupportID, "level": command.Level, "quality": command.Quality})
	if !reflect.DeepEqual(preview, fromMCP) {
		t.Fatal("разный предпросмотр через gRPC и MCP")
	}
	checkInvalidLuaSupports(t, ctx, grpc, snapshot)
	current, err := grpc.Snapshot(ctx, id)
	view, viewErr := application.ViewSnapshot(current)
	if err != nil || viewErr != nil || !reflect.DeepEqual(view, snapshot) {
		t.Fatal("предпросмотр изменил снимок", err, viewErr)
	}
}

func checkFullLuaCatalogPages(t *testing.T, ctx context.Context, grpc *Client, client *mcp.ClientSession) {
	t.Helper()
	for _, query := range []application.CatalogQuery{
		{Kind: "gems", Limit: 1}, {Kind: "gems", Query: "Ice Nova"},
		{Kind: "items", Query: "Rusted Cuirass"}, {Kind: "passives", Query: "41965"}, {Kind: "classes"},
	} {
		page, err := grpc.Catalog(ctx, query)
		if err != nil {
			t.Fatal(err)
		}
		fromMCP := luaToolResult[application.CatalogPage](t, ctx, client, "get_catalog", query)
		if !reflect.DeepEqual(page, fromMCP) || page.DataID != calculation.LocalPoBDataID || page.Total == 0 {
			t.Fatalf("разный или неверный каталог: %+v", query)
		}
	}
}

func checkInvalidLuaSupports(t *testing.T, ctx context.Context, grpc *Client, snapshot application.SnapshotView) {
	t.Helper()
	command := application.PreviewSupport{SessionID: snapshot.SessionID, ExpectedRevision: snapshot.Revision, SupportID: "Rapid Attacks I"}
	preview, err := grpc.PreviewSupport(ctx, command)
	if err != nil || preview.Compatible || preview.Status != "partial" {
		t.Fatalf("несовместимая поддержка: %+v %v", preview, err)
	}
	command.SupportID, command.Level = "Concentrated Area", 40
	if _, err := grpc.PreviewSupport(ctx, command); status.Code(err) != codes.InvalidArgument {
		t.Fatal("отсутствующий уровень поддержки принят", err)
	}
}
