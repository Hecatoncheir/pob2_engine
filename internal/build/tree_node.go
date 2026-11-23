package build

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

type TreeNodeEdit struct {
	NodeID    int              `json:"node_id"`
	Action    string           `json:"action"`
	Attribute PassiveAttribute `json:"attribute,omitempty"`
}

func (e TreeNodeEdit) validate() error {
	if e.NodeID < 1 || e.NodeID > 1000000 {
		return fmt.Errorf("неверный идентификатор пассива")
	}
	switch e.Action {
	case "allocate":
		if e.Attribute == "" || e.Attribute.valid() {
			return nil
		}
	case "set_attribute":
		if e.Attribute.valid() {
			return nil
		}
	case "refund":
		if e.Attribute == "" {
			return nil
		}
		return fmt.Errorf("при снятии пассива атрибут не указывается")
	default:
		return fmt.Errorf("действие дерева: allocate, refund или set_attribute")
	}
	return fmt.Errorf("атрибут пассива: strength, dexterity или intelligence")
}

type PassiveAttribute string

const (
	Strength     PassiveAttribute = "strength"
	Dexterity    PassiveAttribute = "dexterity"
	Intelligence PassiveAttribute = "intelligence"
)

func (a PassiveAttribute) valid() bool {
	return a == Strength || a == Dexterity || a == Intelligence
}

type PassiveAttributeChange struct {
	NodeID int              `json:"node_id"`
	Before PassiveAttribute `json:"before,omitempty"`
	After  PassiveAttribute `json:"after"`
}

// PassiveChange описывает изменение активной спецификации, подтверждённое PoB.
type PassiveChange struct {
	Added      []int                    `json:"added,omitempty"`
	Removed    []int                    `json:"removed,omitempty"`
	Attributes []PassiveAttributeChange `json:"attributes,omitempty"`
}

func (d *Document) WithResolvedTreeEdit(change PassiveChange) (*Document, error) {
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

func (d *Document) applyResolvedTreeEdit(change PassiveChange) error {
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
	nodes, err := updatePassiveNodes(previous, change)
	if err != nil {
		return err
	}
	if len(change.Added)+len(change.Removed) > 0 {
		spec.set("nodes", formatPassiveNodes(nodes))
	}
	if err := validateAttributeChanges(nodes, change.Attributes); err != nil {
		return err
	}
	return updateAttributeOverrides(spec, change)
}

func updatePassiveNodes(nodes []int, change PassiveChange) ([]int, error) {
	if len(change.Added)+len(change.Removed)+len(change.Attributes) == 0 {
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

func validateAttributeChanges(nodes []int, changes []PassiveAttributeChange) error {
	for _, change := range changes {
		if !change.After.valid() || !slices.Contains(nodes, change.NodeID) {
			return fmt.Errorf("PoB вернул неверный атрибут узла %d", change.NodeID)
		}
	}
	return nil
}

func updateAttributeOverrides(spec *element, change PassiveChange) error {
	hasNew := len(change.Attributes) > 0
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

var attributeFields = []struct {
	name  string
	value PassiveAttribute
}{{"strNodes", Strength}, {"dexNodes", Dexterity}, {"intNodes", Intelligence}}

func setAttributeLists(attributes *element, change PassiveChange) error {
	updated := map[PassiveAttribute][]int{}
	clear := slices.Concat(change.Added, change.Removed)
	for _, attribute := range change.Attributes {
		clear = append(clear, attribute.NodeID)
		updated[attribute.After] = append(updated[attribute.After], attribute.NodeID)
	}
	for _, field := range attributeFields {
		if err := changeAttributeList(attributes, field.name, clear, updated[field.value]); err != nil {
			return err
		}
	}
	return nil
}

func changeAttributeList(node *element, name string, clear, added []int) error {
	old, err := parsePassiveNodes(node.attr(name, ""))
	if err != nil {
		return err
	}
	current := slices.DeleteFunc(slices.Clone(old), func(id int) bool {
		return slices.Contains(clear, id)
	})
	current = append(current, added...)
	if len(added) > 0 || !slices.Equal(old, current) || node.attr(name, "") == "" {
		node.set(name, formatPassiveNodes(current))
	}
	return nil
}

// readTreeAttributes показывает сохранённые выборы назначенных узлов.
func readTreeAttributes(spec *element, nodes []int) map[string]PassiveAttribute {
	overrides, err := spec.child("Overrides")
	if err != nil || overrides == nil {
		return nil
	}
	attributes, err := overrides.child("AttributeOverride")
	if err != nil || attributes == nil {
		return nil
	}
	return attributeSelections(attributes, nodes)
}

func attributeSelections(attributes *element, nodes []int) map[string]PassiveAttribute {
	result := make(map[string]PassiveAttribute)
	for _, field := range attributeFields {
		ids, err := parsePassiveNodes(attributes.attr(field.name, ""))
		if err != nil || !addAttributeSelections(result, nodes, ids, field.value) {
			return nil
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func addAttributeSelections(result map[string]PassiveAttribute, nodes, ids []int, value PassiveAttribute) bool {
	for _, id := range ids {
		if !slices.Contains(nodes, id) {
			continue
		}
		key := strconv.Itoa(id)
		if result[key] != "" {
			return false
		}
		result[key] = value
	}
	return true
}

func formatPassiveNodes(nodes []int) string {
	ids := make([]string, 0, len(nodes))
	for _, id := range nodes {
		ids = append(ids, strconv.Itoa(id))
	}
	return strings.Join(ids, ",")
}
