package pob

import (
	"context"
	"fmt"

	"github.com/Hecatoncheir/pob2_engine/internal/build"
)

func (c *Calculator) ReadTreeGeometry(ctx context.Context, doc *build.Document, scene bool) (build.TreeGeometry, error) {
	xml, err := doc.ExportWithoutProgression()
	if err != nil {
		return build.TreeGeometry{}, err
	}
	if !scene || c.treeScene.path == "" {
		return c.calculateTreeGeometry(ctx, xml, scene)
	}
	if err := ctx.Err(); err != nil {
		return build.TreeGeometry{}, err
	}
	key := c.treeScene.key(xml)
	if tree, ok := c.treeScene.read(key); ok {
		c.markTreeAssets(tree.Scene)
		return tree, nil
	}
	tree, err := c.calculateTreeGeometry(ctx, xml, true)
	if err == nil && ctx.Err() == nil {
		c.treeScene.write(key, tree)
	}
	return tree, err
}

func (c *Calculator) calculateTreeGeometry(ctx context.Context, xml []byte, scene bool) (build.TreeGeometry, error) {
	answer, err := c.run(ctx, &request{Operation: "tree_geometry", XML: pobXML.Replace(string(xml)), IncludeScene: scene})
	if err != nil {
		return build.TreeGeometry{}, err
	}
	if answer.Error != "" {
		return build.TreeGeometry{}, fmt.Errorf("PoB Lua: %s", answer.Error)
	}
	if answer.TreeGeometry == nil || len(answer.TreeGeometry.Nodes) == 0 {
		return build.TreeGeometry{}, fmt.Errorf("PoB не вернул геометрию дерева")
	}
	c.markTreeAssets(answer.TreeGeometry.Scene)
	return *answer.TreeGeometry, nil
}
