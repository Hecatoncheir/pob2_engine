package calculation

import (
	"bytes"
	"os"
	"testing"

	"github.com/Hecatoncheir/pob2_engine/internal/build"
)

func classXML(t testing.TB, name string) []byte {
	t.Helper()
	xml, err := os.ReadFile("../../testdata/reference/0.0.2/classes/cases/" + name + ".xml")
	if err != nil {
		t.Fatal(err)
	}
	return xml
}

func TestClassLimitsDoNotBecomeNumericResults(t *testing.T) {
	witch := classXML(t, "witch-fireball-level-20")
	cases := map[string][]byte{
		"unknown class":           bytes.Replace(witch, []byte(`classId="1"`), []byte(`classId="99"`), 1),
		"name and id mismatch":    bytes.Replace(witch, []byte(`className="Witch"`), []byte(`className="Ranger"`), 1),
		"ascendancy id":           bytes.Replace(witch, []byte(`ascendClassId="0"`), []byte(`ascendClassId="1"`), 1),
		"ascendancy name":         bytes.Replace(witch, []byte(`ascendClassName="None"`), []byte(`ascendClassName="Blood Mage"`), 1),
		"internal ascendancy":     bytes.Replace(witch, []byte(`ascendancyInternalId=""`), []byte(`ascendancyInternalId="BloodMage"`), 1),
		"internal class mismatch": bytes.Replace(witch, []byte(`classInternalId="1"`), []byte(`classInternalId="2"`), 1),
		"shared start does not grant Sorceress passives": bytes.Replace(witch, []byte(`nodes="54447"`), []byte(`nodes="54447,4739"`), 1),
		"unknown passive": bytes.Replace(witch, []byte(`nodes="54447"`), []byte(`nodes="54447,999999"`), 1),
		"wrong start":     bytes.Replace(witch, []byte(`nodes="54447"`), []byte(`nodes="50459"`), 1),
	}
	e := testEngine(t)
	for name, xml := range cases {
		t.Run(name, func(t *testing.T) { checkClassRejected(t, e, xml) })
	}
}

func checkClassRejected(t *testing.T, engine *Engine, xml []byte) {
	t.Helper()
	doc, err := build.Parse(xml)
	if err != nil {
		t.Fatal(err)
	}
	result, err := engine.Calculate(t.Context(), doc)
	if err != nil || result.Status != "unsupported" || result.Stats != nil || len(result.Issues) == 0 {
		t.Fatalf("неподдержанный класс/дерево дал числа: %+v %v", result, err)
	}
}

func TestDisabledSkillExtensionsDoNotAffectDamage(t *testing.T) {
	xml, err := os.ReadFile("../../testdata/reference/0.0.2/groups/cases/druid-group-1.xml")
	if err != nil {
		t.Fatal(err)
	}
	e := testEngine(t)
	cases := map[string][]byte{
		"group": bytes.Replace(xml, []byte(`enabled="false"`), []byte(`enabled="false" future="kept"`), 1),
		"gem":   bytes.Replace(xml, []byte(`</Skill>`), []byte(`<Gem enabled="false" future="kept" skillId="UnknownPlayer" nameSpec="Unknown"/></Skill>`), 1),
	}
	for name, modified := range cases {
		t.Run(name, func(t *testing.T) { checkInactiveSkillExtension(t, e, xml, modified) })
	}
}

func checkInactiveSkillExtension(t *testing.T, engine *Engine, original, modified []byte) {
	t.Helper()
	if bytes.Equal(original, modified) {
		t.Fatal("не добавлено неактивное расширение")
	}
	doc, err := build.Parse(modified)
	if err != nil {
		t.Fatal(err)
	}
	result, err := engine.Calculate(t.Context(), doc)
	if err != nil || result.Status != "partial" || result.Stats["TotalDPS"] != 8.916666666666667 || len(result.Issues) != 1 || result.Issues[0].Code != "retained_inactive_xml" {
		t.Fatalf("неактивное расширение исказило результат: %+v %v", result, err)
	}
	exported, err := doc.Export()
	if err != nil || !bytes.Contains(exported, []byte(`future="kept"`)) {
		t.Fatal("потерялось неактивное расширение")
	}
	active := bytes.Replace(modified, []byte(`enabled="false" future="kept"`), []byte(`enabled="true" future="kept"`), 1)
	checkClassRejected(t, engine, active)
}

func TestInactiveConfigDoesNotApplyUnsupportedModifiers(t *testing.T) {
	xml, err := os.ReadFile("../../testdata/reference/0.0.2/groups/cases/druid-group-1.xml")
	if err != nil {
		t.Fatal(err)
	}
	xml = bytes.Replace(xml, []byte(`string="100% increased Spell Damage"`), []byte(`string="unknown modifier"`), 1)
	doc, err := build.Parse(xml)
	if err != nil {
		t.Fatal(err)
	}
	e := testEngine(t)
	result, err := e.Calculate(t.Context(), doc)
	if err != nil || result.Status != "complete" || result.Stats["TotalDPS"] != 8.916666666666667 {
		t.Fatalf("неактивное условие повлияло на результат: %+v %v", result, err)
	}
	selected, err := doc.Edited(build.Edit{Selection: &build.SelectionEdit{Sets: &build.ActiveSets{Config: 2}}}, e.data)
	if err != nil {
		t.Fatal(err)
	}
	result, err = e.Calculate(t.Context(), selected)
	if err != nil || result.Status != "unsupported" || result.Stats != nil {
		t.Fatalf("активированное неподдержанное условие дало числа: %+v %v", result, err)
	}
}
