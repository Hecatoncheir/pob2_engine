package build

import (
	"bytes"
	"testing"

	"github.com/Hecatoncheir/pob2_engine/internal/gamedata"
)

func TestEditsPreserveOriginalAndUnknownContent(t *testing.T) {
	xml := bytes.Replace(fixture(t), []byte(`</PathOfBuilding2>`), []byte(`<Future value="kept"/></PathOfBuilding2>`), 1)
	doc, err := Parse(xml)
	if err != nil {
		t.Fatal(err)
	}
	before, _ := doc.Export()
	data, err := gamedata.Embedded()
	if err != nil {
		t.Fatal(err)
	}
	changed, err := doc.Edited(Edit{Config: &ConfigEdit{Name: "enemyFireResist", Value: &ConfigValue{Type: "number", Value: "25"}}}, data)
	if err != nil {
		t.Fatal(err)
	}
	model, _ := changed.Model()
	if model.Config["enemyFireResist"].Value != "25" {
		t.Fatal("условие не изменено")
	}
	after, _ := changed.Export()
	if !bytes.Contains(after, []byte(`<Future value="kept"></Future>`)) {
		t.Fatal("потеряно неизвестное содержимое")
	}
	original, _ := doc.Export()
	if !bytes.Equal(before, original) {
		t.Fatal("изменён исходный документ")
	}
	_, err = doc.Edited(Edit{Config: &ConfigEdit{Name: "bad", Value: &ConfigValue{Type: "boolean", Value: "not-bool"}}}, data)
	if err == nil {
		t.Fatal("неверная правка принята")
	}
	original, _ = doc.Export()
	if !bytes.Equal(before, original) {
		t.Fatal("ошибка оставила частичную правку")
	}
}

func TestGemsEquipmentPassivesAndSelection(t *testing.T) {
	doc, err := Parse(fixture(t))
	if err != nil {
		t.Fatal(err)
	}
	data, err := gamedata.Embedded()
	if err != nil {
		t.Fatal(err)
	}
	for _, edit := range []Edit{
		{Level: &LevelEdit{Level: 90}},
		{Gems: &GemsEdit{GroupID: 1, Enabled: true, Gems: []GemInput{{ID: "FireballPlayer", Level: 15, Enabled: true}, {ID: "SupportRapidCastingPlayer", Level: 1, Enabled: true}}}},
		{Equipment: &EquipmentEdit{Slot: "Ring 1", Raw: "Rarity: Normal\nLazuli Ring\nImplicits: 1\n+25 to maximum Mana"}},
		{Passives: &PassivesEdit{Nodes: []int{4739, 18845}}},
		{Selection: &SelectionEdit{WeaponSet: 2, MainGroup: 1}},
	} {
		doc, err = doc.Edited(edit, data)
		if err != nil {
			t.Fatal(err)
		}
	}
	m, err := doc.Model()
	if err != nil {
		t.Fatal(err)
	}
	if m.Level != 90 || m.WeaponSet != 2 || m.Slots["Ring 1"] != 1 || len(m.Items) != 1 || len(m.Nodes) != 3 || len(m.Groups[0].Gems) != 2 {
		t.Fatalf("неверная модель: %+v", m)
	}
	xml, err := doc.Export()
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(xml, []byte(`<PlayerStat`)) {
		t.Fatal("не очищен старый числовой кэш")
	}
	if _, err = Parse(xml); err != nil {
		t.Fatal(err)
	}
}
