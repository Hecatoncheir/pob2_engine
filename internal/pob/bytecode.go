package pob

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

var bundledBytecode fs.FS

type bytecodeManifest struct {
	Schema        int    `json:"schema"`
	LibrarySHA256 string `json:"library_sha256"`
}

//go:embed bytecode_bootstrap.lua
var bytecodeBootstrap string

const sourceBridgeProbe = `if os.getenv("POB2_ENGINE_TRACE_LUA") == "1" then io.stderr:write("pob2-lua: source\n") end` + "\n"

func workerEntry(directory, entry string) string {
	if !regularFile(filepath.Join(directory, "compiled", entry)) {
		return sourceBridgeProbe
	}
	return fmt.Sprintf("if (function(entry)\n%s\nend)(%q) then return end\n", bytecodeBootstrap, entry) + sourceBridgeProbe
}

func stageBytecode(layout Layout, directory, command string) {
	if bundledBytecode == nil || os.Getenv("POB2_ENGINE_LUA_SOURCE") == "1" {
		return
	}
	library, _ := luaLibrary(layout.Runtime, layout.Native)
	if !matchesBytecodeLibrary(bundledBytecode, library) {
		return
	}
	name := "bridge"
	if command == "_workbench-lua" {
		name = "workbench"
	}
	staging := filepath.Join(directory, "compiled-staging")
	if unpackLuaFiles(bundledBytecode, name, staging) == nil {
		_ = os.Rename(staging, filepath.Join(directory, "compiled"))
	}
}

func matchesBytecodeLibrary(assets fs.FS, library string) bool {
	data, err := fs.ReadFile(assets, "manifest.json")
	if err != nil {
		return false
	}
	var manifest bytecodeManifest
	if json.Unmarshal(data, &manifest) != nil || manifest.Schema != 1 || len(manifest.LibrarySHA256) != sha256.Size*2 {
		return false
	}
	dll, err := os.Open(library)
	if err != nil {
		return false
	}
	defer dll.Close()
	digest := sha256.New()
	_, err = io.Copy(digest, dll)
	return err == nil && hex.EncodeToString(digest.Sum(nil)) == manifest.LibrarySHA256
}
