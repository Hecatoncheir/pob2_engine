package testreference

import (
	"context"
	"math"
	"testing"

	"github.com/Hecatoncheir/pob2_engine/internal/build"
)

func BenchmarkCalculateSpells(b *testing.B) {
	engine := testEngine(b)
	for _, sample := range []struct {
		name     string
		document *build.Document
		dps      float64
	}{
		{"Fireball", testDocument(b), 8.9166666666667},
		{"Frostbolt", spellDocument(b, "frostbolt"), 9.1},
		{"Spark", spellDocument(b, "spark"), 8.5642857142857},
	} {
		b.Run(sample.name, func(b *testing.B) { benchmarkSpellCalculation(b, engine, sample.document, sample.dps) })
	}
}

func benchmarkSpellCalculation(b *testing.B, engine *Engine, document *build.Document, dps float64) {
	ctx := b.Context()
	result, err := engine.Calculate(ctx, document)
	if err != nil {
		b.Fatal(err)
	}
	if result.Status != "complete" || math.Abs(result.Stats["TotalDPS"]-dps) > 1e-9 {
		b.Fatalf("неверный исходный расчёт: %+v", result)
	}
	// b.Loop исключает подготовку и проверку исходного результата из времени.
	b.ReportAllocs()
	for b.Loop() {
		checkBenchmarkCalculation(b, ctx, engine, document)
	}
}

func checkBenchmarkCalculation(b *testing.B, ctx context.Context, engine *Engine, document *build.Document) {
	result, err := engine.Calculate(ctx, document)
	if err != nil {
		b.Fatal(err)
	}
	if result.Status != "complete" {
		b.Fatalf("неполный расчёт в измерении: %+v", result.Issues)
	}
}
