package pob

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
)

//go:embed all:bridge
var bridge embed.FS

const requireLuaJIT = `assert(type(jit) == "table" and type(jit.version) == "string", "runtime LuaJIT is required")` + "\n"

// RunWorker запускает LuaJIT в отдельном процессе PoB2Engine.
func RunWorker(args []string) error {
	if len(args) != 3 {
		return fmt.Errorf("воркеру нужны каталоги PoB runtime, нативных библиотек и моста")
	}
	runtimeDir, nativeDir, bridgeDir := args[0], args[1], args[2]
	library, extension := luaLibrary(runtimeDir, nativeDir)
	values := map[string]string{"POB2_BRIDGE": bridgeDir, "POB2_RUNTIME": runtimeDir,
		"POB2_NATIVE": nativeDir, "POB2_LIBRARY_EXTENSION": extension}
	for key, value := range values {
		if err := os.Setenv(key, filepath.ToSlash(value)); err != nil {
			return err
		}
	}
	program, err := bridge.ReadFile("bridge/worker.lua")
	if err != nil {
		return err
	}
	return runLua(library, requireLuaJIT+string(program))
}

func unpackBridge(directory string) (string, error) {
	err := fs.WalkDir(bridge, "bridge", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		target := filepath.Join(directory, filepath.FromSlash(path))
		if entry.IsDir() {
			return os.Mkdir(target, 0700)
		}
		data, err := bridge.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0600)
	})
	return filepath.Join(directory, "bridge"), err
}

func luaLibrary(runtimeDir, nativeDir string) (library, extension string) {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(runtimeDir, "lua51.dll"), "dll"
	case "darwin":
		return filepath.Join(nativeDir, "libluajit-5.1.dylib"), "so"
	default:
		return filepath.Join(nativeDir, "libluajit-5.1.so.2"), "so"
	}
}
