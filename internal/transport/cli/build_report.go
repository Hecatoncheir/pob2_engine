package cli

import (
	"context"
	"fmt"
)

func (r runner) buildReport(ctx context.Context, args []string) error {
	fs := r.flags("build-report")
	connection := connectFlagsWithTimeout(fs, buildTimeout)
	id := fs.String("session", "", "явно выбранный session_id")
	revision := fs.Uint64("revision", 0, "ожидаемая ревизия сессии")
	if err := parse(fs, args); err != nil {
		return err
	}
	if *id == "" {
		return &UsageError{Err: fmt.Errorf("нужна --session")}
	}
	ctx, cancel := context.WithTimeout(ctx, connection.timeout)
	defer cancel()
	client, err := connection.dial(ctx)
	if err != nil {
		return err
	}
	defer client.Close()
	view, err := client.BuildReport(ctx, *id, *revision)
	if err != nil {
		return err
	}
	return r.write(view)
}
