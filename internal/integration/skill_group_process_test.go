package integration_test

import (
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

const ballLightningGem = "Metadata/Items/Gems/SkillGemBallLightning"
const rapidCastingGem = "Metadata/Items/Gems/SkillGemRapidCastingSupport"

func TestLuaSkillGroupAcrossProcesses(t *testing.T) {
	if os.Getenv("POB2_ENGINE_POB_DIR") == "" {
		t.Skip("для проверки Lua укажите POB2_ENGINE_POB_DIR")
	}
	t.Setenv("POB2_ENGINE_CALCULATOR", "lua")
	s := newBuildScenarioWithTimeout(t, "tree-specs/warrior-witch.xml", 90*time.Second)
	before := supportSnapshot(t, s)
	label := "Группа штатного редактора"
	edit := &build.Edit{SkillGroup: &build.SkillGroupEdit{Label: &label, Gems: []build.SkillGroupGem{{ID: ballLightningGem, Level: 1}}}}
	added := createAndCompareSkillGroup(t, s, edit)
	group := len(added.Build.Model.Groups)
	if group != len(before.Build.Model.Groups)+1 {
		t.Fatal("штатный Lua не добавил группу")
	}
	if cutNonSkillXML(added.Build.XML) != cutNonSkillXML(before.Build.XML) {
		t.Fatal("правка группы затронула дерево или следующие разделы XML")
	}
	checkSkillGroupUpdate(t, s, group)
	checkSkillGroupDelete(t, s, group, len(before.Build.Model.Groups))
}

func cutNonSkillXML(xml string) string {
	_, suffix, _ := strings.Cut(xml, "</Skills>")
	return suffix
}

func createAndCompareSkillGroup(t *testing.T, s *groupScenario, edit *build.Edit) application.SnapshotView {
	t.Helper()
	preview := compareSkillGroupAdapters(t, s, edit)
	answer := call(t, s.ctx, s.mcp, "edit_build", map[string]any{
		"edit": edit, "request_id": "lua-skill-group-add", "expected_revision": s.revision,
	}, false)
	added := structuredView(t, answer.StructuredContent)
	if added.Revision != s.revision+1 || added.Build.InputSHA256 != preview.Comparison.After.InputSHA256 {
		t.Fatal("создание группы через MCP разошлось с предпросмотром")
	}
	_, err := s.client.Apply(s.ctx, application.ApplyBuild{SessionID: "main", ExpectedRevision: s.revision, RequestID: "stale-group", Edit: edit})
	if status.Code(err) != codes.Aborted {
		t.Fatal("принята устаревшая форма группы", err)
	}
	s.revision = added.Revision
	return added
}

func compareSkillGroupAdapters(t *testing.T, s *groupScenario, edit *build.Edit) application.BuildComparison {
	t.Helper()
	view, err := s.client.Compare(s.ctx, application.CompareBuild{SessionID: "main", ExpectedRevision: s.revision, Edit: edit})
	if err != nil {
		t.Fatal(err)
	}
	fromCLI := compareBuildCLI(t, s, edit, 0)
	answer := call(t, s.ctx, s.mcp, "compare_build", map[string]any{"expected_revision": s.revision, "edit": edit}, false)
	fromMCP := structuredCatalogJSON[application.BuildComparison](t, answer.StructuredContent)
	if !reflect.DeepEqual(view, fromCLI) || !reflect.DeepEqual(view, fromMCP) {
		t.Fatal("предпросмотр группы различается в CLI, MCP и gRPC")
	}
	return view
}

func checkSkillGroupUpdate(t *testing.T, s *groupScenario, group int) {
	t.Helper()
	edit := &build.Edit{SkillGroup: &build.SkillGroupEdit{GroupID: group, Gems: []build.SkillGroupGem{
		{ID: ballLightningGem, Level: 1, SourceIndex: 1}, {ID: rapidCastingGem, Level: 1},
	}}}
	result, err := s.client.Apply(s.ctx, application.ApplyBuild{SessionID: "main", ExpectedRevision: s.revision, RequestID: "lua-skill-group-update", Edit: edit})
	if err != nil || result.Revision != s.revision+1 {
		t.Fatal("штатный Lua не применил поддержку", err)
	}
	s.revision = result.Revision
	updated := supportSnapshot(t, s)
	if len(updated.Build.Model.Groups[group-1].Gems) != 2 {
		t.Fatal("добавленная поддержка отсутствует в XML")
	}
	edit.SkillGroup.Gems[0].ID = rapidCastingGem
	_, err = s.client.Apply(s.ctx, application.ApplyBuild{SessionID: "main", ExpectedRevision: s.revision, RequestID: "bad-source", Edit: edit})
	if status.Code(err) != codes.InvalidArgument || !reflect.DeepEqual(updated, supportSnapshot(t, s)) {
		t.Fatal("неверный исходный камень изменил сборку", err)
	}
}

func checkSkillGroupDelete(t *testing.T, s *groupScenario, group, count int) {
	t.Helper()
	edit := &build.Edit{SkillGroup: &build.SkillGroupEdit{GroupID: group, Delete: true}}
	deleted, err := s.client.Apply(s.ctx, application.ApplyBuild{SessionID: "main", ExpectedRevision: s.revision, RequestID: "lua-skill-group-delete", Edit: edit})
	if err != nil || deleted.Revision != s.revision+1 || len(supportSnapshot(t, s).Build.Model.Groups) != count {
		t.Fatal("штатный Lua не удалил группу", err)
	}
}
