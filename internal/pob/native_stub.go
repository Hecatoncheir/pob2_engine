//go:build (linux || darwin) && !cgo

package pob

import "fmt"

func runLua(_, _ string) error {
	return fmt.Errorf("для LuaJIT нужна сборка PoB2Engine с CGO_ENABLED=1")
}
