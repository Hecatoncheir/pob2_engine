package pob

import (
	"bytes"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/calculation"
	"github.com/Hecatoncheir/pob2_engine/internal/domain"
	"github.com/Hecatoncheir/pob2_engine/internal/storage"
)

func TestLuaRestoresV002Session(t *testing.T) {
	calculator := localCalculator(t)
	fixtures := filepath.Join("..", "..", "testdata", "compat", "0.0.2")
	expected, err := os.ReadFile(filepath.Join(fixtures, "snapshot.json"))
	if err != nil {
		t.Fatal(err)
	}
	base := t.TempDir()
	if _, err := storage.Restore(t.Context(), filepath.Join(fixtures, "sessions.db"), base, "upgrade"); err != nil {
		t.Fatal(err)
	}
	journal, service := openUpgradeSession(t, base, calculator)
	previous := sessionSnapshot(t, service)
	view, err := application.ViewSnapshot(previous)
	if err != nil {
		t.Fatal(err)
	}
	got, err := json.MarshalIndent(view, "", "  ")
	if err != nil || !bytes.Equal(append(got, '\n'), expected) {
		t.Fatalf("снимок v0.0.2 изменился при открытии: %v", err)
	}

	oldCommand := application.CalculateBuild{SessionID: "main", ExpectedRevision: 2, RequestID: "calculate"}
	newCommand := application.CalculateBuild{SessionID: "main", ExpectedRevision: 3, RequestID: "release-upgrade"}
	assertCalculationReplay(t, service, oldCommand, previous)
	updated, err := service.Calculate(t.Context(), newCommand)
	if err != nil {
		t.Fatal(err)
	}
	checkLuaUpgradeCalculation(t, previous, updated)

	if err := journal.Close(); err != nil {
		t.Fatal(err)
	}
	_, reopened := openUpgradeSession(t, base, calculator)
	assertCalculationReplay(t, reopened, oldCommand, previous)
	assertCalculationReplay(t, reopened, newCommand, updated)
	if sessionSnapshot(t, reopened) != updated {
		t.Fatal("повтор старого запроса изменил текущую ревизию")
	}
}

func openUpgradeSession(t *testing.T, base string, calculator *Calculator) (*storage.Journal, *application.Service) {
	t.Helper()
	journal, err := storage.Open(t.Context(), base, "upgrade")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = journal.Close() })
	service, err := application.OpenServiceWithCalculator(t.Context(), journal, calculator)
	if err != nil {
		t.Fatal(err)
	}
	return journal, service
}

func sessionSnapshot(t *testing.T, service *application.Service) domain.Snapshot {
	t.Helper()
	snapshot, err := service.Snapshot(t.Context(), "main")
	if err != nil {
		t.Fatal(err)
	}
	return snapshot
}

func assertCalculationReplay(t *testing.T, service *application.Service, command application.CalculateBuild, expected domain.Snapshot) {
	t.Helper()
	got, err := service.Calculate(t.Context(), command)
	if err != nil || got != expected {
		t.Fatalf("не восстановлен ответ для %s: ревизия %d, ошибка %v", command.RequestID, got.Revision, err)
	}
}

func checkLuaUpgradeCalculation(t *testing.T, previous, updated domain.Snapshot) {
	t.Helper()
	if updated.Revision != 4 || updated.BuildXML != previous.BuildXML || updated.InputSHA256 != previous.InputSHA256 ||
		updated.CalculatedFromRevision != previous.BuildRevision {
		t.Fatal("пересчёт изменил сборку или её исходную ревизию")
	}
	result, err := application.DecodeCalculation(updated)
	if err != nil || result == nil {
		t.Fatalf("не прочитан результат LuaJIT: %v", err)
	}
	if result.DataID != calculation.LocalPoBDataID || result.Status != "complete" || result.Skill != spells[0].name ||
		math.Abs(result.Stats["TotalDPS"]-spells[0].dps) > 1e-8 || len(result.Issues) != 0 {
		t.Fatalf("некорректный результат LuaJIT: %+v", result)
	}
}
