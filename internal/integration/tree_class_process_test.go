package integration_test

import (
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/build"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestLuaTreeClassAcrossProcesses(t *testing.T) {
	if os.Getenv("POB2_ENGINE_POB_DIR") == "" {
		t.Skip("для проверки Lua укажите POB2_ENGINE_POB_DIR")
	}
	t.Setenv("POB2_ENGINE_CALCULATOR", "lua")
	s := newBuildScenarioWithTimeout(t, "tree-specs/warrior-witch.xml", 90*time.Second)
	info, err := s.client.Info(s.ctx)
	if err != nil || !slices.Contains(info.GetCapabilities(), "build.tree_class") {
		t.Fatal("движок не объявил смену класса", err)
	}
	before := supportSnapshot(t, s)
	rejectClassWithoutPath(t, s)
	edit := &build.Edit{TreeClass: &build.TreeClassEdit{ClassID: 1, AscendancyID: 1, Mode: "reset"}}
	preview := compareTreeNodeStatus(t, s, edit, 0)
	if supportSnapshot(t, s).Build.InputSHA256 != before.Build.InputSHA256 {
		t.Fatal("предпросмотр класса изменил сборку")
	}
	after := applyTreeClassMCP(t, s, edit)
	checkNewTreeClass(t, before.Build, after.Build, preview)
}

func rejectClassWithoutPath(t *testing.T, s *groupScenario) {
	t.Helper()
	edit := &build.Edit{TreeClass: &build.TreeClassEdit{ClassID: 1, AscendancyID: 1}}
	_, err := s.client.Apply(s.ctx, application.ApplyBuild{
		SessionID: "main", ExpectedRevision: s.revision, RequestID: "class-needs-path", Edit: edit,
	})
	if status.Code(err) != codes.InvalidArgument || !strings.Contains(err.Error(), "соединения или сброса") {
		t.Fatal("PoB принял смену класса с несвязанным деревом", err)
	}
}

func applyTreeClassMCP(t *testing.T, s *groupScenario, edit *build.Edit) application.SnapshotView {
	t.Helper()
	answer := call(t, s.ctx, s.mcp, "edit_build", map[string]any{
		"edit": edit, "request_id": "class-reset", "expected_revision": s.revision,
	}, false)
	result := structuredView(t, answer.StructuredContent)
	s.revision++
	s.expectEvent(s.revision, "build.model.class")
	return result
}

func checkNewTreeClass(t *testing.T, before, after *application.BuildView, preview application.BuildComparison) {
	t.Helper()
	if after.InputSHA256 != preview.Comparison.After.InputSHA256 ||
		after.Model.Class != "Witch" || after.Model.ClassID != 1 ||
		after.Model.Ascendancy != "Infernalist" || after.Model.AscendancyID != 1 ||
		after.Model.ActiveSets.Tree != 1 {
		t.Fatal("сохранённый класс или восхождение расходится с предпросмотром")
	}
	const marker = `title="Ведьма"`
	_, tail, ok := strings.Cut(before.XML, marker)
	_, updated, found := strings.Cut(after.XML, marker)
	if !ok || !found || updated != tail {
		t.Fatal("смена класса затронула вторую спецификацию или данные за деревом")
	}
}
