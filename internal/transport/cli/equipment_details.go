package cli

import (
	"context"
	"fmt"
	"math"

	"github.com/Hecatoncheir/pob2_engine/internal/build"
)

func (r runner) equipmentDetails(ctx context.Context, args []string) error {
	fs := r.flags("equipment-details")
	connection := connectFlagsWithTimeout(fs, buildTimeout)
	id := fs.String("session", "", "явно выбранный session_id")
	revision := fs.Uint64("revision", 0, "ожидаемая ревизия сессии")
	slot := fs.String("slot", "", "слот экипировки")
	itemID := fs.Uint("item-id", 0, "ID предмета из сборки")
	file := fs.String("file", "", "текст предмета; - для stdin")
	compare := fs.Bool("compare", false, "включить подробное сравнение")
	if err := parse(fs, args); err != nil {
		return err
	}
	if *id == "" || *itemID > math.MaxUint32 || (*itemID == 0) == (*file == "") {
		return &UsageError{Err: fmt.Errorf("нужны --session и один источник предмета: --item-id или --file")}
	}
	query := build.EquipmentDetailsQuery{Slot: *slot, ItemID: uint32(*itemID), Compare: *compare}
	if *file != "" {
		data, err := r.readInput(ctx, *file)
		if err != nil {
			return err
		}
		query.Raw = string(data)
	}
	ctx, cancel := context.WithTimeout(ctx, connection.timeout)
	defer cancel()
	client, err := connection.dial(ctx)
	if err != nil {
		return err
	}
	defer client.Close()
	view, err := client.EquipmentDetails(ctx, *id, *revision, query)
	if err != nil {
		return err
	}
	return r.write(view)
}
