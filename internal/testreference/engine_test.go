package testreference

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
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
	testLuaMatrix(t, "../../testdata/reference", "cases.lua.json", 77)
}

func TestSpellLuaMatrix(t *testing.T) {
	testLuaMatrix(t, "../../testdata/reference/0.0.2", "spells.lua.json", 84)
}

func TestClassLuaMatrix(t *testing.T) {
	testLuaMatrix(t, "../../testdata/reference/0.0.2/classes", "classes.lua.json", 168)
}

func TestGroupLuaMatrix(t *testing.T) {
	testLuaMatrix(t, "../../testdata/reference/0.0.2/groups", "groups.lua.json", 5)
}

type luaReport struct {
	Stats      map[string]float64 `json:"stats"`
	Warnings   []string           `json:"warnings"`
	CustomMods []struct {
		Unparsed []json.RawMessage `json:"unparsed"`
	} `json:"custom_mods"`
}

type luaCase struct {
	ID          string    `json:"id"`
	Skill       string    `json:"skill"`
	ClassID     int       `json:"class_id"`
	ClassName   string    `json:"class_name"`
	StartNode   int       `json:"start_node"`
	InputSHA256 string    `json:"input_sha256"`
	Report      luaReport `json:"report"`
}

func testLuaMatrix(t *testing.T, directory, name string, count int) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(directory, name))
	if err != nil {
		t.Fatal(err)
	}
	var golden struct {
		PoBCommit string    `json:"pob_commit"`
		Cases     []luaCase `json:"cases"`
	}
	if err := json.Unmarshal(data, &golden); err != nil {
		t.Fatal(err)
	}
	engine := testEngine(t)
	if len(golden.Cases) != count || golden.PoBCommit != engine.data.PoBCommit() {
		t.Fatal("изменён состав матрицы либо происхождение данных")
	}
	for _, sample := range golden.Cases {
		t.Run(sample.ID, func(t *testing.T) {
			actual := calculateLuaCase(t, engine, directory, sample)
			checkLuaResult(t, actual, sample)
		})
	}
}

func calculateLuaCase(t *testing.T, engine *Engine, directory string, sample luaCase) Result {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(directory, "cases", sample.ID+".xml"))
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(data)
	if hex.EncodeToString(digest[:]) != sample.InputSHA256 {
		t.Fatal("XML не соответствует сумме эталона")
	}
	doc, err := build.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	if sample.ClassID != 0 {
		checkClassInput(t, doc, sample)
	}
	actual, err := engine.Calculate(t.Context(), doc)
	if err != nil {
		t.Fatal(err)
	}
	return actual
}

func checkClassInput(t *testing.T, doc *build.Document, sample luaCase) {
	t.Helper()
	model, err := doc.Model()
	if err != nil || model.ClassID != sample.ClassID || model.Class != sample.ClassName || len(model.Nodes) != 1 || model.Nodes[0] != sample.StartNode {
		t.Fatalf("эталон использует другой класс или старт: %+v %v", model, err)
	}
}

func checkLuaResult(t *testing.T, actual Result, sample luaCase) {
	t.Helper()
	skill, mode := sample.Skill, "Projectile hits"
	if skill == "" {
		skill = "Fireball"
	}
	if skill == "Spark" {
		mode = "Base hits"
	}
	if actual.Skill != skill || actual.Mode != mode {
		t.Fatalf("подменено умение или режим: %s / %s", actual.Skill, actual.Mode)
	}
	unparsed := 0
	for _, block := range sample.Report.CustomMods {
		unparsed += len(block.Unparsed)
	}
	if unparsed > 0 {
		checkUnparsedResult(t, actual, unparsed)
		return
	}
	expectedStatus := "complete"
	if len(sample.Report.Warnings) > 0 {
		expectedStatus = "partial"
	}
	if actual.Status != expectedStatus {
		t.Fatalf("неполный расчёт: %+v", actual.Issues)
	}
	for name, want := range sample.Report.Stats {
		got, ok := actual.Stats[name]
		if !ok || math.Abs(got-want) > math.Max(1e-9, math.Abs(want)*1e-10) {
			t.Errorf("%s: Go %.15g, Lua %.15g", name, got, want)
		}
	}
	checkLuaWarnings(t, actual, sample.Report.Warnings, skill)
}

func checkUnparsedResult(t *testing.T, actual Result, count int) {
	t.Helper()
	if actual.Status != "unsupported" || actual.Stats != nil || len(actual.Issues) != count {
		t.Fatal("Go скрыл неразобранное свойство Lua")
	}
	for _, issue := range actual.Issues {
		if issue.Code != "unsupported_modifier" || !strings.HasPrefix(issue.Path, "Config/customMods/") {
			t.Fatalf("неподдержанный результат вызван другой причиной: %+v", issue)
		}
	}
}

func checkLuaWarnings(t *testing.T, actual Result, warnings []string, skill string) {
	t.Helper()
	if len(actual.Issues) != len(warnings) {
		t.Fatalf("не совпало число предупреждений: %+v", actual.Issues)
	}
	for _, warning := range warnings {
		if !strings.Contains(warning, "Intelligence requirement of "+skill) {
			t.Fatalf("новая причина предупреждения Lua: %s", warning)
		}
	}
	for _, issue := range actual.Issues {
		if issue.Code != "requirements_not_met" || issue.Path != "Skills/"+skill {
			t.Fatalf("Go подменил предупреждение о требованиях: %+v", issue)
		}
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
