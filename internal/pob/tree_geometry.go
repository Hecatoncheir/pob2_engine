package pob

import (
	"context"
	"fmt"

	"github.com/Hecatoncheir/pob2_engine/internal/build"
)

func (c *Calculator) ReadTreeGeometry(ctx context.Context, doc *build.Document) (build.TreeGeometry, error) {
	xml, err := doc.ExportWithoutProgression()
	if err != nil {
		return build.TreeGeometry{}, err
	}
	answer, err := c.run(ctx, &request{Operation: "tree_geometry", XML: pobXML.Replace(string(xml))})
	if err != nil {
		return build.TreeGeometry{}, err
	}
	if answer.Error != "" {
		return build.TreeGeometry{}, fmt.Errorf("PoB Lua: %s", answer.Error)
	}
	if answer.TreeGeometry == nil || len(answer.TreeGeometry.Nodes) == 0 {
		return build.TreeGeometry{}, fmt.Errorf("PoB не вернул геометрию дерева")
	}
	return *answer.TreeGeometry, nil
}
