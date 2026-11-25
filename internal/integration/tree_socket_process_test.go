package integration_test

import (
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

const warriorSocket = 26725

func TestLuaEmptySocketAcrossProcesses(t *testing.T) {
	if os.Getenv("POB2_ENGINE_POB_DIR") == "" {
		t.Skip("для проверки Lua укажите POB2_ENGINE_POB_DIR")
	}
	t.Setenv("POB2_ENGINE_CALCULATOR", "lua")
	s := newBuildScenarioWithTimeout(t, "0.0.10/warrior.xml", 90*time.Second)
	info, err := s.client.Info(s.ctx)
	if err != nil || !slices.Contains(info.GetCapabilities(), "build.tree_sockets") {
		t.Fatal("движок не объявил поддержку гнёзд", err)
	}
	before := supportSnapshot(t, s)
	rejectSocketEdit(t, s, weaponTreeEdit(warriorSocket, "allocate", build.AllocationWeaponSet1), "общем режиме")
	edit := weaponTreeEdit(warriorSocket, "allocate", build.AllocationCommon)
	preview := compareTreeNode(t, s, edit)
	checkSocketPreview(t, preview)
	if !reflect.DeepEqual(before, supportSnapshot(t, s)) {
		t.Fatal("предпросмотр или отклонённая правка изменили сессию")
	}
	applied := applyTreeNodeCLI(t, s, application.ApplyBuild{
		SessionID: "main", ExpectedRevision: s.revision, RequestID: "socket-allocate", Edit: edit,
	}, preview)
	checkFirstReportAfterEdit(t, s, before, applied)
	refundEmptySocket(t, s, applied)
	checkOccupiedSocket(t, s, applied.Build.XML)
}

func checkSocketPreview(t *testing.T, preview application.BuildComparison) {
	t.Helper()
	change := preview.Passives
	if change == nil || !slices.Contains(change.Added, warriorSocket) || len(change.Removed) != 0 {
		t.Fatalf("неверный путь к гнезду: %+v", change)
	}
	want := build.PassiveAllocationChange{NodeID: warriorSocket, After: build.AllocationCommon}
	if !slices.Contains(change.Allocations, want) || change.Points == nil {
		t.Fatalf("неверные режим или расход очков: %+v", change)
	}
	points := change.Points
	if points.After.Passives-points.Before.Passives != len(change.Added) ||
		points.After.WeaponSet1 != points.Before.WeaponSet1 || points.After.WeaponSet2 != points.Before.WeaponSet2 {
		t.Fatalf("путь к гнезду использует неверные очки: %+v", points)
	}
}

func refundEmptySocket(t *testing.T, s *groupScenario, applied application.SnapshotView) {
	t.Helper()
	if !slices.Contains(applied.Build.Model.Nodes, warriorSocket) {
		t.Fatal("в применённом XML нет гнезда")
	}
	edit := weaponTreeEdit(warriorSocket, "refund", build.AllocationCommon)
	preview := compareTreeNode(t, s, edit)
	if preview.Passives == nil || !slices.Equal(preview.Passives.Removed, []int{warriorSocket}) {
		t.Fatalf("при снятии гнезда затронуты другие узлы: %+v", preview.Passives)
	}
	value := call(t, s.ctx, s.mcp, "edit_build", map[string]any{
		"edit": edit, "expected_revision": s.revision, "request_id": "socket-refund",
	}, false)
	result := structuredView(t, value.StructuredContent)
	if result.Build.InputSHA256 != preview.Comparison.After.InputSHA256 || slices.Contains(result.Build.Model.Nodes, warriorSocket) {
		t.Fatal("снятие гнезда не совпало с предпросмотром")
	}
	s.revision++
	s.expectEvent(s.revision, "build.model.nodes")
}

func checkOccupiedSocket(t *testing.T, s *groupScenario, xml string) {
	t.Helper()
	xml = strings.Replace(xml, "</Items>", "<Item id=\"2\">Rarity: NORMAL\nRuby</Item></Items>", 1)
	xml = strings.Replace(xml, "</Overrides>", `</Overrides><Sockets><Socket nodeId="26725" itemId="2"/></Sockets>`, 1)
	imported, err := s.client.Apply(s.ctx, application.ApplyBuild{
		SessionID: "main", ExpectedRevision: s.revision, RequestID: "socket-import-ruby", XML: xml,
	})
	if err != nil {
		t.Fatal(err)
	}
	s.revision = imported.Revision
	s.expectEvent(s.revision, "build.model.nodes")
	before := supportSnapshot(t, s)
	rejectSocketEdit(t, s, weaponTreeEdit(warriorSocket, "refund", build.AllocationCommon), "выньте самоцвет")
	if !reflect.DeepEqual(before, supportSnapshot(t, s)) {
		t.Fatal("отклонённая правка изменила сборку с самоцветом")
	}
}

func rejectSocketEdit(t *testing.T, s *groupScenario, edit *build.Edit, message string) {
	t.Helper()
	_, err := s.client.Apply(s.ctx, application.ApplyBuild{
		SessionID: "main", ExpectedRevision: s.revision, RequestID: "socket-rejected", Edit: edit,
	})
	if status.Code(err) != codes.InvalidArgument || !strings.Contains(status.Convert(err).Message(), message) {
		t.Fatal("не получен ожидаемый отказ", err)
	}
	call(t, s.ctx, s.mcp, "compare_build", map[string]any{"edit": edit, "expected_revision": s.revision}, true)
}
