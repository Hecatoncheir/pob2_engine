package pob

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Hecatoncheir/pob2_engine/internal/privatefs"
)

// Активная установка остаётся в прежних полях манифеста для старых бинарников.
type installationState struct {
	Installation
	Previous *Installation `json:"previous,omitempty"`
}

func (m *Manager) Installations() (Installation, *Installation, error) {
	entry, err := m.readInstallation()
	return entry.Installation, entry.Previous, err
}

func (m *Manager) readInstallation() (installationState, error) {
	if m.local != "" {
		i, err := localInstallation(m.local, m.localNative)
		return installationState{Installation: i}, err
	}
	var entry installationState
	data, err := os.ReadFile(filepath.Join(m.root, "current.json"))
	if err != nil {
		return entry, err
	}
	if err := json.Unmarshal(data, &entry); err != nil {
		return installationState{}, err
	}
	if !validInstallation(entry.Installation) || entry.Previous != nil &&
		(!validInstallation(*entry.Previous) || !newer(entry.Version, entry.Previous.Version)) {
		return installationState{}, fmt.Errorf("повреждён манифест установки PoB")
	}
	entry.Installation = m.locate(entry.Installation)
	if entry.Previous != nil {
		previous := m.locate(*entry.Previous)
		entry.Previous = &previous
	}
	return entry, nil
}

func validInstallation(i Installation) bool {
	return tags.MatchString(i.Version) && digestPattern.MatchString(i.SHA256) && commitPattern.MatchString(i.Commit)
}

func (i installationState) describe(state State) State {
	state.InstalledVersion, state.InstalledCommit = i.Version, i.Commit
	state.PreviousVersion, state.PreviousCommit = "", ""
	if i.Previous != nil {
		state.PreviousVersion, state.PreviousCommit = i.Previous.Version, i.Previous.Commit
	}
	return state
}

func (m *Manager) writeInstallation(ctx context.Context, entry installationState) error {
	f, err := os.CreateTemp(m.root, ".pob-current-")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if err := writeManifest(f, entry); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return privatefs.Replace(f.Name(), filepath.Join(m.root, "current.json"))
}

// Rollback возвращает проверенную предыдущую установку без сетевых запросов.
func (m *Manager) Rollback(ctx context.Context, apply bool) (State, error) {
	if err := ctx.Err(); err != nil {
		return State{}, err
	}
	if m.local != "" || !apply {
		return m.rollback(ctx, false)
	}
	lock, err := waitForInstaller(ctx, filepath.Join(m.root, "install.lock"))
	if errors.Is(err, os.ErrNotExist) {
		return unavailableRollback(), nil
	}
	if err != nil {
		return State{}, fmt.Errorf("блокировка отката PoB: %w", err)
	}
	defer lock.Close()
	return m.rollback(ctx, true)
}

func unavailableRollback() State {
	return State{Status: "rollback_unavailable", Message: "нет предыдущей управляемой установки PoB"}
}

func (m *Manager) rollback(ctx context.Context, apply bool) (State, error) {
	entry, err := m.readInstallation()
	if errors.Is(err, os.ErrNotExist) {
		return unavailableRollback(), nil
	}
	if err != nil {
		return State{Status: "invalid_installation"}, err
	}
	state := entry.describe(unavailableRollback())
	if entry.Previous == nil {
		return state, nil
	}
	if _, err := entry.Previous.Layout(); err != nil {
		state.Message = "предыдущая установка PoB непригодна: " + err.Error()
		return state, nil
	}
	state.Status, state.Message = "rollback_available", ""
	if !apply {
		return state, nil
	}
	return m.activateRollback(ctx, *entry.Previous, state)
}

func (m *Manager) activateRollback(ctx context.Context, previous Installation, state State) (State, error) {
	if err := m.probe(ctx, previous); err != nil {
		return state, fmt.Errorf("предыдущая установка PoB не прошла проверку: %w", err)
	}
	rolledBack := installationState{Installation: previous}
	if err := m.writeInstallation(ctx, rolledBack); err != nil {
		return state, fmt.Errorf("активация предыдущего PoB: %w", err)
	}
	return rolledBack.describe(State{Status: "rolled_back", RestartRequired: true}), nil
}
