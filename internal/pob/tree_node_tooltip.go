package pob

import (
	"context"
	"fmt"

	"github.com/Hecatoncheir/pob2_engine/internal/build"
)

func (c *Calculator) ReadTreeNodeTooltip(ctx context.Context, doc *build.Document, node uint32, compare, scene bool) (*build.TreeNodeTooltip, error) {
	xml, err := doc.ExportWithoutProgression()
	if err != nil {
		return nil, err
	}
	answer, err := c.run(ctx, &request{
		Operation: "tree_node_tooltip", XML: pobXML.Replace(string(xml)), NodeID: int(node), Compare: compare, IncludeScene: scene,
	})
	if err != nil {
		return nil, err
	}
	if answer.Error != "" {
		return nil, fmt.Errorf("PoB Lua: %s", answer.Error)
	}
	if answer.TreeTooltip != nil {
		c.markTreeAssets(answer.TreeTooltip.Scene)
	}
	return answer.TreeTooltip, nil
}
