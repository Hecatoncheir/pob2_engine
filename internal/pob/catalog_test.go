package pob

import (
	"context"
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
