package pob

import (
	"embed"

	"github.com/Hecatoncheir/pob2_engine/internal/poblegacy"
)

const bridgeRoot = "bridge"

//go:embed all:bridge
var bridge embed.FS

// RunWorker запускает LuaJIT в отдельном процессе PoB2Engine.
func RunWorker(args []string) error {
	program, err := bridge.ReadFile("bridge/worker.lua")
	if err != nil {
		return err
	}
	return runWorkerProgram(args, "worker.lua", string(program))
}

func RunWorkbenchWorker(args []string) error {
	setup, err := bridge.ReadFile("bridge/workbench.lua")
	if err != nil {
		return err
	}
	program, err := poblegacy.Read("lua/headless_server.lua")
	if err != nil {
		return err
	}
	return runWorkerProgram(args, "lua/headless_server.lua", string(setup)+"\n"+string(program))
}

func unpackWorkbenchLua(directory string) error {
	return poblegacy.Unpack(directory, "lua")
}
