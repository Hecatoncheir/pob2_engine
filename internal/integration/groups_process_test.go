package integration_test

import (
	"context"
	"fmt"
	"math"
	"path/filepath"
	"reflect"
	"slices"
	"testing"
	"time"

	pb "github.com/Hecatoncheir/pob2_engine/gen/go/pob2/v1"
	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/build"
	"github.com/Hecatoncheir/pob2_engine/internal/transport/grpcapi"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type groupEvents interface {
	Recv() (*pb.WatchSessionResponse, error)
}

type groupScenario struct {
	t          *testing.T
	ctx        context.Context
	process    *process
	client     *grpcapi.Client
	mcp        *mcp.ClientSession
	events     groupEvents
	connection []string
	dataDir    string
	revision   uint64
}

func newGroupScenario(t *testing.T) *groupScenario {
	t.Helper()
	return newBuildScenario(t, "0.0.2/groups/cases/druid-group-1.xml")
}

func newBuildScenario(t *testing.T, fixture string) *groupScenario {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 40*time.Second)
	t.Cleanup(cancel)
	scenario := &groupScenario{t: t, ctx: ctx, dataDir: t.TempDir(), revision: 2}
	scenario.process = startServer(t, ctx, "--profile", "groups", "--runtime-dir", t.TempDir(), "--data-dir", scenario.dataDir)
	scenario.connect()
	cli(t, ctx, 0, append([]string{"session", "create", "--session", "main", "--name", "Группы", "--request-id", "create"}, scenario.connection...)...)
	var err error
	scenario.events, err = scenario.client.Watch(ctx, "main", 1)
	if err != nil {
		t.Fatal(err)
	}
	file := filepath.Join("..", "..", "testdata", "reference", fixture)
	cli(t, ctx, 0, append([]string{"build", "import", "--session", "main", "--revision", "1", "--request-id", "import", "--file", file}, scenario.connection...)...)
	scenario.expectEvent(2, "build.model.groups")
	scenario.mcp = connectMCP(t, ctx, append([]string{"--attach", "main"}, scenario.connection...)...)
	return scenario
}

func (scenario *groupScenario) connect() {
	scenario.t.Helper()
	scenario.connection = []string{"--address", scenario.process.ready.Address, "--token", scenario.process.ready.Token}
	client, err := grpcapi.Dial(scenario.ctx, scenario.process.ready.Address, scenario.process.ready.Token)
	if err != nil {
		scenario.t.Fatal(err)
	}
	scenario.client = client
	scenario.t.Cleanup(func() { _ = client.Close() })
}

func (scenario *groupScenario) expectEvent(revision uint64, field string) {
	scenario.t.Helper()
	event, err := scenario.events.Recv()
	if err != nil || event.GetSnapshot().GetRevision() != revision || !slices.Contains(event.GetChangedFields(), field) {
		scenario.t.Fatalf("нет события %d / %scenario: %v %v", revision, field, event, err)
	}
}

func (scenario *groupScenario) setEnabled(group int, skill string, enabled bool) {
	scenario.t.Helper()
	edit := build.Edit{Gems: &build.GemsEdit{GroupID: group, Enabled: enabled, Gems: []build.GemInput{{ID: skill + "Player", Level: 1, Enabled: true}}}}
	view := structuredView(scenario.t, call(scenario.t, scenario.ctx, scenario.mcp, "edit_build", map[string]any{
		"edit": edit, "expected_revision": scenario.revision, "request_id": fmt.Sprintf("group-%d-%t", group, enabled),
	}, false).StructuredContent)
	scenario.revision++
	if view.Revision != scenario.revision || view.Build.Calculation != nil {
		scenario.t.Fatal("MCP не изменил группу или оставил старый расчёт")
	}
	scenario.expectEvent(scenario.revision, "build.model.groups")
}

func (scenario *groupScenario) selectSkills(id string, selection build.SelectionEdit, field string) application.ApplyBuild {
	scenario.t.Helper()
	command := application.ApplyBuild{SessionID: "main", ExpectedRevision: scenario.revision, RequestID: id, Edit: &build.Edit{Selection: &selection}}
	value, err := scenario.client.Apply(scenario.ctx, command)
	if err != nil {
		scenario.t.Fatal(err)
	}
	scenario.revision++
	if value.Revision != scenario.revision || value.CalculationJSON != "" {
		scenario.t.Fatal("выбор не сбросил старый результат")
	}
	scenario.expectEvent(scenario.revision, field)
	return command
}

