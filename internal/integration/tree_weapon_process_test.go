package integration_test

import (
	"maps"
	"math"
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

const weaponTarget = 51184

func TestLuaTreeWeaponModesAcrossProcesses(t *testing.T) {
	if os.Getenv("POB2_ENGINE_POB_DIR") == "" {
		t.Skip("для проверки Lua укажите POB2_ENGINE_POB_DIR")
	}
	t.Setenv("POB2_ENGINE_CALCULATOR", "lua")
	s := newBuildScenarioWithTimeout(t, "0.0.12/witch-weapon-paths.xml", 90*time.Second)
	info, err := s.client.Info(s.ctx)
	if err != nil || !slices.Contains(info.GetCapabilities(), "build.tree_allocation_modes") {
		t.Fatal("движок не объявил режимы назначения", err)
	}
	before := calculateRecordedBuild(t, s, "weapon-one-baseline")
	checkWeaponTreeMatrix(t, s, before, 1)
	second := beginSecondWeaponSet(t, s)
	preview := checkWeaponTreeMatrix(t, s, second, 2)
	checkWeaponTreeRejections(t, s)
	command := application.ApplyBuild{SessionID: "main", ExpectedRevision: s.revision,
		RequestID: "allocate-weapon-two", Edit: weaponTreeEdit(weaponTarget, "allocate", build.AllocationWeaponSet2)}
	applied := applyTreeNodeCLI(t, s, command, preview)
	checkWeaponTreeXML(t, second, applied)
	restored := refundWeaponPath(t, s, second)
	checkBuildEditRestart(t, s, command, applied, restored)
}

func weaponTreeEdit(node int, action string, mode build.TreeAllocationMode) *build.Edit {
	return &build.Edit{TreeNode: &build.TreeNodeEdit{NodeID: node, Action: action, AllocationMode: mode}}
}

type weaponTreeCase struct {
	mode       build.TreeAllocationMode
	added      []int
	moved      *build.PassiveAllocationChange
	base       [2]float64
	dps        [2]float64
	used       build.PassivePoints
	diagnostic string
}

func checkWeaponTreeMatrix(t *testing.T, s *groupScenario, before application.SnapshotView, weapon int) application.BuildComparison {
	t.Helper()
	if !maps.Equal(before.Build.Model.PassiveAllocations, map[string]build.TreeAllocationMode{
		"44871": build.AllocationWeaponSet1, "4739": build.AllocationWeaponSet2,
	}) {
		t.Fatal("импортированное распределение узлов потеряно")
	}
	base := [2]float64{15.38125, 19.482916666667}
	cases := []weaponTreeCase{
		{mode: build.AllocationCommon, added: []int{1755, 18845, 41965, weaponTarget},
			moved: &build.PassiveAllocationChange{NodeID: 4739, Before: build.AllocationWeaponSet2, After: build.AllocationCommon},
			base:  base, dps: [2]float64{22.046458333333, 24.61}, used: build.PassivePoints{Passives: 6, WeaponSet1: 1},
			diagnostic: "You have 1 Weapon set 2 passives available"},
		{mode: build.AllocationWeaponSet1, added: []int{2254, 3242, 13769, 29695, 30346, 43736, weaponTarget, 59636},
			base: base, dps: [2]float64{17.432083333333, 19.482916666667}, used: build.PassivePoints{Passives: 9, WeaponSet1: 9, WeaponSet2: 1},
			diagnostic: "You have 8 Weapon set 2 passives available"},
		{mode: build.AllocationWeaponSet2, added: []int{1755, 18845, 41965, weaponTarget},
			base: base, dps: [2]float64{15.38125, 24.61}, used: build.PassivePoints{Passives: 5, WeaponSet1: 1, WeaponSet2: 5},
			diagnostic: "You have 4 Weapon set 2 passives available"},
	}
	var preview application.BuildComparison
	for _, want := range cases {
		preview = compareTreeNodeStatus(t, s, weaponTreeEdit(weaponTarget, "allocate", want.mode), 4)
		checkWeaponTreePreview(t, preview, want, weapon)
	}
	if !reflect.DeepEqual(supportSnapshot(t, s), before) {
		t.Fatal("предпросмотр ветви изменил сессию")
	}
	return preview
}

func checkWeaponTreePreview(t *testing.T, preview application.BuildComparison, want weaponTreeCase, weapon int) {
	t.Helper()
	allocations := make([]build.PassiveAllocationChange, 0, len(want.added)+1)
	for _, id := range want.added {
		allocations = append(allocations, build.PassiveAllocationChange{NodeID: id, After: want.mode})
	}
	if want.moved != nil {
		allocations = slices.Insert(allocations, 1, *want.moved)
	}
	change := preview.Passives
	points := build.PassivePointsChange{Before: build.PassivePoints{Passives: 1, WeaponSet1: 1, WeaponSet2: 1}, After: want.used}
	if change == nil || !slices.Equal(change.Added, want.added) || len(change.Removed) != 0 ||
		!slices.Equal(change.Allocations, allocations) || change.Points == nil || *change.Points != points {
		t.Fatalf("неверный путь %s: %+v", want.mode, change)
	}
	before, after := preview.Comparison.Before, preview.Comparison.After
	if math.Abs(before.Stats["TotalDPS"]-want.base[weapon-1]) > 0.00001 ||
		math.Abs(after.Stats["TotalDPS"]-want.dps[weapon-1]) > 0.00001 || after.Stats["Life"] != 1187 {
		t.Fatalf("неверный расчёт комплекта %d для режима %s: %v → %v", weapon, want.mode, before.Stats["TotalDPS"], after.Stats["TotalDPS"])
	}
	if after.Status != "partial" || !slices.ContainsFunc(after.Issues, func(issue build.Issue) bool {
		return strings.Contains(issue.Message, want.diagnostic)
	}) {
		t.Fatalf("пропущено предупреждение оригинального PoB: %v", after.Issues)
	}
}

func beginSecondWeaponSet(t *testing.T, s *groupScenario) application.SnapshotView {
	t.Helper()
	command := application.ApplyBuild{SessionID: "main", ExpectedRevision: s.revision, RequestID: "weapon-set-two",
		Edit: &build.Edit{Selection: &build.SelectionEdit{WeaponSet: 2}}}
	if _, err := s.client.Apply(s.ctx, command); err != nil {
		t.Fatal(err)
	}
	s.revision++
	s.expectEvent(s.revision, "build.model.weapon_set")
	return calculateRecordedBuild(t, s, "weapon-two-baseline")
}

func checkWeaponTreeRejections(t *testing.T, s *groupScenario) {
	t.Helper()
	for _, edit := range []*build.Edit{
		weaponTreeEdit(4739, "refund", ""),
		weaponTreeEdit(4739, "refund", build.AllocationWeaponSet1),
		weaponTreeEdit(54447, "allocate", build.AllocationWeaponSet2),
	} {
		if _, err := s.client.Compare(s.ctx, application.CompareBuild{SessionID: "main", ExpectedRevision: s.revision, Edit: edit}); status.Code(err) != codes.InvalidArgument {
			t.Fatal("принята недопустимая правка ветви", edit.TreeNode, err)
		}
	}
	call(t, s.ctx, s.mcp, "compare_build", map[string]any{
		"edit": weaponTreeEdit(weaponTarget, "allocate", "unknown"), "expected_revision": s.revision,
	}, true)
}

func checkWeaponTreeXML(t *testing.T, before, after application.SnapshotView) {
	t.Helper()
	want := strings.NewReplacer(
		`nodes="44871,54447,4739"`, `nodes="44871,54447,4739,1755,18845,41965,51184"`,
		`<WeaponSet2 nodes="4739"`, `<WeaponSet2 nodes="4739,1755,18845,41965,51184"`,
	).Replace(before.Build.XML)
	if want == before.Build.XML || want != after.Build.XML {
		t.Fatal("назначение изменило XML за пределами узлов второй ветви")
	}
}

func refundWeaponPath(t *testing.T, s *groupScenario, before application.SnapshotView) application.SnapshotView {
	t.Helper()
	edit := weaponTreeEdit(18845, "refund", build.AllocationWeaponSet2)
	preview := compareTreeNodeStatus(t, s, edit, 4)
	if preview.Passives == nil || !slices.Equal(preview.Passives.Removed, []int{1755, 18845, 41965, weaponTarget}) {
		t.Fatalf("не сняты зависимые узлы второй ветви: %+v", preview.Passives)
	}
	result := call(t, s.ctx, s.mcp, "edit_build", map[string]any{"edit": edit, "expected_revision": s.revision, "request_id": "refund-weapon-two"}, false)
	restored := structuredView(t, result.StructuredContent)
	s.revision++
	s.expectEvent(s.revision, "build.model.passive_allocations")
	if restored.Build.InputSHA256 != preview.Comparison.After.InputSHA256 || restored.Build.XML != before.Build.XML {
		t.Fatal("снятие ветви не восстановило исходный XML")
	}
	return restored
}
