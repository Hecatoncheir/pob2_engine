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
	withScene := compareGeometryTransports(t, s, true)
	checkTreeScene(t, withScene.Scene)
	if _, err := s.client.TreeGeometry(s.ctx, "main", s.revision-1, false); status.Code(err) != codes.Aborted {
		t.Fatal("принята старая ревизия дерева", err)
	}
	if !reflect.DeepEqual(supportSnapshot(t, s), before) {
		t.Fatal("чтение геометрии изменило состояние сборки")
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
