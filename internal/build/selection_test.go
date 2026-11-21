package build

import (
	"bytes"
	"os"
	"testing"

	"github.com/Hecatoncheir/pob2_engine/internal/gamedata"
)

func TestSelectingSmallerSkillSetRequiresExistingMainGroup(t *testing.T) {
	xml, err := os.ReadFile("../../testdata/reference/0.0.2/groups/cases/druid-group-3.xml")
	if err != nil {
		t.Fatal(err)
	}
	doc, err := Parse(xml)
	if err != nil {
		t.Fatal(err)
	}
	data, err := gamedata.Embedded()
	if err != nil {
		t.Fatal(err)
	}
	before, _ := doc.Export()
	if _, err := doc.Edited(Edit{Selection: &SelectionEdit{Sets: &ActiveSets{Skills: 2}}}, data); err == nil {
		t.Fatal("выбран набор без основной группы 3")
	}
	after, _ := doc.Export()
	if !bytes.Equal(before, after) {
		t.Fatal("ошибка выбора изменила исходный документ")
	}
	selected, err := doc.Edited(Edit{Selection: &SelectionEdit{MainGroup: 1, Sets: &ActiveSets{Skills: 2, Config: 2}}}, data)
	if err != nil {
		t.Fatal(err)
	}
	model, err := selected.Model()
	if err != nil || model.ActiveSets.Skills != 2 || model.ActiveSets.Config != 2 || model.MainGroup != 1 || model.Groups[0].Gems[0].SkillID != "SparkPlayer" {
		t.Fatalf("не выбран альтернативный набор: %+v %v", model, err)
	}
}
