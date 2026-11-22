package pob

import (
	"context"
	"errors"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestLuaJITRestartsAfterCanceledCalculation(t *testing.T) {
	engine := localCalculator(t)
	spell := spells[0]
	doc := readBuild(t, spell.file)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	marker := startStalledLuaWorker(t, ctx, engine)

	calculation, abort := context.WithCancel(ctx)
	defer abort()
	finished := make(chan error, 1)
	go func() {
		_, err := engine.Calculate(calculation, doc)
		finished <- err
	}()
	waitForLuaCalculation(t, ctx, marker)
	abort()
	select {
	case err := <-finished:
		if !errors.Is(err, context.Canceled) || engine.worker != nil {
			t.Fatalf("ожидалась остановка зависшего LuaJIT: %v", err)
		}
	case <-ctx.Done():
		t.Fatal("расчёт не завершился после отмены")
	}

	result, err := engine.Calculate(ctx, readBuild(t, "0.0.5/caster-alternate.xml"))
	if err != nil || result.Status != "complete" || result.Skill != "Frostbolt" || math.Abs(result.Stats["FullDPS"]-9.1) > 1e-8 {
		t.Fatalf("перезапущенный LuaJIT не вернул эталонный расчёт: %+v, %v", result, err)
	}
}

func waitForLuaCalculation(t *testing.T, ctx context.Context, marker string) {
	t.Helper()
	tick := time.NewTicker(10 * time.Millisecond)
	defer tick.Stop()
	for !regularFile(marker) {
		select {
		case <-ctx.Done():
			t.Fatal("LuaJIT не начал тестовый расчёт")
		case <-tick.C:
		}
	}
}

func startStalledLuaWorker(t *testing.T, ctx context.Context, engine *Calculator) string {
	t.Helper()
	const bootstrap = `
common = { xml = { ParseXML = function() return { { elem = "PathOfBuilding2" } } end } }
local function stall()
	local file = assert(io.open(os.getenv("POB2_TEST_WORKER_MARKER"), "w"))
	file:close()
	while true do end
end
loadBuildFromXML = stall
main = { LoadTree = stall }
`
	work := t.TempDir()
	marker := filepath.Join(work, "calculation-started")
	bridge, err := unpackBridge(work)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bridge, "lua", "headless_bootstrap.lua"), []byte(bootstrap), 0600); err != nil {
		t.Fatal(err)
	}
	cmd := stalledLuaCommand(t, engine.layout, bridge, marker)
	worker, err := openWorker(cmd)
	if err != nil {
		t.Fatal(err)
	}
	engine.worker = worker
	ready, err := engine.exchange(ctx, nil)
	if err != nil || !ready.Ready {
		t.Fatalf("тестовый LuaJIT не запустился: %v", err)
	}
	return marker
}

func stalledLuaCommand(t *testing.T, layout Layout, bridge, marker string) *exec.Cmd {
	t.Helper()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(executable, "_pob-lua", layout.Runtime, layout.Native, bridge)
	cmd.Dir, cmd.Stderr = layout.Source, os.Stderr
	cmd.Env = append(cmd.Environ(), "POB2_TEST_WORKER_MARKER="+filepath.ToSlash(marker))
	return cmd
}
