package maintenance

import (
	"context"
	"time"

	"github.com/Hecatoncheir/pob2_engine/internal/pob"
)

func updatePoB(ctx context.Context, apply bool) (pob.State, error) {
	manager, err := pob.Local()
	if err != nil {
		return failedComponent(pob.State{}, err), err
	}
	state, err := manager.Update(ctx, apply)
	return failedComponent(state, err), err
}

func failedComponent(state pob.State, err error) pob.State {
	if err != nil {
		if state.Status == "" {
			state.Status = "unavailable"
		}
		state.Message = err.Error()
	}
	return state
}

func (s *Service) engineState(update UpdateResult, err error) pob.State {
	return failedComponent(pob.State{Status: update.Status, InstalledVersion: s.version,
		LatestVersion: update.LatestVersion, ReleaseURL: update.ReleaseURL, RestartRequired: update.RestartRequired}, err)
}

func (s *Service) updateAll(ctx context.Context, apply bool) (UpdateResult, error) {
	// Сначала обновляем комплект PoB, пока текущий процесс ещё обслуживает команды.
	upstream, pobErr := updatePoB(ctx, apply)
	engine, engineErr := s.updateEngine(ctx, apply)
	state := s.engineState(engine, engineErr)
	engine.Engine, engine.PoB = &state, &upstream
	engine.RestartRequired = engine.RestartRequired || upstream.RestartRequired
	switch {
	case engineErr != nil || pobErr != nil:
		engine.Status = "partial"
	case engine.Status == "updated" || upstream.Status == "updated":
		engine.Status = "updated"
	case engine.Status == "available" || upstream.Status == "available":
		engine.Status = "available"
	default:
		engine.Status = "up_to_date"
	}
	return engine, nil
}

func (s *Service) inspectUpdates(ctx context.Context, report *Report) {
	check, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	upstream := make(chan pob.State, 1)
	go func() {
		state, _ := updatePoB(check, false)
		upstream <- state
	}()
	update, err := s.updateEngine(check, false)
	engine := s.engineState(update, err)
	installed := <-upstream
	report.EngineUpdate, report.PoB = &engine, &installed
}
