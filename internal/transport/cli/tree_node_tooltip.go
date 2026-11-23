package cli

import (
	"context"
	"fmt"
	"math"
)

func (r runner) treeNodeTooltip(ctx context.Context, args []string) error {
	fs := r.flags("tree-node-tooltip")
	connection := connectFlagsWithTimeout(fs, buildTimeout)
	id := fs.String("session", "", "явно выбранный session_id")
	revision := fs.Uint64("revision", 0, "ожидаемая ревизия сессии")
	node := fs.Uint("node", 0, "ID узла")
	compare := fs.Bool("compare", false, "показать изменение расчёта от узла")
	scene := fs.Bool("scene", false, "вернуть штатную подсветку пути")
	if err := parse(fs, args); err != nil {
		return err
	}
	if *id == "" || *node == 0 || *node > math.MaxUint32 {
		return &UsageError{Err: fmt.Errorf("нужны --session и допустимый --node")}
	}
	ctx, cancel := context.WithTimeout(ctx, connection.timeout)
	defer cancel()
	client, err := connection.dial(ctx)
	if err != nil {
		return err
	}
	defer client.Close()
	view, err := client.TreeNodeTooltip(ctx, *id, *revision, uint32(*node), *compare, *scene)
	if err != nil {
		return err
	}
	return r.write(view)
}
