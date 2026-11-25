package pob

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func TestBytecodeRequiresTheBuiltLibrary(t *testing.T) {
	library := filepath.Join(t.TempDir(), "luajit")
	data := []byte("версия LuaJIT в поставке")
	writeBytecodeFixture(t, library, string(data))
	assets := fstest.MapFS{"manifest.json": {
		Data: fmt.Appendf(nil, `{"schema":1,"library_sha256":"%x"}`, sha256.Sum256(data)),
	}}
	if !matchesBytecodeLibrary(assets, library) {
		t.Fatal("не выбрана библиотека, которой собран байткод")
	}
	writeBytecodeFixture(t, library, "другая версия LuaJIT")
	if matchesBytecodeLibrary(assets, library) {
		t.Fatal("байткод включён на другой библиотеке")
	}
}

func TestLuaBytecodeLoadsModulesAndFallbacks(t *testing.T) {
	engine := localCalculator(t)
	library, _ := luaLibrary(engine.layout.Runtime, engine.layout.Native)
	root := t.TempDir()
	compiled := filepath.Join(root, "compiled")
	if err := os.Mkdir(compiled, 0700); err != nil {
		t.Fatal(err)
	}
	result := filepath.Join(root, "result.txt")
	t.Setenv("POB2_BRIDGE", filepath.ToSlash(root))
	t.Setenv("POB2_TEST_BYTECODE_RESULT", filepath.ToSlash(result))
	const record = `local function record(value)
		local file = assert(io.open(os.getenv("POB2_TEST_BYTECODE_RESULT"), "wb"))
		file:write(value)
		file:close()
	end
	`
	const main = record + `record(assert(loadfile(os.getenv("POB2_BRIDGE") .. "/module.lua"))())`
	writeBytecodeFixture(t, filepath.Join(root, "module.lua"), `return "исходный модуль"`)
	dumpBytecodeFixture(t, library, filepath.Join(compiled, "worker.lua"), main)
	dumpBytecodeFixture(t, library, filepath.Join(compiled, "module.lua"), `return "байткод модуля"`)
	assertBytecodeResult(t, library, root, result, record, "байткод модуля")

	writeBytecodeFixture(t, filepath.Join(compiled, "module.lua"), "повреждённый байткод")
	assertBytecodeResult(t, library, root, result, record, "исходный модуль")

	writeBytecodeFixture(t, filepath.Join(compiled, "worker.lua"), "повреждённый байткод")
	assertBytecodeResult(t, library, root, result, record, "исходный вход")
}

func assertBytecodeResult(t *testing.T, library, root, result, record, want string) {
	t.Helper()
	if err := runLua(library, workerEntry(root, "worker.lua")+record+`record("исходный вход")`); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(result)
	if err != nil || string(data) != want {
		t.Fatalf("результат %q, ошибка %v, нужен %q", data, err, want)
	}
}

func dumpBytecodeFixture(t *testing.T, library, path, program string) {
	t.Helper()
	code := fmt.Sprintf(`local file = assert(io.open(%q, "wb"))
		file:write(string.dump(assert(loadstring(%q))))
		file:close()`, filepath.ToSlash(path), program)
	if err := runLua(library, code); err != nil {
		t.Fatal(err)
	}
}

func writeBytecodeFixture(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0600); err != nil {
		t.Fatal(err)
	}
}
