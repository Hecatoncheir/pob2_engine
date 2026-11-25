//go:build windows || ((linux || darwin) && cgo)

package pob

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Hecatoncheir/pob2_engine/internal/build"
)

const weaponTreeFixture = "0.0.12/witch-weapon-paths.xml"

//go:embed testdata/hybrid_tree_paths.lua
var hybridTreeCheck string

func TestLuaHybridTreeNativeCallback(t *testing.T) {
	engine := localCalculator(t)
	program := hybridTreeProgram(t, engine.layout)
	t.Chdir(engine.layout.Source)
	t.Setenv(luaTreePaths, "")
	if !compatiblePassiveSpec("Classes/PassiveSpec.lua") {
		t.Fatal("эталон PassiveSpec для проверки гибридного обхода не совпадает")
	}
	library, _ := luaLibrary(engine.layout.Runtime, engine.layout.Native)
	if err := runLua(library, program); err != nil {
		t.Fatal(err)
	}
}

func hybridTreeProgram(t *testing.T, layout Layout) string {
	t.Helper()
	directory, err := unpackBridge(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	_, extension := luaLibrary(layout.Runtime, layout.Native)
	t.Setenv("POB2_BRIDGE", filepath.ToSlash(directory))
	t.Setenv("POB2_RUNTIME", filepath.ToSlash(layout.Runtime))
	t.Setenv("POB2_NATIVE", filepath.ToSlash(layout.Native))
	t.Setenv("POB2_LIBRARY_EXTENSION", extension)
	worker, err := os.ReadFile("bridge/worker.lua")
	if err != nil {
		t.Fatal(err)
	}
	bootstrap, _, found := strings.Cut(string(worker), "if POB2_GO_PATH_CALLBACK")
	if !found {
		t.Fatal("не найдена точка загрузки Lua-обхода")
	}
	return requireLuaJIT + bootstrap +
		fmt.Sprintf("local xml = %q\n", fixtureXML(t, weaponTreeFixture)) + hybridTreeCheck
}

func fixtureXML(t *testing.T, name string) []byte {
	t.Helper()
	xml, err := readBuild(t, name).ExportWithoutProgression()
	if err != nil {
		t.Fatal(err)
	}
	return xml
}

func TestLuaHybridTreeEditParity(t *testing.T) {
	want := luaTreeEdits(t, "lua")
	got := luaTreeEdits(t, "")
	if !reflect.DeepEqual(got, want) {
		t.Fatal("гибридный обход изменил сохранённый XML или отчёт")
	}
}

type treeEditSnapshot struct {
	xml    string
	report any
}

func luaTreeEdits(t *testing.T, mode string) []treeEditSnapshot {
	t.Helper()
	t.Setenv(luaTreePaths, mode)
	engine := localCalculator(t)
	defer engine.Close()
	ctx, cancel := context.WithTimeout(t.Context(), 45*time.Second)
	defer cancel()

	doc := readBuild(t, weaponTreeFixture)
	results := make([]treeEditSnapshot, 0, 2)
	for _, action := range []string{"allocate", "refund"} {
		doc = confirmedTreeEdit(t, ctx, engine, doc, action)
		results = append(results, treeEditResult(t, ctx, engine, doc))
	}
	return results
}

func confirmedTreeEdit(t *testing.T, ctx context.Context, engine *Calculator, doc *build.Document, action string) *build.Document {
	t.Helper()
	change, err := engine.ResolveTreeEdit(ctx, doc, build.TreeNodeEdit{
		NodeID: 51184, Action: action, AllocationMode: "weapon_set_1",
	})
	if err != nil {
		t.Fatal(err)
	}
	next, err := doc.WithResolvedTreeEdit(change)
	if err != nil {
		t.Fatal(err)
	}
	return next
}

func treeEditResult(t *testing.T, ctx context.Context, engine *Calculator, doc *build.Document) treeEditSnapshot {
	t.Helper()
	report, err := engine.ReadBuildReport(ctx, doc, "Witch")
	if err != nil {
		t.Fatal(err)
	}
	xml, err := doc.ExportWithoutProgression()
	if err != nil {
		t.Fatal(err)
	}
	var decoded any
	if err := json.Unmarshal(report, &decoded); err != nil {
		t.Fatal(err)
	}
	return treeEditSnapshot{xml: string(xml), report: decoded}
}
