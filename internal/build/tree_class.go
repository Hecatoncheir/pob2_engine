package build

import "fmt"

type TreeClassEdit struct {
	ClassID      int    `json:"class_id"`
	AscendancyID int    `json:"ascendancy_id"`
	Mode         string `json:"mode,omitempty"`
}

type TreeClassChange struct {
	Identity   TreeIdentity      `json:"identity"`
	Spec       map[string]string `json:"spec"`
	Attributes map[string]string `json:"attributes"`
	WeaponOne  string            `json:"weapon_one"`
	WeaponTwo  string            `json:"weapon_two"`
	URL        string            `json:"url"`
}

func (e TreeClassEdit) validate() error {
	if e.ClassID < 0 || e.ClassID > 1000 || e.AscendancyID < 0 || e.AscendancyID > 1000 {
		return fmt.Errorf("неверный идентификатор класса или восхождения")
	}
	switch e.Mode {
	case "", "connect", "reset":
		return nil
	default:
		return fmt.Errorf("режим смены класса: connect или reset")
	}
}

// WithResolvedTreeClass переносит из Lua только поля текущей спецификации.
func (d *Document) WithResolvedTreeClass(change TreeClassChange) (*Document, error) {
	updated, err := d.copyForEdit()
	if err != nil {
		return nil, err
	}
	tree, err := updated.root.child("Tree")
	if err != nil {
		return nil, err
	}
	if tree == nil {
		return nil, fmt.Errorf("в XML нет дерева")
	}
	spec, _, err := activeTreeSpec(tree)
	if err != nil {
		return nil, err
	}
	if err := updated.setTreeIdentity(change.Identity); err != nil {
		return nil, err
	}
	if err := applyTreeClassSpec(spec, change); err != nil {
		return nil, err
	}
	updated.clearCachedResults()
	return updated, nil
}

func applyTreeClassSpec(spec *element, change TreeClassChange) error {
	for _, name := range []string{"classId", "ascendClassId", "classInternalId", "ascendancyInternalId", "secondaryAscendClassId", "nodes", "masteryEffects"} {
		value, ok := change.Spec[name]
		if !ok {
			return fmt.Errorf("PoB пропустил поле спецификации %s", name)
		}
		spec.set(name, value)
	}
	if err := setClassWeaponNodes(spec, "WeaponSet1", change.WeaponOne); err != nil {
		return err
	}
	if err := setClassWeaponNodes(spec, "WeaponSet2", change.WeaponTwo); err != nil {
		return err
	}
	if err := setClassAttributeNodes(spec, change.Attributes); err != nil {
		return err
	}
	url, err := treeClassChild(spec, "URL")
	if err != nil {
		return err
	}
	url.parts = []part{{text: change.URL}}
	return nil
}

func setClassWeaponNodes(spec *element, name, nodes string) error {
	weapon, err := spec.child(name)
	if err != nil || (weapon == nil && nodes == "") {
		return err
	}
	if weapon == nil {
		weapon = newElement(name)
		spec.parts = append(spec.parts, part{element: weapon})
	}
	weapon.set("nodes", nodes)
	return nil
}

func setClassAttributeNodes(spec *element, values map[string]string) error {
	overrides, err := treeClassChild(spec, "Overrides")
	if err != nil {
		return err
	}
	attributes, err := treeClassChild(overrides, "AttributeOverride")
	if err != nil {
		return err
	}
	for _, field := range attributeFields {
		attributes.set(field.name, values[field.name])
	}
	return nil
}

func treeClassChild(parent *element, name string) (*element, error) {
	child, err := parent.child(name)
	if err != nil {
		return nil, err
	}
	if child == nil {
		child = newElement(name)
		parent.parts = append(parent.parts, part{element: child})
	}
	return child, nil
}