func (scenario *groupScenario) calculate(id, skill string, main, set int, dps float64) application.SnapshotView {
	scenario.t.Helper()
	value := buildView(scenario.t, cli(scenario.t, scenario.ctx, 0, append([]string{"calculate", "--session", "main", "--revision", fmt.Sprint(scenario.revision), "--request-id", id}, scenario.connection...)...))
	scenario.revision++
	scenario.expectEvent(scenario.revision, "calculation")
	model, result := value.Build.Model, value.Build.Calculation
	if value.Revision != scenario.revision || model.ClassID != 11 || model.MainGroup != main || model.ActiveSets.Skills != set || result == nil {
		scenario.t.Fatal("не сохранён выбранный класс или группа")
	}
	if result.Status != "complete" || result.Skill != skill || math.Abs(result.Stats["TotalDPS"]-dps) > 1e-9 {
		scenario.t.Fatalf("рассчитано другое умение: %+v", result)
	}
	return value
}

func TestGroupAndSetSelectionSurvivesCrashWithoutRepeatingCommands(t *testing.T) {
	scenario := newGroupScenario(t)
	steps := []struct {
		skill string
		dps   float64
	}{{"Fireball", 8.9166666666667}, {"Frostbolt", 9.1}, {"Spark", 8.5642857142857}}
	for i, step := range steps {
		if i > 0 {
			scenario.setEnabled(i, steps[i-1].skill, false)
			scenario.setEnabled(i+1, step.skill, true)
			scenario.selectSkills("main-"+step.skill, build.SelectionEdit{MainGroup: i + 1}, "build.model.main_group")
		}
		scenario.calculate("calculate-"+step.skill, step.skill, i+1, 1, step.dps)
	}
	_, err := scenario.client.Apply(scenario.ctx, application.ApplyBuild{SessionID: "main", ExpectedRevision: scenario.revision, RequestID: "missing-main", Edit: &build.Edit{Selection: &build.SelectionEdit{Sets: &build.ActiveSets{Skills: 2}}}})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("не отклонена отсутствующая основная группа: %v", err)
	}
	selection := scenario.selectSkills("alternate", build.SelectionEdit{MainGroup: 1, Sets: &build.ActiveSets{Skills: 2, Config: 2}}, "build.model.active_sets")
	calculation := application.CalculateBuild{SessionID: "main", ExpectedRevision: scenario.revision, RequestID: "calculate-alternate"}
	saved := scenario.calculate(calculation.RequestID, "Spark", 1, 2, 17.1285714285714)
	if saved.Build.Model.ActiveSets.Config != 2 {
		t.Fatal("не выбран набор условий")
	}
	scenario.checkRestart(selection, calculation, saved)
}

func (scenario *groupScenario) checkRestart(selection application.ApplyBuild, calculation application.CalculateBuild, saved application.SnapshotView) {
	scenario.t.Helper()
	_ = scenario.mcp.Close()
	_ = scenario.client.Close()
	if err := scenario.process.cmd.Process.Kill(); err != nil {
		scenario.t.Fatal(err)
	}
	<-scenario.process.done
	scenario.process = startServer(scenario.t, scenario.ctx, "--profile", "groups", "--runtime-dir", scenario.t.TempDir(), "--data-dir", scenario.dataDir)
	scenario.connect()
	current := buildView(scenario.t, cli(scenario.t, scenario.ctx, 0, append([]string{"session", "get", "--session", "main"}, scenario.connection...)...))
	if !reflect.DeepEqual(current, saved) {
		scenario.t.Fatal("перезапуск потерял выбранные наборы или расчёт")
	}
	old, err := scenario.client.Apply(scenario.ctx, selection)
	if err != nil || old.Revision != selection.ExpectedRevision+1 {
		scenario.t.Fatal("не восстановлен исходный ответ выбора")
	}
	replayed, err := scenario.client.Calculate(scenario.ctx, calculation)
	if err != nil {
		scenario.t.Fatal(err)
	}
	view, err := application.ViewSnapshot(replayed)
	if err != nil || !reflect.DeepEqual(view, saved) {
		scenario.t.Fatal("повтор расчёта подменил его исходный ответ")
	}
	current = buildView(scenario.t, cli(scenario.t, scenario.ctx, 0, append([]string{"session", "get", "--session", "main"}, scenario.connection...)...))
	if !reflect.DeepEqual(current, saved) {
		scenario.t.Fatal("повтор старой команды откатил сессию")
	}
	scenario.events, err = scenario.client.Watch(scenario.ctx, "main", selection.ExpectedRevision)
	if err != nil {
		scenario.t.Fatal(err)
	}
	scenario.expectEvent(selection.ExpectedRevision+1, "build.model.active_sets")
	scenario.expectEvent(calculation.ExpectedRevision+1, "calculation")
}
