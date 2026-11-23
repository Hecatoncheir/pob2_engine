package integration_test

import (
	"maps"
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

func TestLuaTreeAttributesAcrossProcesses(t *testing.T) {
	if os.Getenv("POB2_ENGINE_POB_DIR") == "" {
		t.Skip("для проверки Lua укажите POB2_ENGINE_POB_DIR")
	}
	t.Setenv("POB2_ENGINE_CALCULATOR", "lua")
	// Один сценарий включает расчёты через три адаптера и перезапуск сервера.
	s := newBuildScenarioWithTimeout(t, "0.0.10/warrior.xml", 90*time.Second)
	checkAttributeCatalog(t, s)
	before := calculateBeforeEquipment(t, s)
	checkAttributeRejections(t, s)
	checkAttributeChoices(t, s)
	if current := supportSnapshot(t, s); !reflect.DeepEqual(current, before) {
		t.Fatal("предпросмотр атрибутов изменил сессию")
	}
	allocated := allocateAttributeNode(t, s)
	command, swapped := switchAttributeNode(t, s, allocated)
	checkBuildEditRestart(t, s, command, swapped, swapped)
}

func attributeEdit(id int, action string, attribute build.PassiveAttribute) *build.Edit {
	return &build.Edit{TreeNode: &build.TreeNodeEdit{NodeID: id, Action: action, Attribute: attribute}}
}

func checkAttributeCatalog(t *testing.T, s *groupScenario) {
	t.Helper()
	info, err := s.client.Info(s.ctx)
	if err != nil || !slices.Contains(info.GetCapabilities(), "build.tree_attributes") {
		t.Fatal("движок не объявил выбор атрибутов", err)
	}
	args := []string{"catalog", "--kind", "passives", "--query", "13397"}
	fromCLI := decodeCatalogJSON[application.CatalogPage](t, cli(t, s.ctx, 0, append(args, s.connection...)...))
	fromGRPC, err := s.client.Catalog(s.ctx, application.CatalogQuery{Kind: "passives", Query: "13397"})
	answer := call(t, s.ctx, s.mcp, "get_catalog", map[string]any{"kind": "passives", "query": "13397"}, false)
	fromMCP := structuredCatalogJSON[application.CatalogPage](t, answer.StructuredContent)
	if err != nil || !reflect.DeepEqual(fromCLI, fromGRPC) || !reflect.DeepEqual(fromCLI, fromMCP) {
		t.Fatal("каталоги CLI, MCP и gRPC различаются", err)
	}
	if len(fromGRPC.Entries) != 1 || !fromGRPC.Entries[0].AttributeChoice {
		t.Fatalf("каталог не указал выбор атрибута: %+v", fromGRPC.Entries)
	}
}

func checkAttributeChoices(t *testing.T, s *groupScenario) {
	t.Helper()
	for _, attribute := range []build.PassiveAttribute{build.Strength, build.Dexterity, build.Intelligence} {
		preview := compareTreeNode(t, s, attributeEdit(13397, "allocate", attribute))
		checkAttributePreview(t, preview, []int{3936, 13397}, []build.PassiveAttributeChange{{NodeID: 13397, After: attribute}}, map[build.PassiveAttribute]float64{attribute: 5})
	}
	path := compareTreeNode(t, s, attributeEdit(1207, "allocate", build.Intelligence))
	checkAttributePreview(t, path, []int{1207, 3936, 13397}, []build.PassiveAttributeChange{
		{NodeID: 1207, After: build.Intelligence}, {NodeID: 13397, After: build.Intelligence},
	}, map[build.PassiveAttribute]float64{build.Intelligence: 10})
}

func checkAttributePreview(t *testing.T, preview application.BuildComparison, added []int, attributes []build.PassiveAttributeChange, want map[build.PassiveAttribute]float64) {
	t.Helper()
	if preview.Passives == nil || !slices.Equal(preview.Passives.Added, added) ||
		len(preview.Passives.Removed) != 0 || !slices.Equal(preview.Passives.Attributes, attributes) {
		t.Fatalf("неверный состав правки: %+v", preview.Passives)
	}
	before, after := preview.Comparison.Before, preview.Comparison.After
	for attribute, stat := range map[build.PassiveAttribute]string{build.Strength: "Str", build.Dexterity: "Dex", build.Intelligence: "Int"} {
		if after.Stats[stat]-before.Stats[stat] != want[attribute] {
			t.Fatalf("неверное изменение %s: %g → %g", stat, before.Stats[stat], after.Stats[stat])
		}
	}
	if !slices.Equal(before.Issues, after.Issues) || after.Status != "complete" {
		t.Fatalf("изменение диагностики PoB: %v → %v", before.Issues, after.Issues)
	}
}

func checkAttributeRejections(t *testing.T, s *groupScenario) {
	t.Helper()
	for _, edit := range []*build.Edit{
		attributeEdit(3936, "allocate", build.Strength),
		attributeEdit(13397, "set_attribute", build.Dexterity),
	} {
		if _, err := s.client.Compare(s.ctx, application.CompareBuild{SessionID: "main", ExpectedRevision: s.revision, Edit: edit}); status.Code(err) != codes.InvalidArgument {
			t.Fatal("принята недопустимая правка", edit.TreeNode, err)
		}
		call(t, s.ctx, s.mcp, "compare_build", map[string]any{"edit": edit, "expected_revision": s.revision}, true)
	}
	bad := attributeEdit(13397, "allocate", "luck")
	call(t, s.ctx, s.mcp, "edit_build", map[string]any{"edit": bad, "expected_revision": s.revision, "request_id": "bad-attribute"}, true)
}

func allocateAttributeNode(t *testing.T, s *groupScenario) application.SnapshotView {
	t.Helper()
	edit := attributeEdit(1207, "allocate", build.Dexterity)
	preview := compareTreeNode(t, s, edit)
	return applyTreeNodeCLI(t, s, application.ApplyBuild{SessionID: "main", ExpectedRevision: s.revision, RequestID: "allocate-dexterity", Edit: edit}, preview)
}

func switchAttributeNode(t *testing.T, s *groupScenario, allocated application.SnapshotView) (application.ApplyBuild, application.SnapshotView) {
	t.Helper()
	edit := attributeEdit(13397, "set_attribute", build.Intelligence)
	preview := compareTreeNode(t, s, edit)
	checkAttributePreview(t, preview, nil, []build.PassiveAttributeChange{{NodeID: 13397, Before: build.Dexterity, After: build.Intelligence}},
		map[build.PassiveAttribute]float64{build.Dexterity: -5, build.Intelligence: 5})
	command := application.ApplyBuild{SessionID: "main", ExpectedRevision: s.revision, RequestID: "set-intelligence", Edit: edit}
	value := call(t, s.ctx, s.mcp, "edit_build", map[string]any{"edit": edit, "expected_revision": s.revision, "request_id": command.RequestID}, false)
	result := structuredView(t, value.StructuredContent)
	s.revision++
	s.expectEvent(s.revision, "build.model.passive_attributes")
	checkSwitchedXML(t, allocated, result, preview)
	return command, result
}

func checkSwitchedXML(t *testing.T, before, after application.SnapshotView, preview application.BuildComparison) {
	t.Helper()
	if after.Build.InputSHA256 != preview.Comparison.After.InputSHA256 || after.Build.Calculation != nil {
		t.Fatal("смена атрибута разошлась с предпросмотром")
	}
	if !maps.Equal(after.Build.Model.PassiveAttributes, map[string]build.PassiveAttribute{"1207": build.Dexterity, "13397": build.Intelligence}) {
		t.Fatalf("неверные выбранные атрибуты: %v", after.Build.Model.PassiveAttributes)
	}
	want := strings.NewReplacer(`dexNodes="23570,1207,13397"`, `dexNodes="23570,1207"`, `intNodes=""`, `intNodes="13397"`).Replace(before.Build.XML)
	if want == before.Build.XML || after.Build.XML != want {
		t.Fatal("смена атрибута изменила XML за пределами выбора узла")
	}
}
