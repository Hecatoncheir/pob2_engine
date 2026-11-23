package integration_test

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/Hecatoncheir/pob2_engine/internal/build"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type workshopEquipmentDetails struct {
	Revision uint64         `json:"revision"`
	Details  map[string]any `json:"details"`
}

func TestLuaEquipmentDetailsAcrossProcesses(t *testing.T) {
	if os.Getenv("POB2_ENGINE_POB_DIR") == "" {
		t.Skip("для проверки Lua укажите POB2_ENGINE_POB_DIR")
	}
	t.Setenv("POB2_ENGINE_CALCULATOR", "lua")
	s := newBuildScenario(t, "tree-specs/warrior-witch.xml")
	before := supportSnapshot(t, s)
	query := build.EquipmentDetailsQuery{Slot: "Weapon 1", ItemID: 1, Compare: true}
	view := compareEquipmentDetails(t, s, query)
	checkEquipmentDetails(t, view)
	if _, err := s.client.EquipmentDetails(s.ctx, "main", s.revision-1, query); status.Code(err) != codes.Aborted {
		t.Fatal("принята старая ревизия примерки", err)
	}
	query.ItemID, query.Raw, query.Compare = 0, "Rarity: NORMAL\nWooden Club", false
	if _, err := s.client.EquipmentDetails(s.ctx, "main", s.revision, query); err != nil {
		t.Fatal("не разобран текст предмета", err)
	}
	if !reflect.DeepEqual(supportSnapshot(t, s), before) {
		t.Fatal("описание и сравнение изменили XML или ревизию сборки")
	}
}

func compareEquipmentDetails(t *testing.T, s *groupScenario, query build.EquipmentDetailsQuery) workshopEquipmentDetails {
	t.Helper()
	reply, err := s.client.EquipmentDetails(s.ctx, "main", s.revision, query)
	if err != nil {
		t.Fatal(err)
	}
	view := workshopEquipmentDetails{Revision: reply.Revision, Details: decodeCatalogJSON[map[string]any](t, reply.Details)}
	args := []string{"equipment-details", "--session", "main", "--revision", fmt.Sprint(s.revision),
		"--slot", query.Slot, "--item-id", fmt.Sprint(query.ItemID), "--compare"}
	fromCLI := decodeCatalogJSON[workshopEquipmentDetails](t, cli(t, s.ctx, 0, append(args, s.connection...)...))
	answer := call(t, s.ctx, s.mcp, "get_equipment_details", map[string]any{
		"expected_revision": s.revision, "slot": query.Slot, "item_id": query.ItemID, "compare": query.Compare,
	}, false)
	fromMCP := structuredCatalogJSON[workshopEquipmentDetails](t, answer.StructuredContent)
	if !reflect.DeepEqual(view, fromCLI) || !reflect.DeepEqual(view, fromMCP) {
		t.Fatal("CLI, MCP и gRPC вернули разные описания или сравнения")
	}
	return view
}

func checkEquipmentDetails(t *testing.T, view workshopEquipmentDetails) {
	t.Helper()
	if view.Details["source"] != "original-lua" || view.Details["compatible"] != true ||
		!strings.Contains(fmt.Sprint(view.Details["item"]), "Wooden Club") {
		t.Fatal("Lua не подтвердил базу или совместимость предмета")
	}
	if !strings.Contains(fmt.Sprint(view.Details["comparison"]), "AverageDamage") ||
		!strings.Contains(fmt.Sprint(view.Details["lines"]), "Wooden Club") {
		t.Fatal("Lua не вернул строки характеристик и подсказки")
	}
}
