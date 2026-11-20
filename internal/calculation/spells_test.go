package calculation

import (
	"bytes"
	"context"
	"os"
	"sync"
	"testing"

	"github.com/Hecatoncheir/pob2_engine/internal/build"
)

func spellDocument(t testing.TB, name string) *build.Document {
	t.Helper()
	xml, err := os.ReadFile("../../testdata/reference/0.0.2/cases/" + name + "-minimal.xml")
	if err != nil {
		t.Fatal(err)
	}
	doc, err := build.Parse(xml)
	if err != nil {
		t.Fatal(err)
	}
	return doc
}

func TestComparingDifferentSpellsKeepsEachInput(t *testing.T) {
	e := testEngine(t)
	documents := []*build.Document{testDocument(t), spellDocument(t, "frostbolt"), spellDocument(t, "spark")}
	expected := []string{"Fireball", "Frostbolt", "Spark"}
	before := make([][]byte, len(documents))
	for i, doc := range documents {
		before[i], _ = doc.Export()
	}
	var workers sync.WaitGroup
	for range 8 {
		workers.Go(func() {
			for i, doc := range documents {
				next := (i + 1) % len(documents)
				result, err := e.Compare(context.Background(), doc, documents[next])
				if err != nil || result.Before.Skill != expected[i] || result.After.Skill != expected[next] || result.Before.Status != "complete" || result.After.Status != "complete" {
					t.Errorf("сравнение загрязнило другое умение: %+v %v", result, err)
					return
				}
			}
		})
	}
	workers.Wait()
	for i, doc := range documents {
		after, _ := doc.Export()
		if !bytes.Equal(before[i], after) {
			t.Fatal("сравнение изменило исходную сборку")
		}
	}
}

func TestNewSpellUnsupportedModeAndConditionsHaveNoNumbers(t *testing.T) {
	e := testEngine(t)
	for _, name := range []string{"frostbolt", "spark"} {
		xml, err := spellDocument(t, name).Export()
		if err != nil {
			t.Fatal(err)
		}
		for _, invalid := range [][]byte{
			bytes.Replace(xml, []byte(`statSetIndex="nil"`), []byte(`statSetIndex="2"`), 1),
			bytes.Replace(xml, []byte(`quality="0"`), []byte(`quality="1"`), 1),
			bytes.Replace(xml, []byte(`</ConfigSet>`), []byte(`<Input name="conditionEnemyShocked" boolean="true"/></ConfigSet>`), 1),
		} {
			doc, err := build.Parse(invalid)
			if err != nil {
				t.Fatal(err)
			}
			actual, err := e.Calculate(t.Context(), doc)
			if err != nil || actual.Status != "unsupported" || actual.Stats != nil || len(actual.Issues) == 0 {
				t.Fatalf("неподдержанный режим %s стал числовым результатом: %+v %v", name, actual, err)
			}
		}
	}
}
