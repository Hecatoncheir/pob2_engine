package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/Hecatoncheir/pob2_engine/internal/build"
)

func (r runner) treePower(ctx context.Context, args []string) error {
	fs := r.flags("tree-power")
	connection := connectFlagsWithTimeout(fs, 2*time.Minute)
	id := fs.String("session", "", "явно выбранный session_id")
	revision := fs.Uint64("revision", 0, "ожидаемая ревизия сессии")
	metric := fs.String("metric", "", "характеристика; пустая строка выводит каталог")
	distance := fs.Uint("max-distance", 0, "расстояние от 1 до 200; ноль без ограничения")
	allocation := fs.String("allocation", "unallocated", "all, allocated или unallocated")
	if err := parse(fs, args); err != nil {
		return err
	}
	if *id == "" || *distance > 200 {
		return &UsageError{Err: fmt.Errorf("нужны --session и допустимое расстояние до узла")}
	}
	ctx, cancel := context.WithTimeout(ctx, connection.timeout)
	defer cancel()
	client, err := connection.dial(ctx)
	if err != nil {
		return err
	}
	defer client.Close()
	query := build.TreePowerQuery{Metric: *metric, MaxDistance: uint32(*distance), Allocation: *allocation}
	view, err := client.TreePower(ctx, *id, *revision, query)
	if err != nil {
		return err
	}
	return r.write(view)
}
