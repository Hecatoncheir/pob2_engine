package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Hecatoncheir/pob2_engine/internal/application"
)

func TestPublished001SessionAndOriginalResultsSurviveUpgrade(t *testing.T) {
	root := "../../testdata/compat/0.0.1"
	data, err := os.ReadFile(filepath.Join(root, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Release string            `json:"release"`
		Files   map[string]string `json:"files"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil || manifest.Release != "v0.0.1" {
		t.Fatalf("неверный эталон совместимости: %v", err)
	}
	for name, digest := range manifest.Files {
		content, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(content)
		if hex.EncodeToString(sum[:]) != digest {
			t.Fatalf("изменён эталон %s", name)
		}
	}
	data, err = os.ReadFile(filepath.Join(root, "snapshot.json"))
	if err != nil {
		t.Fatal(err)
	}
	var expected application.SnapshotView
	if err := json.Unmarshal(data, &expected); err != nil {
		t.Fatal(err)
	}
	base := t.TempDir()
	if _, err := Restore(t.Context(), filepath.Join(root, "sessions.db"), base, "upgrade"); err != nil {
		t.Fatal(err)
	}
	_, service := openService(t, base, "upgrade")
	saved, err := service.Snapshot(t.Context(), "main")
	if err != nil {
		t.Fatal(err)
	}
	view, err := application.ViewSnapshot(saved)
	if err != nil || !reflect.DeepEqual(view, expected) {
		t.Fatalf("открытие изменило ответ прежней версии: %v", err)
	}
	old := application.CalculateBuild{SessionID: "main", ExpectedRevision: 2, RequestID: "calculate"}
	if replay, err := service.Calculate(t.Context(), old); err != nil || replay != saved {
		t.Fatal("не восстановлен исходный request_id")
	}
	updated, err := service.Calculate(t.Context(), application.CalculateBuild{SessionID: "main", ExpectedRevision: 3, RequestID: "new-version"})
	if err != nil || updated.Revision != 4 || updated.BuildXML != saved.BuildXML {
		t.Fatalf("новый расчёт после обновления: %v", err)
	}
	previousResult, err := application.DecodeCalculation(saved)
	if err != nil {
		t.Fatal(err)
	}
	newResult, err := application.DecodeCalculation(updated)
	if err != nil || newResult.DataID == previousResult.DataID || newResult.DataSHA256 == previousResult.DataSHA256 || !reflect.DeepEqual(newResult.Stats, previousResult.Stats) {
		t.Fatalf("новый расчёт не использует новый набор либо меняет прежние числа: %v", err)
	}
	if replay, err := service.Calculate(t.Context(), old); err != nil || replay != saved {
		t.Fatal("новый расчёт подменил исторический ответ")
	}
	if current, err := service.Snapshot(t.Context(), "main"); err != nil || current != updated {
		t.Fatal("повтор старого запроса откатил сессию")
	}
}
