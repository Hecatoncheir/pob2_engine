package integration_test

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/build"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type workshopSkillChoice struct {
	Snapshot struct {
		BuildInfo struct {
			StatSets struct {
				Selected int `json:"selected"`
				Entries  []struct {
					ID    int    `json:"id"`
					Label string `json:"label"`
				} `json:"entries"`
			} `json:"statSets"`
		} `json:"buildInfo"`
	} `json:"snapshot"`
	Stats struct {
		TotalDPS float64 `json:"TotalDPS"`
	} `json:"stats"`
}

func TestLuaSkillChoiceAcrossProcesses(t *testing.T) {
	if os.Getenv("POB2_ENGINE_POB_DIR") == "" {
		t.Skip("для проверки Lua укажите POB2_ENGINE_POB_DIR")
	}
	t.Setenv("POB2_ENGINE_CALCULATOR", "lua")
	s := newBuildScenarioWithTimeout(t, "tree-specs/warrior-witch.xml", 90*time.Second)
	addVariantGroup(t, s)
	selectSkillGroupCLI(t, s)
	before := supportSnapshot(t, s)
	baseline := readVariantReport(t, s)
	id := infusedVariant(t, baseline)
	chooseVariantMCP(t, s, id)
	after := supportSnapshot(t, s)
	variant := readVariantReport(t, s)
	if variant.Snapshot.BuildInfo.StatSets.Selected != id || variant.Stats.TotalDPS == baseline.Stats.TotalDPS {
		t.Fatal("Lua не восстановил выбранный вариант из сохранённого XML", variant.Snapshot.BuildInfo.StatSets)
	}
	_, original, _ := strings.Cut(before.Build.XML, "<Tree")
	_, updated, _ := strings.Cut(after.Build.XML, "<Tree")
	if original == "" || original != updated {
		t.Fatal("выбор умения изменил дерево, условия или заметки")
	}
}

func addVariantGroup(t *testing.T, s *groupScenario) {
	t.Helper()
	edit := &build.Edit{Gems: &build.GemsEdit{Label: "Ball Lightning", Enabled: true,
		Gems: []build.GemInput{{ID: "Metadata/Items/Gems/SkillGemBallLightning", Level: 1, Enabled: true}}}}
	view, err := s.client.Apply(s.ctx, application.ApplyBuild{
		SessionID: "main", ExpectedRevision: s.revision, RequestID: "add-ball-lightning", Edit: edit,
	})
	if err != nil || view.Revision != s.revision+1 {
		t.Fatal("не добавлена вторая группа", err)
	}
	s.revision = view.Revision
}

func selectSkillGroupCLI(t *testing.T, s *groupScenario) {
	t.Helper()
	file := filepath.Join(t.TempDir(), "choice.json")
	if err := os.WriteFile(file, []byte(`{"skill_choice":{"control":"groups","option_id":2}}`), 0600); err != nil {
		t.Fatal(err)
	}
	args := []string{"build", "edit", "--session", "main", "--revision", fmt.Sprint(s.revision), "--request-id", "choose-group", "--file", file}
	view := decodeCatalogJSON[application.SnapshotView](t, cli(t, s.ctx, 0, append(args, s.connection...)...))
	if view.Build.Model.MainGroup != 2 || view.Revision != s.revision+1 {
		t.Fatal("не выбрана группа Ball Lightning")
	}
	s.revision = view.Revision
}

func readVariantReport(t *testing.T, s *groupScenario) workshopSkillChoice {
	t.Helper()
	view, err := s.client.BuildReport(s.ctx, "main", s.revision)
	if err != nil {
		t.Fatal(err)
	}
	return decodeCatalogJSON[workshopSkillChoice](t, view.Report)
}

func infusedVariant(t *testing.T, report workshopSkillChoice) int {
	t.Helper()
	for _, option := range report.Snapshot.BuildInfo.StatSets.Entries {
		if option.Label == "Fire-Infused" && option.ID != report.Snapshot.BuildInfo.StatSets.Selected {
			return option.ID
		}
	}
	t.Fatal("потерян вариант Fire-Infused")
	return 0
}

func chooseVariantMCP(t *testing.T, s *groupScenario, id int) {
	t.Helper()
	edit := &build.Edit{SkillChoice: &build.SkillChoiceEdit{Control: "stat_sets", OptionID: id}}
	before := supportSnapshot(t, s)
	preview, err := s.client.Compare(s.ctx, application.CompareBuild{SessionID: "main", ExpectedRevision: s.revision, Edit: edit})
	if err != nil || preview.Comparison.Before.InputSHA256 != before.Build.InputSHA256 {
		t.Fatal("gRPC не построил предпросмотр выбранного варианта", err)
	}
	answer := call(t, s.ctx, s.mcp, "edit_build", map[string]any{
		"edit": edit, "request_id": "choose-infused", "expected_revision": s.revision,
	}, false)
	result := structuredView(t, answer.StructuredContent)
	if result.Revision != s.revision+1 || result.Build.InputSHA256 != preview.Comparison.After.InputSHA256 {
		t.Fatal("MCP применил вариант с другим XML")
	}
	_, err = s.client.Apply(s.ctx, application.ApplyBuild{SessionID: "main", ExpectedRevision: s.revision,
		RequestID: "obsolete-variant", Edit: edit})
	if status.Code(err) != codes.Aborted {
		t.Fatal("gRPC принял выбор по устаревшей ревизии", err)
	}
	s.revision = result.Revision
	info, err := s.client.Info(s.ctx)
	if err != nil || !slices.Contains(info.GetCapabilities(), "build.skill_choice") {
		t.Fatal("движок не объявил выбор умения", err)
	}
}
