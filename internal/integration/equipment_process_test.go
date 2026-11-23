package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/build"
	"github.com/Hecatoncheir/pob2_engine/internal/calculation"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const armourText = "Rarity: RARE\nRelease Ward\nIron Cuirass\n--------\n+30 to maximum Life\n+20% to Fire Resistance"

func TestLuaEquipmentAcrossProcesses(t *testing.T) {
	if os.Getenv("POB2_ENGINE_POB_DIR") == "" {
		t.Skip("для проверки Lua укажите POB2_ENGINE_POB_DIR")
	}
	t.Setenv("POB2_ENGINE_CALCULATOR", "lua")
	s := newBuildScenario(t, "0.0.5/attacker.xml")
	info, err := s.client.Info(s.ctx)
	if err != nil || !slices.Contains(info.GetCapabilities(), "build.equipment_preview") {
		t.Fatal("не объявлена примерка экипировки", err)
	}
	before := calculateBeforeEquipment(t, s)
	checkRejectedEquipment(t, s)
	edit := &build.Edit{Equipment: &build.EquipmentEdit{Slot: "Body Armour", Raw: armourText}}
	preview := compareEquipment(t, s, edit)
	if current := supportSnapshot(t, s); !reflect.DeepEqual(current, before) {
		t.Fatal("примерка либо отказ изменили XML, ревизию или сохранённый расчёт")
	}
	command := application.ApplyBuild{SessionID: "main", ExpectedRevision: s.revision, RequestID: "armour", Edit: edit}
	applied := applyPreviewedEquipment(t, s, command, before, preview)
	removed := removeEquipment(t, s, applied)
	checkBuildEditRestart(t, s, command, applied, removed)
}

func calculateBeforeEquipment(t *testing.T, s *groupScenario) application.SnapshotView {
	t.Helper()
	snapshot, err := s.client.Calculate(s.ctx, application.CalculateBuild{SessionID: "main", ExpectedRevision: s.revision, RequestID: "before-armour"})
	if err != nil {
		t.Fatal(err)
	}
	s.revision = snapshot.Revision
	s.expectEvent(s.revision, "calculation")
	return supportSnapshot(t, s)
}

func compareEquipment(t *testing.T, s *groupScenario, edit *build.Edit) application.BuildComparison {
	t.Helper()
	command := application.CompareBuild{SessionID: "main", ExpectedRevision: s.revision, Edit: edit}
	preview, err := s.client.Compare(s.ctx, command)
	if err != nil {
		t.Fatal(err)
	}
	fromCLI := compareEquipmentCLI(t, s, edit)
	answer := call(t, s.ctx, s.mcp, "compare_build", map[string]any{"expected_revision": s.revision, "edit": edit}, false)
	fromMCP := structuredCatalogJSON[application.BuildComparison](t, answer.StructuredContent)
	if !reflect.DeepEqual(preview, fromCLI) || !reflect.DeepEqual(preview, fromMCP) {
		t.Fatal("ответы gRPC, CLI и MCP при примерке различаются")
	}
	checkArmourReference(t, preview.Comparison)
	return preview
}

func checkArmourReference(t *testing.T, comparison calculation.Comparison) {
	t.Helper()
	for name, values := range map[string][2]float64{
		"Life": {321, 353}, "Armour": {45, 115}, "FireResist": {10, 30},
		"TotalDPS": {2.3287313871353, 2.3287313871353},
	} {
		if math.Abs(comparison.Before.Stats[name]-values[0]) > 1e-8 || math.Abs(comparison.After.Stats[name]-values[1]) > 1e-8 {
			t.Fatalf("%s отличается от Lua-эталона: %.12g → %.12g", name, comparison.Before.Stats[name], comparison.After.Stats[name])
		}
	}
	warning := build.Issue{Code: "pob_warning", Message: "You do not meet the Strength requirement of Release Ward, Iron Cuirass"}
	if len(comparison.Before.Issues) != 0 || !slices.Equal(comparison.After.Issues, []build.Issue{warning}) {
		t.Fatalf("диагностика отличается от Lua-эталона: %v → %v", comparison.Before.Issues, comparison.After.Issues)
	}
}

func compareEquipmentCLI(t *testing.T, s *groupScenario, edit *build.Edit) application.BuildComparison {
	t.Helper()
	args := []string{"build", "compare", "--session", "main", "--revision", fmt.Sprint(s.revision),
		"--edit", "--file", equipmentEditFile(t, edit)}
	command := exec.CommandContext(s.ctx, binary, append(args, s.connection...)...)
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()
	// У эталонного предмета есть предупреждение о требовании силы.
	if command.ProcessState == nil || command.ProcessState.ExitCode() != 4 {
		t.Fatalf("CLI: %v, %s", err, &stderr)
	}
	return decodeCatalogJSON[application.BuildComparison](t, stdout.Bytes())
}

