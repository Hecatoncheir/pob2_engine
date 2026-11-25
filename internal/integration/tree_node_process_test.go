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

func TestLuaTreeNodeAcrossProcesses(t *testing.T) {
	if os.Getenv("POB2_ENGINE_POB_DIR") == "" {
		t.Skip("для проверки Lua укажите POB2_ENGINE_POB_DIR")
	}
	t.Setenv("POB2_ENGINE_CALCULATOR", "lua")
	s := newBuildScenario(t, "0.0.10/warrior.xml")
	info, err := s.client.Info(s.ctx)
	if err != nil || !slices.Contains(info.GetCapabilities(), "build.tree_preview") {
		t.Fatal("предпросмотр дерева отсутствует в возможностях движка", err)
	}
	before := calculateBeforeEquipment(t, s)
	checkTreeNodeRejections(t, s)
	allocate := treeNodeEdit(13397, "allocate")
	preview := compareTreeNode(t, s, allocate)
	checkAllocatedPath(t, preview)
	if afterPreview := supportSnapshot(t, s); !reflect.DeepEqual(before, afterPreview) {
		t.Fatal("предпросмотр изменил сессию")
	}
	command := application.ApplyBuild{SessionID: "main", ExpectedRevision: s.revision, RequestID: "passives-allocate", Edit: allocate}
	applied := applyTreeNodeCLI(t, s, command, preview)
	checkAllocatedXML(t, s)
	checkFirstReportAfterEdit(t, s, before, applied)
	refund := refundTreeNode(t, s)
	if refund.Build.InputSHA256 != before.Build.InputSHA256 {
		t.Fatal("снятие ветви изменило прочие данные сборки")
	}
	checkBuildEditRestart(t, s, command, applied, refund)
}

func treeNodeEdit(id int, action string) *build.Edit {
	return &build.Edit{TreeNode: &build.TreeNodeEdit{NodeID: id, Action: action}}
}

func compareTreeNode(t *testing.T, s *groupScenario, edit *build.Edit) application.BuildComparison {
	t.Helper()
	return compareTreeNodeStatus(t, s, edit, 0)
}

func compareTreeNodeStatus(t *testing.T, s *groupScenario, edit *build.Edit, exitCode int) application.BuildComparison {
	t.Helper()
	fromCLI := compareBuildCLI(t, s, edit, exitCode)
	answer := call(t, s.ctx, s.mcp, "compare_build", map[string]any{"edit": edit, "expected_revision": s.revision}, false)
	fromMCP := structuredCatalogJSON[application.BuildComparison](t, answer.StructuredContent)
	fromGRPC, err := s.client.Compare(s.ctx, application.CompareBuild{SessionID: "main", ExpectedRevision: s.revision, Edit: edit})
	if err != nil || !reflect.DeepEqual(fromCLI, fromMCP) || !reflect.DeepEqual(fromCLI, fromGRPC) {
		t.Fatal("отличаются результаты CLI, MCP и gRPC", err)
	}
	return fromGRPC
}

func checkAllocatedPath(t *testing.T, preview application.BuildComparison) {
	t.Helper()
	if preview.Passives == nil || !slices.Equal(preview.Passives.Added, []int{3936, 13397}) || len(preview.Passives.Removed) != 0 {
		t.Fatalf("неверный путь: %+v", preview.Passives)
	}
	before, after := preview.Comparison.Before.Stats, preview.Comparison.After.Stats
	if after["Str"] != before["Str"]+5 || after["TotalDPS"] <= before["TotalDPS"] {
		t.Fatalf("в расчёте нет силы или бонуса урона: Str %g → %g, TotalDPS %g → %g", before["Str"], after["Str"], before["TotalDPS"], after["TotalDPS"])
	}
}

func checkTreeNodeRejections(t *testing.T, s *groupScenario) {
	t.Helper()
	for _, edit := range []*build.Edit{treeNodeEdit(1000000, "allocate"), treeNodeEdit(47175, "refund"), treeNodeEdit(13397, "refund")} {
		if _, err := s.client.Compare(s.ctx, application.CompareBuild{SessionID: "main", ExpectedRevision: s.revision, Edit: edit}); status.Code(err) != codes.InvalidArgument {
			t.Fatal("принята недопустимая правка дерева", edit.TreeNode, err)
		}
		call(t, s.ctx, s.mcp, "compare_build", map[string]any{"expected_revision": s.revision, "edit": edit}, true)
	}
}

func applyTreeNodeCLI(t *testing.T, s *groupScenario, command application.ApplyBuild, preview application.BuildComparison) application.SnapshotView {
	t.Helper()
	args := []string{"build", "edit", "--session", "main", "--revision", fmt.Sprint(s.revision), "--request-id", command.RequestID, "--file", equipmentEditFile(t, command.Edit)}
	applied := buildView(t, cli(t, s.ctx, 0, append(args, s.connection...)...))
	if applied.Build.InputSHA256 != preview.Comparison.After.InputSHA256 || applied.Build.Calculation != nil {
		t.Fatal("применение не совпало с предпросмотром или сохранило старый расчёт")
	}
	s.revision++
	s.expectEvent(s.revision, "build.model.nodes")
	command.RequestID = "passives-stale"
	if _, err := s.client.Apply(s.ctx, command); status.Code(err) != codes.Aborted {
		t.Fatal("принята правка дерева для старой ревизии", err)
	}
	return applied
}

func checkAllocatedXML(t *testing.T, s *groupScenario) {
	t.Helper()
	result, err := s.client.Export(s.ctx, "main")
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{`nodes="47175,38646,3936,13397"`, `nodes="47175,3936"`, `strNodes="13397"`, `dexNodes="23570"`, `SavedExtension value="keep"`, `LocalExtension value="keep"`} {
		if !strings.Contains(result.XML, field) {
			t.Fatal("в экспорте пропущено поле", field)
		}
	}
}

func refundTreeNode(t *testing.T, s *groupScenario) application.SnapshotView {
	t.Helper()
	edit := treeNodeEdit(3936, "refund")
	preview := compareTreeNode(t, s, edit)
	if preview.Passives == nil || !slices.Equal(preview.Passives.Removed, []int{3936, 13397}) || len(preview.Passives.Added) != 0 {
		t.Fatalf("пропущены зависимые узлы: %+v", preview.Passives)
	}
	value := call(t, s.ctx, s.mcp, "edit_build", map[string]any{"edit": edit, "expected_revision": s.revision, "request_id": "passives-refund"}, false)
	result := structuredView(t, value.StructuredContent)
	if result.Build.InputSHA256 != preview.Comparison.After.InputSHA256 {
		t.Fatal("снятие не совпало с предпросмотром")
	}
	s.revision++
	s.expectEvent(s.revision, "build.model.nodes")
	return result
}
