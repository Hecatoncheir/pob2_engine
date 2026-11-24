package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/Hecatoncheir/pob2_engine/internal/calculation"
	"github.com/Hecatoncheir/pob2_engine/internal/maintenance"
	"github.com/Hecatoncheir/pob2_engine/internal/pob"
)

func openCalculator(ctx context.Context) (calculation.Calculator, func(), error) {
	switch os.Getenv("POB2_ENGINE_CALCULATOR") {
	case "", "luajit", "lua":
		return openLuaJITCalculator(ctx)
	default:
		return nil, nil, fmt.Errorf("pob2_engine использует LuaJIT; удалите значение POB2_ENGINE_CALCULATOR из окружения")
	}
}

func openLuaJITCalculator(ctx context.Context) (calculation.Calculator, func(), error) {
	m, err := pob.Local()
	if err != nil {
		return nil, nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, pob.InstallTimeout)
	defer cancel()
	i, err := m.Ensure(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("подготовка PoB Lua: %w", err)
	}
	engine, err := pob.NewCalculator(i)
	if err != nil {
		return nil, nil, err
	}
	maintenance.Local().SetRunningPoB(i)
	return engine, func() { _ = engine.Close() }, nil
}
