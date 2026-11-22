package pob

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestLuaCatalogUsesOriginalPoB(t *testing.T) {
	calculator := localCalculator(t)
	ctx, cancel := context.WithTimeout(t.Context(), 40*time.Second)
	defer cancel()
	data, err := calculator.CatalogData(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !data.IsRuntime() || data.TreeVersion() != "0_5" || len(data.Gems()) < 100 || len(data.Items()) < 100 || len(data.Nodes()) < 1000 {
		t.Fatalf("неполный каталог PoB: gems=%d items=%d nodes=%d tree=%s", len(data.Gems()), len(data.Items()), len(data.Nodes()), data.TreeVersion())
	}
	for _, id := range []string{"Ice Nova", "Arc", "Fireball", "Rapid Casting I"} {
		gem, ok := data.Gem(id)
		if !ok || len(gem.Levels) == 0 || len(gem.Qualities) == 0 {
			t.Fatalf("нет камня %s: %+v", id, gem)
		}
	}
	if _, ok := data.Item("Rusted Cuirass"); !ok {
		t.Fatal("нет базы брони")
	}
	repeated, err := calculator.CatalogData(ctx)
	if err != nil || repeated != data {
		t.Fatal("каталог не сохранился после успешной загрузки", err)
	}
}

func TestLuaCatalogConcurrentReadsShareSuccessfulData(t *testing.T) {
	calculator := localCalculator(t)
	ctx, cancel := context.WithTimeout(t.Context(), 40*time.Second)
	defer cancel()
	var group sync.WaitGroup
	for range 8 {
		group.Go(func() {
			data, err := calculator.CatalogData(ctx)
			if err != nil || data == nil {
				t.Errorf("параллельный каталог: %v", err)
			}
		})
	}
	group.Wait()
}

func TestLuaCatalogRestartsAfterCancellation(t *testing.T) {
	calculator := localCalculator(t)
	ctx, cancel := context.WithTimeout(t.Context(), 40*time.Second)
	defer cancel()
	marker := startStalledLuaWorker(t, ctx, calculator)
	request, abort := context.WithCancel(ctx)
	defer abort()
	finished := make(chan error, 1)
	go func() {
		_, err := calculator.CatalogData(request)
		finished <- err
	}()
	waitForLuaCalculation(t, ctx, marker)
	abort()
	select {
	case err := <-finished:
		if !errors.Is(err, context.Canceled) || calculator.data != nil || calculator.worker != nil {
			t.Fatalf("отмена каталога сохранила ошибочное состояние: %v", err)
		}
	case <-ctx.Done():
		t.Fatal("каталог не завершился после отмены")
	}
	data, err := calculator.CatalogData(ctx)
	if err != nil || data == nil {
		t.Fatal("каталог не восстановился после отмены", err)
	}
}

func TestLuaFullCatalogBuildsMatchIndependentReference(t *testing.T) {
	calculator := localCalculator(t)
	ctx, cancel := context.WithTimeout(t.Context(), 40*time.Second)
	defer cancel()
	root := filepath.Join("..", "..", "testdata", "reference", "0.0.4")
	for _, sample := range readCatalogReference(t, root) {
		t.Run(sample.ID, func(t *testing.T) {
			xml, err := os.ReadFile(filepath.Join(root, sample.ID+".xml"))
			if err != nil || fmt.Sprintf("%x", sha256.Sum256(xml)) != sample.InputSHA256 {
				t.Fatal("изменён эталон XML", err)
			}
			result, err := calculator.Calculate(ctx, readBuild(t, "0.0.4/"+sample.ID+".xml"))
			if err != nil || result.Status != "complete" || len(result.Issues) != 0 {
				t.Fatalf("неполный расчёт нового умения: %+v %v", result, err)
			}
			checkReferenceStats(t, result.Stats, sample.Stats)
		})
	}
}

type catalogReferenceCase struct {
	ID          string         `json:"id"`
	InputSHA256 string         `json:"input_sha256"`
	Stats       map[string]any `json:"stats"`
}

func readCatalogReference(t *testing.T, root string) []catalogReferenceCase {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, "golden.json"))
	if err != nil {
		t.Fatal(err)
	}
	var golden struct {
		Cases []catalogReferenceCase `json:"cases"`
	}
	if err := json.Unmarshal(data, &golden); err != nil {
		t.Fatal(err)
	}
	return golden.Cases
}

func checkReferenceStats(t *testing.T, actual map[string]float64, expected map[string]any) {
	t.Helper()
	for name, value := range expected {
		number, numeric := value.(float64)
		if !numeric {
			continue
		}
		got, found := actual[name]
		if !found || math.Abs(got-number) > math.Max(1e-8, math.Abs(number)*1e-10) {
			t.Errorf("%s: получено %.12g вместо %.12g", name, got, number)
		}
	}
}
