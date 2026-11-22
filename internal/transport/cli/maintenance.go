package cli

import (
	"context"
	"fmt"
	"time"

	notices "github.com/Hecatoncheir/pob2_engine"
	"github.com/Hecatoncheir/pob2_engine/internal/maintenance"
)

func (r runner) maintenance(ctx context.Context, command string, args []string) error {
	fs := r.flags(command)
	timeout := 5 * time.Second
	if command == "update" {
		timeout = maintenance.UpdateTimeout
	}
	connection := connectFlagsWithTimeout(fs, timeout)
	connection.allowMemory = true
	var check, licenses bool
	target := "all"
	switch command {
	case "update":
		fs.BoolVar(&check, "check", false, "проверить наличие обновления без записи")
		fs.StringVar(&target, "target", "all", "что обновлять: all, engine или pob")
	case "doctor":
		fs.BoolVar(&licenses, "licenses", false, "вывести тексты встроенных лицензий")
	}
	if err := parse(fs, args); err != nil {
		return err
	}
	if err := validateMaintenance(connection, licenses); err != nil {
		return err
	}
	if licenses {
		files, err := notices.Files()
		if err != nil {
			return err
		}
		return r.write(files)
	}
	ctx, cancel := context.WithTimeout(ctx, connection.timeout)
	defer cancel()
	return r.runMaintenance(ctx, command, connection, !check, target)
}

func validateMaintenance(c *connection, licenses bool) error {
	switch {
	case c.timeout <= 0:
		return &UsageError{Err: fmt.Errorf("таймаут должен быть положительным")}
	case c.startServer:
		return &UsageError{Err: fmt.Errorf("команда обслуживания не запускает сервер")}
	case c.profile == "" && (c.baseDir != "" || c.dataDir != ""):
		return &UsageError{Err: fmt.Errorf("для каталогов нужен --profile")}
	case licenses && (c.profile != "" || c.address != ""):
		return &UsageError{Err: fmt.Errorf("--licenses выводит только лицензии локального бинарника")}
	default:
		return nil
	}
}

func (r runner) runMaintenance(ctx context.Context, command string, connection *connection, apply bool, target string) error {
	var commands maintenance.Commands = maintenance.Local()
	if connection.address != "" || connection.profile != "" {
		client, err := connection.dial(ctx)
		if err != nil {
			return err
		}
		defer client.Close()
		commands = client
	}
	if command == "update" {
		result, err := commands.Update(ctx, apply, target)
		if err != nil {
			return err
		}
		return r.write(result)
	}
	report, err := commands.Doctor(ctx)
	if err != nil {
		return err
	}
	return r.write(report)
}
