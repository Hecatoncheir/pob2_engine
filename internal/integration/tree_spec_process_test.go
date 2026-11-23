package integration_test

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/build"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestLuaTreeSpecsAcrossProcesses(t *testing.T) {
	if os.Getenv("POB2_ENGINE_POB_DIR") == "" {
		t.Skip("для проверки Lua укажите POB2_ENGINE_POB_DIR")
	}
	t.Setenv("POB2_ENGINE_CALCULATOR", "lua")
	s := newBuildScenarioWithTimeout(t, "tree-specs/warrior-witch.xml", 75*time.Second)
	before := supportSnapshot(t, s)
	rejectTreeSpec(t, s, 3, s.revision, codes.InvalidArgument)
	edit := treeSpecEdit(2)
	preview := compareTreeNodeStatus(t, s, edit, 4)
	checkTreeSpecCalculation(t, preview)
	if !reflect.DeepEqual(before, supportSnapshot(t, s)) {
		t.Fatal("отказ или предпросмотр изменили сессию")
	}
	selected := applyTreeSpecCLI(t, s, edit)
	checkSelectedTreeSpec(t, before.Build, selected.Build, preview)
	rejectTreeSpec(t, s, 1, before.Revision, codes.Aborted)
	restoreTreeSpecMCP(t, s, before)
}

func checkTreeSpecCalculation(t *testing.T, preview application.BuildComparison) {
	t.Helper()
	before, after := preview.Comparison.Before, preview.Comparison.After
	if before.Stats["FireResist"] != 10 || after.Stats["FireResist"] != 27 {
		t.Fatal("Lua не учёл самоцвет выбранной спецификации")
	}
	warning := build.Issue{Code: "pob_warning", Message: "You have 1 Weapon set 2 passives available"}
	if len(before.Issues) != 0 || !reflect.DeepEqual(after.Issues, []build.Issue{warning}) {
		t.Fatalf("диагностика Lua изменилась: %v → %v", before.Issues, after.Issues)
	}
}

func treeSpecEdit(id int) *build.Edit {
	return &build.Edit{Selection: &build.SelectionEdit{Sets: &build.ActiveSets{Tree: id}}}
}

func rejectTreeSpec(t *testing.T, s *groupScenario, id int, revision uint64, expected codes.Code) {
	t.Helper()
	_, err := s.client.Apply(s.ctx, application.ApplyBuild{
		SessionID: "main", ExpectedRevision: revision, RequestID: "spec-rejected", Edit: treeSpecEdit(id),
	})
	if status.Code(err) != expected {
		t.Fatalf("неверный ответ для спецификации %d: %v", id, err)
	}
}

func applyTreeSpecCLI(t *testing.T, s *groupScenario, edit *build.Edit) application.SnapshotView {
	t.Helper()
	args := []string{"build", "edit", "--session", "main", "--revision", fmt.Sprint(s.revision),
		"--request-id", "spec-witch", "--file", equipmentEditFile(t, edit)}
	value := buildView(t, cli(t, s.ctx, 0, append(args, s.connection...)...))
	s.revision++
	s.expectEvent(s.revision, "build.model.active_sets")
	return value
}

func checkSelectedTreeSpec(t *testing.T, before, after *application.BuildView, preview application.BuildComparison) {
	t.Helper()
	if after.InputSHA256 != preview.Comparison.After.InputSHA256 || after.Model.ActiveSets.Tree != 2 ||
		after.Model.Class != "Witch" || after.Model.ClassID != 1 || after.Model.PassiveJewels["61419"] != 2 ||
		after.Model.PassiveAllocations["4739"] != build.AllocationWeaponSet2 {
		t.Fatal("изменение спецификации не совпало с предпросмотром или Lua")
	}
	want := strings.Replace(before.XML, `activeSpec="1"`, `activeSpec="2"`, 1)
	want = strings.Replace(want, `className="Warrior"`, `className="Witch"`, 1)
	if after.XML != want {
		t.Fatal("при переключении изменились данные вне выбора спецификации и класса")
	}
}

func restoreTreeSpecMCP(t *testing.T, s *groupScenario, before application.SnapshotView) {
	t.Helper()
	value := call(t, s.ctx, s.mcp, "edit_build", map[string]any{
		"edit": treeSpecEdit(1), "expected_revision": s.revision, "request_id": "spec-warrior",
	}, false)
	result := structuredView(t, value.StructuredContent)
	s.revision++
	s.expectEvent(s.revision, "build.model.active_sets")
	if result.Build.InputSHA256 != before.Build.InputSHA256 || result.Build.XML != before.Build.XML {
		t.Fatal("возврат к исходной спецификации не восстановил исходный XML")
	}
}
