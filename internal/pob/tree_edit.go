package pob

import (
	"context"
	"fmt"
	"slices"

	"github.com/Hecatoncheir/pob2_engine/internal/build"
)

// ResolveTreeEdit вычисляет изменение пассивов средствами оригинального PoB.
func (c *Calculator) ResolveTreeEdit(ctx context.Context, doc *build.Document, edit build.TreeNodeEdit) (build.ResolvedTreeEdit, error) {
	model, err := doc.Model()
	if err != nil {
		return build.ResolvedTreeEdit{}, err
	}
	xml, err := doc.Export()
	if err != nil {
		return build.ResolvedTreeEdit{}, err
	}
	answer, err := c.run(ctx, &request{Operation: "edit_tree", XML: pobXML.Replace(string(xml)),
		NodeID: edit.NodeID, Action: edit.Action, TreeSpec: model.ActiveSets.Tree, TreeVersion: model.TreeVersion})
	if err != nil {
		return build.ResolvedTreeEdit{}, err
	}
	if answer.Error != "" {
		return build.ResolvedTreeEdit{}, fmt.Errorf("PoB Lua: %s", answer.Error)
	}
	if answer.TreeEdit == nil || !confirmsTreeEdit(*answer.TreeEdit, edit) {
		return build.ResolvedTreeEdit{}, fmt.Errorf("PoB не подтвердил правку дерева")
	}
	return *answer.TreeEdit, nil
}

func confirmsTreeEdit(change build.ResolvedTreeEdit, edit build.TreeNodeEdit) bool {
	if edit.Action == "allocate" {
		return len(change.Removed) == 0 && slices.Contains(change.Added, edit.NodeID)
	}
	return len(change.Added) == 0 && slices.Contains(change.Removed, edit.NodeID)
}
