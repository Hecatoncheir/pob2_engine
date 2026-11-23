package pob

import (
	"context"
	"fmt"

	"github.com/Hecatoncheir/pob2_engine/internal/build"
)

func (c *Calculator) ReadTreePower(ctx context.Context, doc *build.Document, query build.TreePowerQuery) (*build.TreePower, error) {
	xml, err := doc.ExportWithoutProgression()
	if err != nil {
		return nil, err
	}
	input := &request{Operation: "tree_power", XML: pobXML.Replace(string(xml)), TreePower: &query}
	answer, err := c.runTreePower(ctx, input)
	if err != nil {
		return nil, err
	}
	if answer.Error != "" {
		return nil, fmt.Errorf("PoB Lua: %s", answer.Error)
	}
	return answer.TreePower, nil
}

func (c *Calculator) runTreePower(ctx context.Context, input *request) (response, error) {
	if input.TreePower.Metric == "" {
		return c.run(ctx, input)
	}
	release, err := acquireSlot(ctx, c.powerGate)
	if err != nil {
		return response{}, err
	}
	defer release()
	// Долгий расчёт использует отдельный Lua и не задерживает обычные правки.
	worker := Calculator{layout: c.layout}
	defer worker.stop()
	return worker.execute(ctx, input)
}
