package pob

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Hecatoncheir/pob2_engine/internal/build"
)

func (c *Calculator) ReadEquipmentDetails(ctx context.Context, doc *build.Document, query build.EquipmentDetailsQuery) (json.RawMessage, error) {
	xml, err := doc.ExportWithoutProgression()
	if err != nil {
		return nil, err
	}
	answer, err := c.run(ctx, &request{
		Operation: "equipment_details", XML: pobXML.Replace(string(xml)), EquipmentDetails: &query,
	})
	if err != nil {
		return nil, err
	}
	if answer.Error != "" {
		return nil, fmt.Errorf("PoB Lua: %s", answer.Error)
	}
	if string(answer.EquipmentDetails) == "null" {
		return nil, nil
	}
	return answer.EquipmentDetails, nil
}
