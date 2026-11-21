package calculation

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/Hecatoncheir/pob2_engine/internal/build"
)

func TestSupportPreviewMatchesRecordedLua(t *testing.T) {
	engine := testEngine(t)
	checkPreviewLuaMatrix(t, engine, "../../testdata/reference", "cases.lua.json", previewBase(t, "../../testdata/reference/cases/gem-15.xml"), "")
	for _, spell := range []string{"frostbolt", "spark"} {
		checkPreviewLuaMatrix(t, engine, "../../testdata/reference/0.0.2", "spells.lua.json", previewBase(t, "../../testdata/reference/0.0.2/cases/"+spell+"-gem-10.xml"), spell+"-")
	}
}

func checkPreviewLuaMatrix(t *testing.T, engine *Engine, directory, file string, before *build.Document, prefix string) {
	t.Helper()
	data, err := os.ReadFile(directory + "/" + file)
	if err != nil {
		t.Fatal(err)
	}
	var golden struct {
		Cases []luaCase `json:"cases"`
	}
	if err := json.Unmarshal(data, &golden); err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, sample := range golden.Cases {
		ids := previewSupportIDs(strings.TrimPrefix(sample.ID, prefix))
		if len(ids) == 0 {
			continue
		}
		count++
		t.Run(sample.ID, func(t *testing.T) { checkPreviewLuaCase(t, engine, before, sample, ids) })
	}
	if count != 3 {
		t.Fatalf("изменён состав эталонов поддержки: %d", count)
	}
}

func previewSupportIDs(id string) []string {
	switch id {
	case "support-rapid":
		return []string{"Rapid Casting I"}
	case "support-controlled":
		return []string{"Controlled Destruction"}
	case "support-combined":
		return []string{"Rapid Casting I", "Controlled Destruction"}
	default:
		return nil
	}
}

func previewBase(t *testing.T, path string) *build.Document {
	t.Helper()
	xml, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	doc, err := build.Parse(xml)
	if err != nil {
		t.Fatal(err)
	}
	return doc
}

func checkPreviewLuaCase(t *testing.T, engine *Engine, before *build.Document, sample luaCase, ids []string) {
	t.Helper()
	after := before
	for _, id := range ids {
		gem, ok := engine.data.Gem(id)
		if !ok {
			t.Fatal("нет эталонной поддержки")
		}
		var err error
		after, _, err = after.WithSupport(gem)
		if err != nil {
			t.Fatal(err)
		}
	}
	result, err := engine.Compare(t.Context(), before, after)
	if err != nil {
		t.Fatal(err)
	}
	checkLuaResult(t, result.After, sample)
}
