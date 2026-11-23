package integration_test

import (
	"fmt"
	"os"
	"reflect"
	"slices"
	"testing"

	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/build"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestLuaTreeGeometryAcrossProcesses(t *testing.T) {
	if os.Getenv("POB2_ENGINE_POB_DIR") == "" {
		t.Skip("для проверки Lua укажите POB2_ENGINE_POB_DIR")
	}
	t.Setenv("POB2_ENGINE_CALCULATOR", "lua")
	s := newBuildScenario(t, "tree-specs/warrior-witch.xml")
	before := supportSnapshot(t, s)
	tree := compareGeometryTransports(t, s)
	checkGeometry(t, tree)
	if _, err := s.client.TreeGeometry(s.ctx, "main", s.revision-1); status.Code(err) != codes.Aborted {
		t.Fatal("принята старая ревизия дерева", err)
	}
	if !reflect.DeepEqual(supportSnapshot(t, s), before) {
		t.Fatal("чтение геометрии изменило состояние сборки")
	}
}

func compareGeometryTransports(t *testing.T, s *groupScenario) application.TreeGeometryView {
	t.Helper()
	tree, err := s.client.TreeGeometry(s.ctx, "main", s.revision)
	if err != nil {
		t.Fatal(err)
	}
	args := []string{"tree-geometry", "--session", "main", "--revision", fmt.Sprint(s.revision)}
	fromCLI := decodeCatalogJSON[application.TreeGeometryView](t, cli(t, s.ctx, 0, append(args, s.connection...)...))
	answer := call(t, s.ctx, s.mcp, "get_tree_geometry", map[string]any{"expected_revision": s.revision}, false)
	fromMCP := structuredCatalogJSON[application.TreeGeometryView](t, answer.StructuredContent)
	if !reflect.DeepEqual(tree, fromCLI) || !reflect.DeepEqual(tree, fromMCP) {
		t.Fatal("CLI, MCP и gRPC вернули разные координаты или связи")
	}
	return tree
}

func checkGeometry(t *testing.T, tree application.TreeGeometryView) {
	t.Helper()
	if tree.Version != "0_5" || len(tree.Nodes) < 1000 || len(tree.Edges) < len(tree.Nodes) {
		t.Fatal("неполный граф дерева Lua", tree.Version, len(tree.Nodes), len(tree.Edges))
	}
	if !slices.ContainsFunc(tree.Nodes, func(node build.TreeGeometryNode) bool {
		return node.ID == 13397 && node.Attribute
	}) {
		t.Fatal("пропал узел с выбором атрибута")
	}
	if !slices.ContainsFunc(tree.Nodes, func(node build.TreeGeometryNode) bool {
		return node.ID == 38646 && node.Allocated
	}) {
		t.Fatal("пропал назначенный узел активной спецификации")
	}
}
