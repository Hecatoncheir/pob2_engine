package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Hecatoncheir/pob2_engine/internal/pob"
	"github.com/Hecatoncheir/pob2_engine/internal/transport/cli"
)

func main() { os.Exit(run()) }

func run() int {
	if len(os.Args) > 1 && os.Args[1] == "_pob-lua" {
		return exitCode(pob.RunWorker(os.Args[2:]))
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	return exitCode(cli.Run(ctx, os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func exitCode(err error) int {
	if err == nil {
		return 0
	}
	fmt.Fprintln(os.Stderr, err)
	var usage *cli.UsageError
	if errors.As(err, &usage) {
		return 2
	}
	var result *cli.ResultStatusError
	if errors.As(err, &result) {
		if result.Status == "partial" {
			return 4
		}
		return 3
	}
	return 1
}
