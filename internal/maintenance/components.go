package maintenance

import (
	"context"
	"strings"
	"time"

	"github.com/Hecatoncheir/pob2_engine/internal/pob"
)

func (s *Service) updatePoB(ctx context.Context, apply bool, target string) (pob.State, error) {
	manager, err := pob.Local()
	if err != nil {
		return failedComponent(pob.State{}, err), err
	}
	var state pob.State
	if target == "pob-rollback" {
		state, err = manager.Rollback(ctx, apply)
	} else if tag, ok := strings.CutPrefix(target, "pob@"); ok {
		state, err = manager.UpdateTo(ctx, apply, tag)
	} else {
		state, err = manager.Update(ctx, apply)
	}
	s.reportRunningPoB(&state)
	return failedComponent(state, err), err
}

func (s *Service) reportRunningPoB(state *pob.State) {
	if running := s.running.Load(); running != nil {
		state.RunningVersion, state.RunningCommit = running.Version, running.Commit
		if state.InstalledVersion != "" {
			state.RestartRequired = state.InstalledVersion != running.Version || state.InstalledCommit != running.Commit
		}
	}
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
	upstream, pobErr := s.updatePoB(ctx, apply, "pob")
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
		state, _ := s.updatePoB(check, false, "pob")
		upstream <- state
	}()
	update, err := s.updateEngine(check, false)
	engine := s.engineState(update, err)
	installed := <-upstream
	report.EngineUpdate, report.PoB = &engine, &installed
	report.RestartRequired = report.RestartRequired || installed.RestartRequired
}
