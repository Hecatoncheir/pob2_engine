package application

import (
	"context"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func catalogPage(t *testing.T, s *Service, query CatalogQuery) CatalogPage {
	t.Helper()
	value, err := s.Catalog(t.Context(), query)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func TestCatalogCoverageAndPagination(t *testing.T) {
	s := NewService()
	for kind, count := range map[string]int{"gems": 5, "classes": 8, "items": 5, "passives": 11} {
		t.Run(kind, func(t *testing.T) { checkCatalogPages(t, s, kind, count) })
	}
	query := CatalogQuery{Kind: "gems", Query: " RAPIDCASTINGSUPPORT "}
	page := catalogPage(t, s, query)
	if page.Total != 1 || !page.Entries[0].Support || page.Entries[0].Family == "" {
		t.Fatalf("не найден псевдоним поддержки: %+v", page)
	}
	if !reflect.DeepEqual(page.Entries[0].Levels, []int{1}) || !reflect.DeepEqual(page.Entries[0].Qualities, []int{0}) {
		t.Fatal("каталог обещает неподдержанные уровни")
	}
	page.Entries[0].Coverage[0] = "изменено"
	if reflect.DeepEqual(page, catalogPage(t, s, query)) {
		t.Fatal("изменён общий каталог")
	}
}

func checkCatalogPages(t *testing.T, s *Service, kind string, count int) {
	t.Helper()
	page := catalogPage(t, s, CatalogQuery{Kind: kind, Limit: 100})
	if page.Total != count || len(page.Entries) != count || page.DataID == "" || len(page.DataSHA256) != 64 {
		t.Fatalf("неполный каталог: %+v", page)
	}
	for offset, entry := range page.Entries {
		part := catalogPage(t, s, CatalogQuery{Kind: kind, Offset: offset, Limit: 1})
		if !reflect.DeepEqual(part.Entries, []CatalogEntry{entry}) || len(entry.Coverage) == 0 {
			t.Fatalf("разные страницы или нет границ покрытия: %+v", part)
		}
	}
	empty := catalogPage(t, s, CatalogQuery{Kind: kind, Offset: 10000})
	if len(empty.Entries) != 0 || empty.Entries == nil {
		t.Fatal("конец каталога должен возвращать пустой массив")
	}
}

func TestCatalogDoesNotPromiseSorceressPassivesForOtherClasses(t *testing.T) {
	s := NewService()
	chain := catalogPage(t, s, CatalogQuery{Kind: "passives", Query: "18845"}).Entries[0]
	if !reflect.DeepEqual(chain.ClassIDs, []int{7}) {
		t.Fatal("цепочка доступна чужому классу")
	}
	monk := catalogPage(t, s, CatalogQuery{Kind: "passives", Query: "44683"}).Entries[0]
	if !reflect.DeepEqual(monk.ClassIDs, []int{10}) {
		t.Fatal("чужой старт доступен Sorceress")
	}
	shared := catalogPage(t, s, CatalogQuery{Kind: "passives", Query: "54447"}).Entries[0]
	if !slices.Contains(shared.ClassIDs, 1) || !slices.Contains(shared.ClassIDs, 7) {
		t.Fatal("потерян общий старт Witch/Sorceress")
	}
}

func TestCatalogRejectsUnboundedAndInvalidQueries(t *testing.T) {
	s := NewService()
	for _, query := range []CatalogQuery{
		{Kind: "all"}, {Kind: "gems", Limit: 101}, {Kind: "gems", Limit: -1},
		{Kind: "gems", Offset: -1}, {Kind: "gems", Offset: 10001}, {Kind: "gems", Query: strings.Repeat("я", 129)},
		{Kind: "gems", Query: string([]byte{0xff})},
	} {
		_, err := s.Catalog(t.Context(), query)
		expectCode(t, err, InvalidArgument)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := s.Catalog(ctx, CatalogQuery{Kind: "gems"}); err != context.Canceled {
		t.Fatal(err)
	}
}
