package application

import (
	"context"
	"os"
	"testing"

	"github.com/Hecatoncheir/pob2_engine/internal/domain"
)

func BenchmarkPreviewSupport(b *testing.B) {
	for _, spell := range []struct{ name, file string }{
		{"Fireball", "../../testdata/reference/fireball.xml"},
		{"Frostbolt", "../../testdata/reference/0.0.2/cases/frostbolt-minimal.xml"},
		{"Spark", "../../testdata/reference/0.0.2/cases/spark-minimal.xml"},
	} {
		b.Run(spell.name, func(b *testing.B) { benchmarkSpellSupports(b, spell.file) })
	}
}

func benchmarkSpellSupports(b *testing.B, file string) {
	for _, support := range []struct{ name, id string }{
		{"RapidCasting", "Rapid Casting I"},
		{"ControlledDestruction", "Controlled Destruction"},
	} {
		b.Run(support.name, func(b *testing.B) { benchmarkSupportPreview(b, file, support.id) })
	}
}

func benchmarkPreviewSession(b *testing.B, file string) (*Service, domain.Snapshot) {
	xml, err := os.ReadFile(file)
	if err != nil {
		b.Fatal(err)
	}
	service := NewService()
	_, err = service.Create(b.Context(), CreateSession{SessionID: "bench", BuildName: "Измерение", RequestID: "create"})
	if err != nil {
		b.Fatal(err)
	}
	snapshot, err := service.Apply(b.Context(), ApplyBuild{SessionID: "bench", ExpectedRevision: 1, RequestID: "import", XML: string(xml)})
	if err != nil {
		b.Fatal(err)
	}
	return service, snapshot
}

func benchmarkSupportPreview(b *testing.B, file, support string) {
	service, before := benchmarkPreviewSession(b, file)
	command := PreviewSupport{SessionID: before.SessionID, ExpectedRevision: before.Revision, SupportID: support}
	ctx := b.Context()
	checkBenchmarkPreview(b, ctx, service, command)
	b.ReportAllocs()
	for b.Loop() {
		checkBenchmarkPreview(b, ctx, service, command)
	}
	// После b.Loop таймер остановлен. Предпросмотр должен оставить сессию прежней.
	after, err := service.Snapshot(ctx, before.SessionID)
	if err != nil {
		b.Fatal(err)
	}
	if after != before {
		b.Fatal("предпросмотр изменил измеряемую сессию")
	}
}

func checkBenchmarkPreview(b *testing.B, ctx context.Context, service *Service, command PreviewSupport) {
	result, err := service.PreviewSupport(ctx, command)
	if err != nil {
		b.Fatal(err)
	}
	if result.Status != "complete" || !result.Compatible {
		b.Fatalf("неполный предпросмотр в измерении: %+v", result)
	}
}
