package pob

import (
	"crypto/sha256"
	"fmt"
	"os"
)

const luaTreePaths = "POB2_ENGINE_TREE_PATHS"

// SHA-256 файла Classes/PassiveSpec.lua из архива, закреплённого initialRelease.
const passiveSpecSHA256 = "24bb08cea7d18a46185dc467d12491aa133965715efeee180b8cb19b0debc9e8"

func treePathProgram(program string, callback func() uintptr) string {
	if os.Getenv(luaTreePaths) == "lua" || !compatiblePassiveSpec("Classes/PassiveSpec.lua") {
		return program
	}
	return withTreePathCallback(program, callback())
}

func compatiblePassiveSpec(path string) bool {
	source, err := os.ReadFile(path)
	return err == nil && fmt.Sprintf("%x", sha256.Sum256(source)) == passiveSpecSHA256
}

func withTreePathCallback(program string, callback uintptr) string {
	// Целочисленный литерал LuaJIT сохраняет все биты указателя.
	return fmt.Sprintf("POB2_GO_PATH_CALLBACK = 0x%xULL\n%s", callback, program)
}
