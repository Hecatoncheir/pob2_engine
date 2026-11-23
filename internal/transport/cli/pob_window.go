package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Hecatoncheir/pob2_engine/internal/pobwindow"
)

var pobWindowMethods = map[string]string{
	"export":     "exportBuild",
	"characters": "listCharacters",
	"character":  "getCharacter",
}

type pobWindowOptions struct {
	command, state, directory, name string
	ensure                          bool
	wait                            time.Duration
}

func (r runner) pobWindow(ctx context.Context, args []string) error {
	flags := r.flags("pob-window")
	var options pobWindowOptions
	flags.StringVar(&options.state, "state-dir", "", "каталог конфигурации моста PoB")
	flags.StringVar(&options.directory, "pob-dir", "", "папка выбранной установки PoB")
	flags.StringVar(&options.name, "name", "", "имя персонажа для character")
	flags.BoolVar(&options.ensure, "ensure", false, "сохранить уже установленный мост")
	flags.DurationVar(&options.wait, "wait", 0, "время ожидания подключения в status")
	if err := parseFlags(flags, args); err != nil {
		return err
	}
	command, err := pobWindowCommand(flags.Args(), options)
	if err != nil {
		return err
	}
	options.command = command
	if options.state == "" {
		options.state, err = pobwindow.StateDirectory()
		if err != nil {
			return err
		}
	}
	return r.runPobWindow(ctx, options)
}

const pobWindowUsage = "укажите pob-window [ПАРАМЕТРЫ] discover|check|install|uninstall|status|export|build-directory|characters|character|wine-prefix"

func pobWindowCommand(args []string, options pobWindowOptions) (string, error) {
	if len(args) != 1 {
		return "", &UsageError{Err: fmt.Errorf("%s", pobWindowUsage)}
	}
	command := args[0]
	if !validPobWindowOptions(command, options) {
		return "", &UsageError{Err: fmt.Errorf("неверные параметры команды окна PoB %q", command)}
	}
	return command, nil
}

func validPobWindowOptions(command string, options pobWindowOptions) bool {
	if (command == "character") != (options.name != "") || options.ensure && command != "install" {
		return false
	}
	if options.wait < 0 || options.wait > 2*time.Minute || options.wait > 0 && command != "status" {
		return false
	}
	switch command {
	case "check", "install", "uninstall", "wine-prefix":
		return options.directory != ""
	case "discover", "status", "build-directory":
		return true
	default:
		return pobWindowMethods[command] != ""
	}
}

func (r runner) runPobWindow(ctx context.Context, options pobWindowOptions) error {
	switch options.command {
	case "discover":
		return r.write(pobwindow.Discover(options.state))
	case "check":
		return pobwindow.CheckBridge(options.state, options.directory)
	case "install":
		result, err := pobwindow.InstallBridge(options.state, options.directory, options.ensure)
		if err != nil {
			return err
		}
		return r.write(result)
	case "uninstall":
		result, err := pobwindow.UninstallBridge(options.state, options.directory)
		if err != nil {
			return err
		}
		return r.write(result)
	case "wine-prefix":
		prefix, err := pobwindow.WinePrefix(options.directory, "")
		if err != nil {
			return err
		}
		return r.write(map[string]string{"prefix": prefix})
	default:
		return r.readPobWindow(ctx, options)
	}
}

func (r runner) readPobWindow(ctx context.Context, options pobWindowOptions) error {
	client, err := dialPobWindow(ctx, options)
	if err != nil {
		return err
	}
	defer client.Close()
	switch options.command {
	case "status":
		return r.write(client.Status())
	case "build-directory":
		location, err := client.BuildDirectory(ctx)
		if err != nil {
			return err
		}
		return r.write(map[string]string{"directory": location})
	}
	var arguments map[string]any
	if options.command == "character" {
		arguments = map[string]any{"name": options.name}
	}
	result, err := client.Call(ctx, pobWindowMethods[options.command], arguments)
	if err != nil {
		return err
	}
	return r.write(result)
}

func dialPobWindow(ctx context.Context, options pobWindowOptions) (*pobwindow.Client, error) {
	if options.wait == 0 {
		return pobwindow.Dial(ctx, options.state, options.directory)
	}
	wait, cancel := context.WithTimeout(ctx, options.wait)
	defer cancel()
	retry := time.NewTicker(250 * time.Millisecond)
	defer retry.Stop()
	for {
		client, err := pobwindow.Dial(wait, options.state, options.directory)
		if err == nil {
			return client, nil
		}
		select {
		case <-wait.Done():
			return nil, err
		case <-retry.C:
		}
	}
}

func (r runner) pobWindowControl(args []string) error {
	flags := r.flags("pob-window-control")
	executable := flags.String("executable", "", "имя исполняемого файла окна PoB")
	if err := parseFlags(flags, args); err != nil {
		return err
	}
	if *executable == "" || flags.NArg() != 1 {
		return &UsageError{Err: fmt.Errorf("укажите pob-window-control --executable ИМЯ check|minimize|restore|close")}
	}
	found, err := pobwindow.ControlWineWindow(flags.Arg(0), *executable, os.Getenv("POB2_UI_WINE_OWNER"))
	if err != nil {
		return err
	}
	return r.write(map[string]bool{"found": found})
}
