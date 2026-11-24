package pob

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Hecatoncheir/pob2_engine/internal/build"
	"github.com/Hecatoncheir/pob2_engine/internal/calculation"
	"github.com/Hecatoncheir/pob2_engine/internal/gamedata"
)

func TestMain(m *testing.M) {
	if len(os.Args) > 1 && os.Args[1] == "_pob-lua" {
		if err := RunWorker(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}

var spells = []struct {
	name, file string
	dps        float64
}{
	{"Fireball", "fireball.xml", 8.9166666666667},
	{"Frostbolt", "0.0.2/cases/frostbolt-minimal.xml", 9.1},
	{"Spark", "0.0.2/cases/spark-minimal.xml", 8.5642857142857},
}

func localCalculator(t testing.TB) *Calculator {
	t.Helper()
	if os.Getenv("POB2_ENGINE_POB_DIR") == "" {
		t.Skip("для проверки Lua укажите POB2_ENGINE_POB_DIR")
	}
	manager, err := Local()
	if err != nil {
		t.Fatal(err)
	}
	installation, err := manager.Installed()
	if err != nil {
		t.Fatal(err)
	}
	engine, err := NewCalculator(installation)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = engine.Close() })
	return engine
}

func readBuild(t testing.TB, path string) *build.Document {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "reference", path))
	if err != nil {
		t.Fatal(err)
	}
	doc, err := build.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	return doc
}

func TestLuaCalculatesOriginalBuilds(t *testing.T) {
	engine := localCalculator(t)
	for _, spell := range spells {
		t.Run(spell.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
			defer cancel()
			result, err := engine.Calculate(ctx, readBuild(t, spell.file))
			if err != nil {
				t.Fatal(err)
			}
			if result.Skill != spell.name || math.Abs(result.Stats["TotalDPS"]-spell.dps) > 1e-8 ||
				result.Status != "complete" || len(result.Issues) != 0 {
				t.Fatalf("неожиданный расчёт: %s, DPS %.9f, статус %s, предупреждения %v",
					result.Skill, result.Stats["TotalDPS"], result.Status, result.Issues)
			}
		})
	}
}

func TestLuaReportsUnsupportedModifierAfterEdit(t *testing.T) {
	engine := localCalculator(t)
	data, err := gamedata.Embedded()
	if err != nil {
		t.Fatal(err)
	}
	const unsupported = "Unknown custom modifier for this fixture"
	value := &build.ConfigValue{Type: "string", Value: "100% increased Spell Damage\n" + unsupported}
	doc, err := readBuild(t, spells[0].file).Edited(build.Edit{Config: &build.ConfigEdit{Name: "customMods", Value: value}}, data)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	result, err := engine.Calculate(ctx, doc)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(result.Stats["TotalDPS"]-2*spells[0].dps) > 1e-8 || result.Status != "partial" {
		t.Fatalf("неожиданный расчёт после правки: %.9f, %s", result.Stats["TotalDPS"], result.Status)
	}
	for _, issue := range result.Issues {
		if issue.Code == "unsupported_modifier" && issue.Message == unsupported {
			return
		}
	}
	t.Fatalf("не получено предупреждение о модификаторе: %v", result.Issues)
}

func BenchmarkCalculators(b *testing.B) {
	lua := localCalculator(b)
	for _, spell := range spells {
		doc := readBuild(b, spell.file)
		b.Run(spell.name+"/Lua", func(b *testing.B) { benchmarkCalculation(b, lua, doc) })
	}
}

func benchmarkCalculation(b *testing.B, engine calculation.Calculator, doc *build.Document) {
	ctx, cancel := context.WithTimeout(b.Context(), 40*time.Second)
	defer cancel()
	if _, err := engine.Calculate(ctx, doc); err != nil {
		b.Fatal(err)
	}
	for b.Loop() {
		if _, err := engine.Calculate(ctx, doc); err != nil {
			b.Fatal(err)
		}
	}
}