func equipmentEditFile(t *testing.T, edit *build.Edit) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "equipment.json")
	value, err := json.Marshal(edit)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, value, 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func checkRejectedEquipment(t *testing.T, s *groupScenario) {
	t.Helper()
	for _, example := range []struct{ slot, raw, hint string }{
		{"Body Armour", "Rarity: NORMAL\nMissing reference base", "не распознал базу"},
		{"Body Armour", "Rarity: NORMAL\nIron Ring", "несовместим со слотом"},
		{"Missing test slot", "Rarity: NORMAL\nIron Ring", "не распознал слот"},
	} {
		edit := &build.Edit{Equipment: &build.EquipmentEdit{Slot: example.slot, Raw: example.raw}}
		preview := application.CompareBuild{SessionID: "main", ExpectedRevision: s.revision, Edit: edit}
		if _, err := s.client.Compare(s.ctx, preview); status.Code(err) != codes.InvalidArgument || !strings.Contains(err.Error(), example.hint) {
			t.Fatalf("неверная ошибка примерки: %v", err)
		}
		command := application.ApplyBuild{SessionID: "main", ExpectedRevision: s.revision, RequestID: "reject-armour", Edit: edit}
		if _, err := s.client.Apply(s.ctx, command); status.Code(err) != codes.InvalidArgument || !strings.Contains(err.Error(), example.hint) {
			t.Fatalf("неверная ошибка применения: %v", err)
		}
	}
	checkEquipmentWarnings(t, s)
}

func checkEquipmentWarnings(t *testing.T, s *groupScenario) {
	t.Helper()
	const unsupported = "Unknown modifier for the equipment test"
	edit := &build.Edit{Equipment: &build.EquipmentEdit{Slot: "Body Armour", Raw: armourText + "\n" + unsupported}}
	preview, err := s.client.Compare(s.ctx, application.CompareBuild{SessionID: "main", ExpectedRevision: s.revision, Edit: edit})
	if err != nil {
		t.Fatal(err)
	}
	after := preview.Comparison.After
	if after.Status != "partial" || !slices.Contains(after.Issues, build.Issue{
		Code: "unsupported_modifier", Path: "Items/Body Armour", Message: unsupported,
	}) {
		t.Fatalf("пропущено неизвестное свойство предмета: %v", after.Issues)
	}
}

func applyPreviewedEquipment(t *testing.T, s *groupScenario, cmd application.ApplyBuild, before application.SnapshotView, preview application.BuildComparison) application.SnapshotView {
	t.Helper()
	args := []string{"build", "edit", "--session", "main", "--revision", fmt.Sprint(cmd.ExpectedRevision),
		"--request-id", cmd.RequestID, "--file", equipmentEditFile(t, cmd.Edit)}
	applied := buildView(t, cli(t, s.ctx, 0, append(args, s.connection...)...))
	s.expectEvent(s.revision+1, "build.model.slots")
	if applied.Build.InputSHA256 != preview.Comparison.After.InputSHA256 || applied.Build.Calculation != nil {
		t.Fatal("правка отличается от предпросмотра или оставила прежний расчёт")
	}
	checkChangedArmour(t, before.Build, applied.Build)
	cmd.RequestID = "stale-armour"
	if _, err := s.client.Apply(s.ctx, cmd); status.Code(err) != codes.Aborted {
		t.Fatal("устаревшая правка экипировки принята", err)
	}
	return applied
}

func checkChangedArmour(t *testing.T, before, after *application.BuildView) {
	t.Helper()
	old, updated := before.Model, after.Model
	if len(updated.Items) != len(old.Items)+1 || !reflect.DeepEqual(old.Items, updated.Items[:len(old.Items)]) {
		t.Fatal("замена предмета изменила прежние записи")
	}
	for slot, id := range old.Slots {
		if slot == "Body Armour" {
			id = updated.Items[len(old.Items)].ID
		}
		if updated.Slots[slot] != id {
			t.Fatalf("изменён неожиданный слот %s", slot)
		}
	}
	const set = `<ItemSet id="2"`
	_, tail, hasSet := strings.Cut(before.XML, set)
	content, _, hasEnd := strings.Cut(tail, "</ItemSet>")
	if !hasSet || !hasEnd || !strings.Contains(after.XML, set+content+"</ItemSet>") {
		t.Fatal("правка изменила неактивный набор экипировки")
	}
}

func removeEquipment(t *testing.T, s *groupScenario, applied application.SnapshotView) application.SnapshotView {
	t.Helper()
	answer := call(t, s.ctx, s.mcp, "edit_build", map[string]any{
		"expected_revision": applied.Revision, "request_id": "remove-armour",
		"edit": build.Edit{Equipment: &build.EquipmentEdit{Slot: "Body Armour"}},
	}, false)
	removed := structuredView(t, answer.StructuredContent)
	s.expectEvent(applied.Revision+1, "build.model.slots")
	if removed.Build.Model.Slots["Body Armour"] != 0 || !reflect.DeepEqual(removed.Build.Model.Items, applied.Build.Model.Items) {
		t.Fatal("снятие не очистило слот или удалило предмет из XML")
	}
	return removed
}
