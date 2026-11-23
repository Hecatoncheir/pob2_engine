package integration_test

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/build"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestLuaTreePowerAcrossProcesses(t *testing.T) {
	if os.Getenv("POB2_ENGINE_POB_DIR") == "" {
		t.Skip("для проверки Lua укажите POB2_ENGINE_POB_DIR")
	}
	t.Setenv("POB2_ENGINE_CALCULATOR", "lua")
	s := newBuildScenarioWithTimeout(t, "tree-specs/warrior-witch.xml", 90*time.Second)
	before := supportSnapshot(t, s)
	catalog, err := s.client.TreePower(s.ctx, "main", s.revision, build.TreePowerQuery{})
	if err != nil || !slices.Contains(catalog.Metrics, build.TreePowerMetric{ID: "Life", Label: "Life"}) {
		t.Fatal("в каталоге PoB нет здоровья", err)
	}
	query := build.TreePowerQuery{Metric: "Life", MaxDistance: 8, Allocation: "unallocated"}
	view := comparePowerTransports(t, s, query)
	if view.Metric == nil || view.Metric.ID != query.Metric || len(view.Entries) == 0 {
		t.Fatal("Lua не вернул влияние узлов на здоровье")
	}
	for _, entry := range view.Entries {
		if entry.Allocated || entry.Distance > query.MaxDistance || entry.Power == 0 {
			t.Fatalf("в отчёт попал узел вне фильтра: %+v", entry)
		}
	}
	checkTreePowerCancellation(t, s)
	if !reflect.DeepEqual(supportSnapshot(t, s), before) {
		t.Fatal("расчёт влияния изменил состояние сборки")
	}
}

func comparePowerTransports(t *testing.T, s *groupScenario, query build.TreePowerQuery) application.TreePowerView {
	t.Helper()
	view, err := s.client.TreePower(s.ctx, "main", s.revision, query)
	if err != nil {
		t.Fatal(err)
	}
	args := []string{"tree-power", "--session", "main", "--revision", fmt.Sprint(s.revision), "--metric", query.Metric,
		"--max-distance", fmt.Sprint(query.MaxDistance), "--allocation", query.Allocation}
	fromCLI := decodeCatalogJSON[application.TreePowerView](t, cli(t, s.ctx, 0, append(args, s.connection...)...))
	answer := call(t, s.ctx, s.mcp, "get_tree_power", map[string]any{
		"expected_revision": s.revision, "metric": query.Metric, "max_distance": query.MaxDistance, "allocation": query.Allocation,
	}, false)
	fromMCP := structuredCatalogJSON[application.TreePowerView](t, answer.StructuredContent)
	if !reflect.DeepEqual(view, fromCLI) || !reflect.DeepEqual(view, fromMCP) {
		t.Fatal("CLI, MCP и gRPC вернули разное влияние узлов")
	}
	return view
}

func checkTreePowerCancellation(t *testing.T, s *groupScenario) {
	t.Helper()
	ctx, cancel := context.WithTimeout(s.ctx, 100*time.Millisecond)
	defer cancel()
	query := build.TreePowerQuery{Metric: "FullDPS"}
	if _, err := s.client.TreePower(ctx, "main", s.revision, query); status.Code(err) != codes.DeadlineExceeded {
		t.Fatal("отмена расчёта влияния не дошла до клиента", err)
	}
	if _, err := s.client.BuildReport(s.ctx, "main", s.revision); err != nil {
		t.Fatal("отмена помешала чтению отчёта общей сборки", err)
	}
}
