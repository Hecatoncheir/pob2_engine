package pob

import (
	"context"
	"fmt"
	"slices"

	"github.com/Hecatoncheir/pob2_engine/internal/build"
)

// ResolveTreeClass получает состояние спецификации после смены класса в Lua.
func (c *Calculator) ResolveTreeClass(ctx context.Context, doc *build.Document, edit build.TreeClassEdit) (build.TreeClassChange, error) {
	model, err := doc.Model()
	if err != nil {
		return build.TreeClassChange{}, err
	}
	xml, err := doc.ExportWithoutProgression()
	if err != nil {
		return build.TreeClassChange{}, err
	}
	answer, err := c.run(ctx, &request{Operation: "edit_tree_class", XML: pobXML.Replace(string(xml)),
		TreeSpec: model.ActiveSets.Tree, TreeVersion: model.TreeVersion, TreeClass: &edit})
	if err != nil {
		return build.TreeClassChange{}, err
	}
	if answer.Error != "" {
		return build.TreeClassChange{}, fmt.Errorf("PoB Lua: %s", answer.Error)
	}
	if answer.TreeClass == nil {
		return build.TreeClassChange{}, fmt.Errorf("PoB не подтвердил смену класса")
	}
	return *answer.TreeClass, nil
}

// ResolveTreeSpec подтверждает спецификацию и получает имена её класса из Lua.
func (c *Calculator) ResolveTreeSpec(ctx context.Context, doc *build.Document) (build.TreeIdentity, error) {
	model, err := doc.Model()
	if err != nil {
		return build.TreeIdentity{}, err
	}
	xml, err := doc.ExportWithoutProgression()
	if err != nil {
		return build.TreeIdentity{}, err
	}
	answer, err := c.run(ctx, &request{Operation: "select_tree_spec", XML: pobXML.Replace(string(xml)),
		TreeSpec: model.ActiveSets.Tree, TreeVersion: model.TreeVersion})
	if err != nil {
		return build.TreeIdentity{}, err
	}
	if answer.Error != "" {
		return build.TreeIdentity{}, fmt.Errorf("PoB Lua: %s", answer.Error)
	}
	if answer.TreeIdentity == nil {
		return build.TreeIdentity{}, fmt.Errorf("PoB не подтвердил спецификацию")
	}
	return *answer.TreeIdentity, nil
}

// ResolveTreeEdit вычисляет изменение пассивов средствами оригинального PoB.
func (c *Calculator) ResolveTreeEdit(ctx context.Context, doc *build.Document, edit build.TreeNodeEdit) (build.PassiveChange, error) {
	model, err := doc.Model()
	if err != nil {
		return build.PassiveChange{}, err
	}
	xml, err := doc.ExportWithoutProgression()
	if err != nil {
		return build.PassiveChange{}, err
	}
	release, err := c.acquire(ctx)
	if err != nil {
		return build.PassiveChange{}, err
	}
	defer release()
	answer, err := c.execute(ctx, &request{Operation: "edit_tree", XML: pobXML.Replace(string(xml)),
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
	if edit.Action == "allocate" || edit.Action == "refund" {
		c.rememberTreeReportXML(doc, *answer.TreeEdit)
	}
	return *answer.TreeEdit, nil
}

func (c *Calculator) rememberTreeReportXML(doc *build.Document, change build.PassiveChange) {
	// Сервис сохраняет XML через то же преобразование подтверждённой правки.
	updated, err := doc.WithResolvedTreeEdit(change)
	if err != nil {
		return
	}
	xml, err := updated.ExportWithoutProgression()
	if err == nil {
		c.treeReportXML = pobXML.Replace(string(xml))
	}
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
