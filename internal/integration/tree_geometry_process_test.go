package integration_test

import (
	"fmt"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

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
	s := newBuildScenarioWithTimeout(t, "tree-specs/warrior-witch.xml", 90*time.Second)
	before := supportSnapshot(t, s)
	tree := compareGeometryTransports(t, s, false)
	checkGeometry(t, tree)
	checkSceneLinks(t, tree.Edges)
	withScene := compareGeometryTransports(t, s, true)
	checkTreeScene(t, withScene.Scene)
	if _, err := s.client.TreeGeometry(s.ctx, "main", s.revision-1, false); status.Code(err) != codes.Aborted {
		t.Fatal("принята старая ревизия дерева", err)
	}
	if !reflect.DeepEqual(supportSnapshot(t, s), before) {
		t.Fatal("чтение геометрии изменило состояние сборки")
	}
}

func checkSceneLinks(t *testing.T, edges []build.TreeGeometryEdge) {
	t.Helper()
	for _, pair := range [][2]int{{3936, 57253}, {38646, 57253}, {4739, 12795}, {12795, 44871}} {
		if !slices.Contains(edges, build.TreeGeometryEdge{FromNode: pair[0], ToNode: pair[1], HiddenInScene: true}) {
			t.Fatalf("служебная связь %v потеряна или попала в отрисовку", pair)
		}
	}
	if !slices.ContainsFunc(edges, func(edge build.TreeGeometryEdge) bool {
		return !edge.HiddenInScene && (edge.FromNode == 38646 || edge.ToNode == 38646)
	}) {
		t.Fatal("пропали обычные линии у стартовых пассивов")
	}
}

func compareGeometryTransports(t *testing.T, s *groupScenario, scene bool) application.TreeGeometryView {
	t.Helper()
	tree, err := s.client.TreeGeometry(s.ctx, "main", s.revision, scene)
	if err != nil {
		t.Fatal(err)
	}
	args := []string{"tree-geometry", "--session", "main", "--revision", fmt.Sprint(s.revision)}
	if scene {
		args = append(args, "--scene")
	}
	fromCLI := decodeCatalogJSON[application.TreeGeometryView](t, cli(t, s.ctx, 0, append(args, s.connection...)...))
	answer := call(t, s.ctx, s.mcp, "get_tree_geometry", map[string]any{"expected_revision": s.revision, "include_scene": scene}, false)
	fromMCP := structuredCatalogJSON[application.TreeGeometryView](t, answer.StructuredContent)
	if !reflect.DeepEqual(tree, fromCLI) || !reflect.DeepEqual(tree, fromMCP) {
		t.Fatal("CLI, MCP и gRPC вернули разные координаты или связи")
	}
	return tree
}

func checkTreeScene(t *testing.T, scene *build.TreeScene) {
	t.Helper()
	if scene == nil || scene.Width != 1800 || scene.Height != 1800 || scene.Scale <= 0 || len(scene.Commands) < 1000 {
		t.Fatal("неполная сцена оригинального дерева")
	}
	if !slices.ContainsFunc(scene.Commands, func(command build.TreeDrawCommand) bool {
		return strings.HasSuffix(command.File, ".dds.zst") && command.Index > 0 &&
			len(command.Positions) == 8 && len(command.UV) == 8 && len(command.Color) == 4
	}) {
		t.Fatal("пропали команды для текстур оригинального формата")
	}
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
