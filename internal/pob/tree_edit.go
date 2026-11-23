package pob

import (
	"context"
	"fmt"
	"slices"

	"github.com/Hecatoncheir/pob2_engine/internal/build"
)

// ResolveTreeEdit вычисляет изменение пассивов средствами оригинального PoB.
func (c *Calculator) ResolveTreeEdit(ctx context.Context, doc *build.Document, edit build.TreeNodeEdit) (build.PassiveChange, error) {
	model, err := doc.Model()
	if err != nil {
		return build.PassiveChange{}, err
	}
	xml, err := doc.Export()
	if err != nil {
		return build.PassiveChange{}, err
	}
	answer, err := c.run(ctx, &request{Operation: "edit_tree", XML: pobXML.Replace(string(xml)),
		NodeID: edit.NodeID, Action: edit.Action, Attribute: edit.Attribute, AllocationMode: edit.AllocationMode,
		TreeSpec: model.ActiveSets.Tree, TreeVersion: model.TreeVersion})
	if err != nil {
		return build.PassiveChange{}, err
	}
	if answer.Error != "" {
		return build.PassiveChange{}, fmt.Errorf("PoB Lua: %s", answer.Error)
	}
	if answer.TreeEdit == nil || !confirmsTreeEdit(*answer.TreeEdit, edit) {
		return build.PassiveChange{}, fmt.Errorf("PoB не подтвердил правку дерева")
	}
	return *answer.TreeEdit, nil
}

func confirmsTreeEdit(change build.PassiveChange, edit build.TreeNodeEdit) bool {
	if !confirmsTreeMode(change, edit) {
		return false
	}
	return confirmsTreeAction(change, edit)
}

func confirmsTreeAction(change build.PassiveChange, edit build.TreeNodeEdit) bool {
	attributeMatches := edit.Attribute == "" || slices.ContainsFunc(change.Attributes, func(attribute build.PassiveAttributeChange) bool {
		return attribute.NodeID == edit.NodeID && attribute.After == edit.Attribute
	})
	switch edit.Action {
	case "allocate":
		return len(change.Removed) == 0 && slices.Contains(change.Added, edit.NodeID) && attributeMatches
	case "refund":
		return len(change.Added) == 0 && slices.Contains(change.Removed, edit.NodeID)
	case "set_attribute":
		return len(change.Added)+len(change.Removed)+len(change.Allocations) == 0 && len(change.Attributes) == 1 && attributeMatches
	default:
		return false
	}
}

func confirmsTreeMode(change build.PassiveChange, edit build.TreeNodeEdit) bool {
	if edit.AllocationMode == "" {
		return len(change.Allocations) == 0 && change.Points == nil
	}
	if change.Points == nil {
		return false
	}
	if edit.Action == "set_attribute" {
		return true
	}
	allocation := build.PassiveAllocationChange{NodeID: edit.NodeID, After: edit.AllocationMode}
	if edit.Action == "refund" {
		allocation.Before, allocation.After = edit.AllocationMode, ""
	}
	return slices.Contains(change.Allocations, allocation)
}
