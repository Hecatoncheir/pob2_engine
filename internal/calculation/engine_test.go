package calculation

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/Hecatoncheir/pob2_engine/internal/build"
	"github.com/Hecatoncheir/pob2_engine/internal/gamedata"
)

func testEngine(t testing.TB) *Engine {
	t.Helper()
	data, err := gamedata.Embedded()
	if err != nil {
		t.Fatal(err)
	}
	engine, err := New(data)
	if err != nil {
		t.Fatal(err)
	}
	return engine
}

func testDocument(t testing.TB) *build.Document {
	t.Helper()
	data, err := os.ReadFile("../../testdata/reference/fireball.xml")
	if err != nil {
		t.Fatal(err)
	}
	doc, err := build.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	return doc
}

func TestLuaMatrix(t *testing.T) {
	data, err := os.ReadFile("../../testdata/reference/cases.lua.json")
	if err != nil {
		t.Fatal(err)
	}
	var golden struct {
		Cases []struct {
			ID     string `json:"id"`
			Report struct {
				Stats      map[string]float64 `json:"stats"`
				Warnings   []string           `json:"warnings"`
				CustomMods []struct {
					Unparsed []json.RawMessage `json:"unparsed"`
				} `json:"custom_mods"`
			} `json:"report"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(data, &golden); err != nil {
		t.Fatal(err)
	}
	engine := testEngine(t)
	for _, caseData := range golden.Cases {
		t.Run(caseData.ID, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join("../../testdata/reference/cases", caseData.ID+".xml"))
			if err != nil {
				t.Fatal(err)
			}
			doc, err := build.Parse(data)
			if err != nil {
				t.Fatal(err)
			}
			actual, err := engine.Calculate(context.Background(), doc)
			if err != nil {
				t.Fatal(err)
			}
			unparsed := 0
			for _, block := range caseData.Report.CustomMods {
				unparsed += len(block.Unparsed)
			}
			if unparsed > 0 {
				if actual.Status != "unsupported" || actual.Stats != nil {
					t.Fatal("Go скрыл неразобранное свойство Lua")
				}
				return
			}
			expectedStatus := "complete"
			if len(caseData.Report.Warnings) > 0 {
				expectedStatus = "partial"
			}
			if actual.Status != expectedStatus {
				t.Fatalf("неполный расчёт: %+v", actual.Issues)
			}
			for name, want := range caseData.Report.Stats {
				got, ok := actual.Stats[name]
				if !ok || math.Abs(got-want) > math.Max(1e-9, math.Abs(want)*1e-10) {
					t.Errorf("%s: Go %.15g, Lua %.15g", name, got, want)
				}
			}
			for _, warning := range caseData.Report.Warnings {
				if !strings.Contains(warning, "Intelligence requirement of Fireball") {
					t.Fatalf("новая причина предупреждения Lua: %s", warning)
				}
				found := false
				for _, issue := range actual.Issues {
					found = found || issue.Code == "requirements_not_met"
				}
				if !found {
					t.Fatal("Go потерял предупреждение о требованиях камня")
				}
			}
		})
	}
}

func TestConcurrentComparisonKeepsInputs(t *testing.T) {
	e, doc := testEngine(t), testDocument(t)
	xml, _ := doc.Export()
	modified := bytes.Replace(xml, []byte(`</ConfigSet>`), []byte(`<Input name="customMods" string="100% increased Spell Damage"/></ConfigSet>`), 1)
	other, err := build.Parse(modified)
	if err != nil {
		t.Fatal(err)
	}
	var group sync.WaitGroup
	for range 8 {
		group.Go(func() {
			result, err := e.Compare(context.Background(), doc, other)
			if err != nil {
				t.Error(err)
				return
			}
			if result.Before.Stats["TotalDPS"] >= result.After.Stats["TotalDPS"] || len(result.Deltas) == 0 {
				t.Error("неверное сравнение")
			}
		})
	}
	group.Wait()
	after, _ := doc.Export()
	if !bytes.Equal(xml, after) {
		t.Fatal("сравнение изменило исходный XML")
	}
}

func TestUnknownMechanicDoesNotBecomeZero(t *testing.T) {
	doc := testDocument(t)
	xml, err := doc.Export()
	if err != nil {
		t.Fatal(err)
	}
	for _, replacement := range []string{`<Input name="customMods" string="unknown modifier"/>`, `<Input name="conditionUnknown" boolean="true"/>`} {
		modified := bytes.Replace(xml, []byte(`</ConfigSet>`), []byte(replacement+`</ConfigSet>`), 1)
		parsed, err := build.Parse(modified)
		if err != nil {
			t.Fatal(err)
		}
		result, err := testEngine(t).Calculate(context.Background(), parsed)
		if err != nil || result.Status != "unsupported" || result.Stats != nil || len(result.Issues) == 0 {
			t.Fatalf("неизвестная механика скрыта: %+v %v", result, err)
		}
	}
}

func TestCancellationAndInputBinding(t *testing.T) {
	e, doc := testEngine(t), testDocument(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := e.Calculate(ctx, doc); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	before, _ := doc.InputSHA256()
	result, err := e.Calculate(context.Background(), doc)
	if err != nil {
		t.Fatal(err)
	}
	after, _ := doc.InputSHA256()
	if before != after || result.InputSHA256 != before || result.DataID == "" || result.Stats["TotalDPS"] <= 0 {
		t.Fatalf("неверная привязка: %+v", result)
	}
}

func TestCachedPoBNumbersAreNotCalculationInputs(t *testing.T) {
	e, doc := testEngine(t), testDocument(t)
	before, err := e.Calculate(context.Background(), doc)
	if err != nil {
		t.Fatal(err)
	}
	xml, _ := doc.Export()
	xml = bytes.Replace(xml, []byte(`</Build>`), []byte(`<PlayerStat stat="TotalDPS" value="999999999"/></Build>`), 1)
	poisoned, err := build.Parse(xml)
	if err != nil {
		t.Fatal(err)
	}
	actual, err := e.Calculate(context.Background(), poisoned)
	if err != nil || actual.Stats["TotalDPS"] != before.Stats["TotalDPS"] {
		t.Fatalf("прочитан кэш PoB: %+v %v", actual, err)
	}
}

func TestUnsupportedSelectionHasNoNumbers(t *testing.T) {
	e, doc := testEngine(t), testDocument(t)
	xml, _ := doc.Export()
	support := `<Gem skillId="SupportRapidCastingPlayer" nameSpec="Rapid Casting I"/>`
	cases := map[string][]byte{
		"duplicate support": bytes.Replace(xml, []byte(`</Skill>`), []byte(support+support+`</Skill>`), 1),
		"missing main":      bytes.Replace(xml, []byte(`mainSocketGroup="1"`), []byte(`mainSocketGroup="99"`), 1),
		"other mode":        bytes.Replace(xml, []byte(`statSetIndex="nil"`), []byte(`statSetIndex="2"`), 1),
		"unknown passive":   bytes.Replace(xml, []byte(`nodes="54447"`), []byte(`nodes="54447,999999"`), 1),
	}
	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			parsed, err := build.Parse(data)
			if err != nil {
				t.Fatal(err)
			}
			result, err := e.Calculate(context.Background(), parsed)
			if err != nil || result.Status != "unsupported" || result.Stats != nil {
				t.Fatalf("неверный результат: %+v %v", result, err)
			}
			comparison, err := e.Compare(context.Background(), doc, parsed)
			if err != nil || len(comparison.Deltas) != 0 {
				t.Fatal("неподдержанное значение сравнивается как ноль")
			}
		})
	}
}

func BenchmarkCalculate(b *testing.B) {
	e, doc := testEngine(b), testDocument(b)
	b.ReportAllocs()
	for b.Loop() {
		if _, err := e.Calculate(context.Background(), doc); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCalculateParallel(b *testing.B) {
	e, doc := testEngine(b), testDocument(b)
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if _, err := e.Calculate(context.Background(), doc); err != nil {
				b.Error(err)
				return
			}
		}
	})
}
