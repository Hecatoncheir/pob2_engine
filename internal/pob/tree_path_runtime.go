package pob

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
)

const luaTreePaths = "POB2_ENGINE_TREE_PATHS"

// SHA-256 закреплённого Classes/PassiveSpec.lua с окончаниями строк LF.
const passiveSpecSHA256 = "10c713e7083fefe4cf4b21d13bc6f4adfa36b84e7faeab50e0d2f5007025867c"

func treePathProgram(program string, callback func() uintptr) string {
	if os.Getenv(luaTreePaths) == "lua" || !compatiblePassiveSpec("Classes/PassiveSpec.lua") {
		return program
	}
	return withTreePathCallback(program, callback())
}

func compatiblePassiveSpec(path string) bool {
	source, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	source = bytes.ReplaceAll(source, []byte("\r\n"), []byte("\n"))
	return fmt.Sprintf("%x", sha256.Sum256(source)) == passiveSpecSHA256
}

func withTreePathCallback(program string, callback uintptr) string {
	// Целочисленный литерал LuaJIT сохраняет все биты указателя.
	return fmt.Sprintf("POB2_GO_PATH_CALLBACK = 0x%xULL\n%s", callback, program)
}
