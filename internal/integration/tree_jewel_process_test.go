package integration_test

import (
	"fmt"
	"os"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/build"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const rubyText = "Rarity: RARE\nReference Ruby\nRuby\n--------\n+17% to Fire Resistance"

func TestLuaJewelAcrossProcesses(t *testing.T) {
	if os.Getenv("POB2_ENGINE_POB_DIR") == "" {
		t.Skip("для проверки Lua укажите POB2_ENGINE_POB_DIR")
	}
	t.Setenv("POB2_ENGINE_CALCULATOR", "lua")
	s := newBuildScenarioWithTimeout(t, "0.0.10/warrior.xml", 120*time.Second)
	info, err := s.client.Info(s.ctx)
	if err != nil || !slices.Contains(info.GetCapabilities(), "build.tree_jewels") {
		t.Fatal("движок не объявил поддержку самоцветов", err)
	}
	edit := jewelEdit(0, rubyText)
	rejectJewelEdit(t, s, edit, "Сначала назначьте гнездо")
	allocate := weaponTreeEdit(warriorSocket, "allocate", build.AllocationCommon)
	applyTreeNodeCLI(t, s, application.ApplyBuild{
		SessionID: "main", ExpectedRevision: s.revision, RequestID: "jewel-path", Edit: allocate,
	}, compareTreeNode(t, s, allocate))
	before := supportSnapshot(t, s)
	call(t, s.ctx, s.mcp, "edit_build", map[string]any{
		"edit": jewelEdit(1, rubyText), "expected_revision": s.revision, "request_id": "jewel-both-sources",
	}, true)
	rejectJewelEdit(t, s, jewelEdit(128, ""), "отсутствует в сборке")
	rejectJewelEdit(t, s, jewelEdit(1, ""), "несовместим со слотом")
	preview := compareTreeNode(t, s, edit)
	if !reflect.DeepEqual(before, supportSnapshot(t, s)) {
		t.Fatal("примерка или отказ изменили сборку")
	}
	stats := preview.Comparison
	if stats.After.Stats["FireResist"]-stats.Before.Stats["FireResist"] != 17 {
		t.Fatalf("PoB не применил сопротивление самоцвета: %+v", stats.After.Stats)
	}
	applied := installJewelFromText(t, s, edit, preview)
	removed := removeJewelThroughMCP(t, s, applied, stats.Before.Stats["FireResist"])
	installExistingJewelCLI(t, s, applied.Build.Model.PassiveJewels[strconv.Itoa(warriorSocket)], removed, applied)
}

func jewelEdit(itemID int, raw string) *build.Edit {
	return &build.Edit{Jewel: &build.JewelEdit{NodeID: warriorSocket, ItemID: itemID, Raw: raw}}
}

func rejectJewelEdit(t *testing.T, s *groupScenario, edit *build.Edit, message string) {
	t.Helper()
	_, err := s.client.Apply(s.ctx, application.ApplyBuild{
		SessionID: "main", ExpectedRevision: s.revision, RequestID: "jewel-rejected", Edit: edit,
	})
	if status.Code(err) != codes.InvalidArgument || !strings.Contains(status.Convert(err).Message(), message) {
		t.Fatal("не получен ожидаемый отказ", err)
	}
}

func installJewelFromText(t *testing.T, s *groupScenario, edit *build.Edit, preview application.BuildComparison) application.SnapshotView {
	t.Helper()
	command := application.ApplyBuild{SessionID: "main", ExpectedRevision: s.revision, RequestID: "jewel-from-text", Edit: edit}
	_, err := s.client.Apply(s.ctx, command)
	if err != nil {
		t.Fatal(err)
	}
	s.revision++
	expectJewelEvent(t, s, 2)
	value := supportSnapshot(t, s)
	if value.Build.InputSHA256 != preview.Comparison.After.InputSHA256 || value.Build.Model.PassiveJewels[strconv.Itoa(warriorSocket)] != 2 {
		t.Fatal("запись самоцвета не совпала с предпросмотром")
	}
	command.RequestID = "jewel-stale"
	if _, err := s.client.Apply(s.ctx, command); status.Code(err) != codes.Aborted {
		t.Fatal("принята правка самоцвета по старой ревизии", err)
	}
	return value
}

func expectJewelEvent(t *testing.T, s *groupScenario, itemID uint32) {
	t.Helper()
	event, err := s.events.Recv()
	if err != nil || event.GetSnapshot().GetRevision() != s.revision ||
		!slices.Contains(event.GetChangedFields(), "build.model.passive_jewels") ||
		event.GetSnapshot().GetBuild().GetModel().GetPassiveJewels()[warriorSocket] != itemID {
		t.Fatal("поток gRPC не доставил новую ревизию и назначение самоцвета", err)
	}
}

func removeJewelThroughMCP(t *testing.T, s *groupScenario, installed application.SnapshotView, fireResist float64) application.SnapshotView {
	t.Helper()
	edit := jewelEdit(0, "")
	preview := compareTreeNode(t, s, edit)
	if preview.Comparison.After.Stats["FireResist"] != fireResist {
		t.Fatal("снятый самоцвет продолжает влиять на расчёт Lua")
	}
	answer := call(t, s.ctx, s.mcp, "edit_build", map[string]any{
		"edit": edit, "expected_revision": s.revision, "request_id": "jewel-remove",
	}, false)
	removed := structuredView(t, answer.StructuredContent)
	s.revision++
	expectJewelEvent(t, s, 0)
	if removed.Build.InputSHA256 != preview.Comparison.After.InputSHA256 ||
		len(removed.Build.Model.PassiveJewels) != 0 || !reflect.DeepEqual(installed.Build.Model.Items, removed.Build.Model.Items) ||
		!slices.Equal(installed.Build.Model.Nodes, removed.Build.Model.Nodes) {
		t.Fatal("снятие самоцвета изменило предметы, узлы или результат предпросмотра")
	}
	return removed
}

func installExistingJewelCLI(t *testing.T, s *groupScenario, itemID int, removed, installed application.SnapshotView) {
	t.Helper()
	edit := jewelEdit(itemID, "")
	preview := compareTreeNode(t, s, edit)
	args := []string{"build", "edit", "--session", "main", "--revision", fmt.Sprint(s.revision), "--request-id", "jewel-from-list", "--file", equipmentEditFile(t, edit)}
	value := buildView(t, cli(t, s.ctx, 0, append(args, s.connection...)...))
	s.revision++
	expectJewelEvent(t, s, uint32(itemID))
	if value.Build.InputSHA256 != installed.Build.InputSHA256 || value.Build.InputSHA256 != preview.Comparison.After.InputSHA256 ||
		!reflect.DeepEqual(value.Build.Model.Items, removed.Build.Model.Items) {
		t.Fatal("выбор предмета из сборки создал копию или изменил данные")
	}
}
