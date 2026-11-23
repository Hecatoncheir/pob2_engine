package integration_test

import (
	"fmt"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/build"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestLuaTreeNodeTooltipAcrossProcesses(t *testing.T) {
	if os.Getenv("POB2_ENGINE_POB_DIR") == "" {
		t.Skip("для проверки Lua укажите POB2_ENGINE_POB_DIR")
	}
	t.Setenv("POB2_ENGINE_CALCULATOR", "lua")
	s := newBuildScenario(t, "tree-specs/warrior-witch.xml")
	before := supportSnapshot(t, s)
	const node = 13397
	tip := compareTooltipTransports(t, s, node)
	if tip.NodeID != node || tip.Revision != s.revision || !hasTooltipText(tip, "+5 to any Attribute") {
		t.Fatalf("Lua не вернул штатную подсказку атрибута: %+v", tip)
	}
	if tip.Points != 2 || !slices.Equal(tip.Path, []uint32{node, 3936}) || !hasTooltipText(tip, "Full DPS (+12.5%)") {
		t.Fatalf("Lua не вернул ожидаемые путь и влияние узла: %+v", tip)
	}
	if tip.Scene == nil || tip.Scene.AssetID == "" || len(tip.Scene.Commands) == 0 || len(tip.Scene.Commands) > 100 {
		t.Fatal("Lua не вернул отдельную подсветку пути")
	}
	plain, err := s.client.TreeNodeTooltip(s.ctx, "main", s.revision, node, false, false)
	if err != nil || plain.Scene != nil || hasTooltipText(plain, "Full DPS") || !slices.Equal(tip.Path, plain.Path) {
		t.Fatal("отключение сравнения изменило путь или оставило показатели", err)
	}
	if _, err := s.client.TreeNodeTooltip(s.ctx, "main", s.revision-1, node, true, true); status.Code(err) != codes.Aborted {
		t.Fatal("принята старая ревизия дерева", err)
	}
	if _, err := s.client.TreeNodeTooltip(s.ctx, "main", s.revision, 9999999, false, true); status.Code(err) != codes.InvalidArgument {
		t.Fatal("принят неизвестный узел", err)
	}
	if !reflect.DeepEqual(supportSnapshot(t, s), before) {
		t.Fatal("чтение подсказки изменило состояние сборки")
	}
}

func compareTooltipTransports(t *testing.T, s *groupScenario, node uint32) application.TreeNodeTooltipView {
	t.Helper()
	tip, err := s.client.TreeNodeTooltip(s.ctx, "main", s.revision, node, true, true)
	if err != nil {
		t.Fatal(err)
	}
	args := []string{"tree-node-tooltip", "--session", "main", "--revision", fmt.Sprint(s.revision), "--node", fmt.Sprint(node), "--compare", "--scene"}
	fromCLI := decodeCatalogJSON[application.TreeNodeTooltipView](t, cli(t, s.ctx, 0, append(args, s.connection...)...))
	answer := call(t, s.ctx, s.mcp, "get_tree_node_tooltip", map[string]any{
		"expected_revision": s.revision, "node_id": node, "compare": true, "include_scene": true,
	}, false)
	fromMCP := structuredCatalogJSON[application.TreeNodeTooltipView](t, answer.StructuredContent)
	if !reflect.DeepEqual(tip, fromCLI) || !reflect.DeepEqual(tip, fromMCP) {
		t.Fatal("CLI, MCP и gRPC вернули разные подсказки или пути")
	}
	return tip
}

func hasTooltipText(tip application.TreeNodeTooltipView, expected string) bool {
	return slices.ContainsFunc(tip.Lines, func(line build.TreeTooltipLine) bool {
		return strings.Contains(line.Text, expected)
	})
}
