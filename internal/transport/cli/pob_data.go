package cli

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/Hecatoncheir/pob2_engine/internal/pob"
)

type pobDataInstallations struct {
	Current  *pobDataInstallation `json:"current"`
	Previous *pobDataInstallation `json:"previous"`
}

type pobDataInstallation struct {
	Tag             string `json:"tag"`
	Commit          string `json:"commit"`
	Directory       string `json:"directory"`
	NativeDirectory string `json:"nativeDirectory,omitempty"`
	ArchiveSHA256   string `json:"archiveSha256,omitempty"`
}

func pobDataView(value pob.Installation) *pobDataInstallation {
	return &pobDataInstallation{value.Version, value.Commit, value.Directory, value.NativeDirectory, value.SHA256}
}

func (r runner) pobData(ctx context.Context, args []string) error {
	flags := r.flags("pob-data")
	tag := flags.String("tag", "", "указанный стабильный тег PoB")
	if err := parseFlags(flags, args); err != nil {
		return err
	}
	if flags.NArg() != 1 || *tag != "" && flags.Arg(0) != "update" {
		return &UsageError{Err: fmt.Errorf("укажите pob-data status|ensure|update|rollback")}
	}
	manager, err := pob.Local()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, pob.InstallTimeout)
	defer cancel()
	if err := runPobData(ctx, manager, flags.Arg(0), *tag); err != nil {
		return err
	}
	return r.pobDataStatus(manager)
}

func runPobData(ctx context.Context, manager *pob.Manager, action, tag string) error {
	switch action {
	case "status":
		return nil
	case "ensure":
		_, err := manager.Ensure(ctx)
		return err
	case "update":
		return updatePobData(ctx, manager, tag)
	case "rollback":
		state, err := manager.Rollback(ctx, true)
		return pobDataOperationError(state, err)
	default:
		return &UsageError{Err: fmt.Errorf("неизвестное действие с PoB %q", action)}
	}
}

func updatePobData(ctx context.Context, manager *pob.Manager, tag string) error {
	if tag != "" {
		state, err := manager.UpdateTo(ctx, true, tag)
		return pobDataOperationError(state, err)
	}
	state, err := manager.Update(ctx, true)
	return pobDataOperationError(state, err)
}

func pobDataOperationError(state pob.State, err error) error {
	if err != nil {
		return err
	}
	switch state.Status {
	case "local_installation", "up_to_date", "updated", "rolled_back":
		return nil
	default:
		return fmt.Errorf("PoB: %s %s", state.Status, state.Message)
	}
}

func (r runner) pobDataStatus(manager *pob.Manager) error {
	current, previous, err := manager.Installations()
	if errors.Is(err, os.ErrNotExist) {
		return r.write(pobDataInstallations{})
	}
	if err != nil {
		return err
	}
	value := pobDataInstallations{Current: pobDataView(current)}
	if previous != nil {
		value.Previous = pobDataView(*previous)
	}
	return r.write(value)
}
