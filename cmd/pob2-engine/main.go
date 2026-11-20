package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Hecatoncheir/pob2_engine/internal/transport/cli"
)

func main() { os.Exit(run()) }

func run() int {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	if err := cli.Run(ctx, os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		var usage *cli.UsageError
		if errors.As(err, &usage) {
			return 2
		}
		return 1
	}
	return 0
}
