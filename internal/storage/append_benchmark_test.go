package storage

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/Hecatoncheir/pob2_engine/internal/application"
)

func BenchmarkPublishedCalculationAppend(b *testing.B) {
	base := b.TempDir()
	fixture := filepath.Join("..", "..", "testdata", "compat", "0.0.5", "sessions.db")
	if _, err := Restore(b.Context(), fixture, base, "measured"); err != nil {
		b.Fatal(err)
	}
	journal, err := Open(b.Context(), base, "measured")
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { _ = journal.Close() })
	var sample application.Change
	if err := journal.Load(b.Context(), func(change application.Change) error {
		if change.Calculate != nil {
			sample = change
		}
		return nil
	}); err != nil {
		b.Fatal(err)
	}
	if sample.Calculate == nil {
		b.Fatal("в эталонной базе нет сохранённого расчёта")
	}
	for b.Loop() {
		sample.Calculate.ExpectedRevision = sample.Result.Revision
		sample.Result.Revision++
		sample.Calculate.RequestID = fmt.Sprintf("measure-%d", sample.Result.Revision)
		if err := journal.Append(b.Context(), sample); err != nil {
			b.Fatal(err)
		}
	}
}
