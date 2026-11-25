//go:build windows || ((linux || darwin) && cgo)

package pob

import (
	"bytes"
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

func TestLuaHybridTreeAcceptsBothLineEndings(t *testing.T) {
	engine := localCalculator(t)
	source, err := os.ReadFile(filepath.Join(engine.layout.Source, "Classes", "PassiveSpec.lua"))
	if err != nil {
		t.Fatal(err)
	}
	lf := bytes.ReplaceAll(source, []byte("\r\n"), []byte("\n"))
	crlf := bytes.ReplaceAll(lf, []byte("\n"), []byte("\r\n"))
	for name, data := range map[string][]byte{"LF": lf, "CRLF": crlf} {
		path := filepath.Join(t.TempDir(), "PassiveSpec.lua")
		if err := os.WriteFile(path, data, 0600); err != nil {
			t.Fatal(err)
		}
		if !compatiblePassiveSpec(path) {
			t.Errorf("закреплённый PassiveSpec с окончаниями %s не распознан", name)
		}
	}
}

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
	worker, err := os.ReadFile("bridge/worker.lua")
	if err != nil {
		t.Fatal(err)
	}
	bootstrap, _, found := strings.Cut(string(worker), "if POB2_GO_PATH_CALLBACK")
	if !found {
		t.Fatal("не найдена точка загрузки Lua-обхода")
	}
	return requireLuaJIT + hybridTreeEnvironment(layout, directory) + bootstrap +
		fmt.Sprintf("local xml = %q\n", fixtureXML(t, weaponTreeFixture)) + hybridTreeCheck
}

func hybridTreeEnvironment(layout Layout, directory string) string {
	_, extension := luaLibrary(layout.Runtime, layout.Native)
	// Проверка запускает Lua внутри процесса тестов, где библиотека уже могла
	// прочитать окружение до подстановки каталогов.
	return fmt.Sprintf(`
local getenv = os.getenv
local paths = { POB2_BRIDGE = %q, POB2_RUNTIME = %q, POB2_NATIVE = %q, POB2_LIBRARY_EXTENSION = %q }
os.getenv = function(name) return paths[name] or getenv(name) end
`, filepath.ToSlash(directory), filepath.ToSlash(layout.Runtime), filepath.ToSlash(layout.Native), extension)
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
