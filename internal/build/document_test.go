package build

import (
	"bytes"
	"os"
	"reflect"
	"strings"
	"testing"
)

func fixture(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile("../../testdata/reference/fireball.xml")
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestFixtureRoundTrip(t *testing.T) {
	doc, err := Parse(fixture(t))
	if err != nil {
		t.Fatal(err)
	}
	m, err := doc.Model()
	if err != nil {
		t.Fatal(err)
	}
	if m.Class != "Sorceress" || m.ClassID != 7 || m.Groups[0].Gems[0].SkillID != "FireballPlayer" || len(m.Issues) != 0 {
		t.Fatalf("неверная модель: %+v", m)
	}
	for i := 0; i < 3; i++ {
		xml, err := doc.Export()
		if err != nil {
			t.Fatal(err)
		}
		next, err := Parse(xml)
		if err != nil {
			t.Fatal(err)
		}
		value, err := next.Model()
		if err != nil || !reflect.DeepEqual(m, value) {
			t.Fatalf("изменена модель: %+v %v", value, err)
		}
		before, _ := doc.InputSHA256()
		after, _ := next.InputSHA256()
		if before != after {
			t.Fatal("изменился хеш нормализованного XML")
		}
		doc = next
	}
}

func TestPreserveUnknownContentAndReportIt(t *testing.T) {
	data := bytes.Replace(fixture(t), []byte("</PathOfBuilding2>"), []byte(`<Future enabled="yes"><Value>kept &amp; safe</Value> <Value> second </Value></Future></PathOfBuilding2>`), 1)
	doc, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	m, _ := doc.Model()
	if len(m.Issues) == 0 || m.Issues[0].Code != "unsupported_xml" {
		t.Fatal("расширение молча проигнорировано")
	}
	exported, err := doc.Export()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(exported, []byte(`<Future enabled="yes"><Value>kept &amp; safe</Value> <Value> second </Value></Future>`)) {
		t.Fatal("расширение потерялось")
	}
}

func TestRejectMalformedOrExcessiveXML(t *testing.T) {
	cases := map[string][]byte{
		"namespace":       []byte(`<PathOfBuilding2 xmlns="urn:other"/>`),
		"DTD":             []byte(`<!DOCTYPE x><PathOfBuilding2/>`),
		"entity":          []byte(`<PathOfBuilding2>&unknown;</PathOfBuilding2>`),
		"root":            []byte(`<PathOfBuilding/>`),
		"two roots":       append(fixture(t), []byte(`<extra/>`)...),
		"large":           bytes.Repeat([]byte("a"), MaxXMLSize+1),
		"depth":           []byte(strings.Repeat("<a>", 65) + strings.Repeat("</a>", 65)),
		"level":           bytes.Replace(fixture(t), []byte(`level="1"`), []byte(`level="999"`), 1),
		"active set":      bytes.Replace(fixture(t), []byte(`activeItemSet="1"`), []byte(`activeItemSet="99"`), 1),
		"slot reference":  bytes.Replace(fixture(t), []byte(`itemId="0"`), []byte(`itemId="999"`), 1),
		"duplicate attr":  bytes.Replace(fixture(t), []byte(`className="Sorceress"`), []byte(`className="Sorceress" className="Witch"`), 1),
		"duplicate build": bytes.Replace(fixture(t), []byte(`</PathOfBuilding2>`), []byte(`<Build/></PathOfBuilding2>`), 1),
	}
	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := Parse(data); err == nil {
				t.Fatal("некорректный XML принят")
			}
		})
	}
}

func TestSelectActiveSetsAndKeepInactiveSets(t *testing.T) {
	data := bytes.Replace(fixture(t), []byte(`</Skills>`), []byte(`<SkillSet id="2"><Skill label="second"><Gem nameSpec="Other" skillId="OtherPlayer"/></Skill></SkillSet></Skills>`), 1)
	data = bytes.Replace(data, []byte(`activeSkillSet="1"`), []byte(`activeSkillSet="2"`), 1)
	doc, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	m, _ := doc.Model()
	if m.ActiveSets.Skills != 2 || m.Groups[0].Gems[0].Name != "Other" {
		t.Fatalf("не выбран второй набор: %+v", m)
	}
	xml, _ := doc.Export()
	if !bytes.Contains(xml, []byte(`skillId="FireballPlayer"`)) {
		t.Fatal("потерян неактивный набор")
	}
}

func TestInactiveExtensionsAreReportedSeparately(t *testing.T) {
	data := bytes.Replace(fixture(t), []byte(`</Skills>`), []byte(`<SkillSet id="2"><Future/></SkillSet></Skills>`), 1)
	doc, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	m, _ := doc.Model()
	if len(m.Issues) != 1 || m.Issues[0].Code != "retained_inactive_xml" {
		t.Fatalf("неверная диагностика: %+v", m.Issues)
	}
}
