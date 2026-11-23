package build

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

type TreeNodeEdit struct {
	NodeID int    `json:"node_id"`
	Action string `json:"action"`
}

func (e TreeNodeEdit) validate() error {
	if e.NodeID < 1 || e.NodeID > 1000000 {
		return fmt.Errorf("неверный идентификатор пассива")
	}
	if e.Action != "allocate" && e.Action != "refund" {
		return fmt.Errorf("действие дерева: allocate или refund")
	}
	return nil
}

// PassiveChange описывает узлы, которые оригинальный PoB добавит или снимет.
type PassiveChange struct {
	Added   []int `json:"added,omitempty"`
	Removed []int `json:"removed,omitempty"`
}

// ResolvedTreeEdit содержит изменения активной спецификации, подтверждённые Lua.
type ResolvedTreeEdit struct {
	PassiveChange
	Strength     []int `json:"strength,omitempty"`
	Dexterity    []int `json:"dexterity,omitempty"`
	Intelligence []int `json:"intelligence,omitempty"`
}

func (d *Document) WithResolvedTreeEdit(change ResolvedTreeEdit) (*Document, error) {
	updated, err := d.copyForEdit()
	if err != nil {
		return nil, err
	}
	if err := updated.applyResolvedTreeEdit(change); err != nil {
		return nil, err
	}
	updated.clearCachedResults()
	if _, err := updated.Export(); err != nil {
		return nil, err
	}
	return updated, nil
}

func (d *Document) applyResolvedTreeEdit(change ResolvedTreeEdit) error {
	model, err := d.Model()
	if err != nil {
		return err
	}
	if model.ActiveSets.Tree == 0 {
		return fmt.Errorf("в сборке нет активной спецификации дерева")
	}
	tree, _ := d.root.child("Tree")
	spec := tree.children("Spec")[model.ActiveSets.Tree-1]
	raw := spec.attr("nodes", "-")
	if raw == "-" {
		return fmt.Errorf("в XML дерева нет явного списка пассивов")
	}
	previous, err := parsePassiveNodes(raw)
	if err != nil {
		return err
	}
	nodes, err := updatePassiveNodes(previous, change.PassiveChange)
	if err != nil {
		return err
	}
	spec.set("nodes", formatPassiveNodes(nodes))
	return updateAttributeOverrides(spec, change)
}

func updatePassiveNodes(nodes []int, change PassiveChange) ([]int, error) {
	if len(change.Added)+len(change.Removed) == 0 {
		return nil, fmt.Errorf("PoB не изменил дерево")
	}
	for _, id := range change.Removed {
		if !slices.Contains(nodes, id) {
			return nil, fmt.Errorf("снимаемый пассив %d отсутствует в XML", id)
		}
	}
	for _, id := range change.Added {
		if slices.Contains(nodes, id) || id < 1 || id > 1000000 {
			return nil, fmt.Errorf("неверный добавляемый пассив %d", id)
		}
	}
	updated := slices.DeleteFunc(slices.Clone(nodes), func(id int) bool { return slices.Contains(change.Removed, id) })
	return append(updated, change.Added...), nil
}

func updateAttributeOverrides(spec *element, change ResolvedTreeEdit) error {
	hasNew := len(change.Strength)+len(change.Dexterity)+len(change.Intelligence) > 0
	overrides, err := spec.child("Overrides")
	if err != nil || (overrides == nil && !hasNew) {
		return err
	}
	if overrides == nil {
		overrides = newElement("Overrides")
		spec.parts = append(spec.parts, part{element: overrides})
	}
	attributes, err := overrides.child("AttributeOverride")
	if err != nil || (attributes == nil && !hasNew) {
		return err
	}
	if attributes == nil {
		attributes = newElement("AttributeOverride")
		overrides.parts = append(overrides.parts, part{element: attributes})
	}
	return setAttributeLists(attributes, change)
}

func setAttributeLists(attributes *element, change ResolvedTreeEdit) error {
	for _, entry := range []struct {
		name string
		ids  []int
	}{{"strNodes", change.Strength}, {"dexNodes", change.Dexterity}, {"intNodes", change.Intelligence}} {
		if err := changeAttributeList(attributes, entry.name, entry.ids, change); err != nil {
			return err
		}
	}
	return nil
}

func changeAttributeList(node *element, name string, added []int, change ResolvedTreeEdit) error {
	old, err := parsePassiveNodes(node.attr(name, ""))
	if err != nil {
		return err
	}
	current := slices.DeleteFunc(slices.Clone(old), func(id int) bool {
		return slices.Contains(change.Removed, id) || slices.Contains(change.Added, id)
	})
	current = append(current, added...)
	if len(added) > 0 || !slices.Equal(old, current) || node.attr(name, "") == "" {
		node.set(name, formatPassiveNodes(current))
	}
	return nil
}

func formatPassiveNodes(nodes []int) string {
	ids := make([]string, 0, len(nodes))
	for _, id := range nodes {
		ids = append(ids, strconv.Itoa(id))
	}
	return strings.Join(ids, ",")
}
