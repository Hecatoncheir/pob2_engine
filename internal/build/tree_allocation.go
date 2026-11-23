package build

import (
	"fmt"
	"slices"
	"strconv"
)

type TreeAllocationMode string

const (
	AllocationCommon     TreeAllocationMode = "common"
	AllocationWeaponSet1 TreeAllocationMode = "weapon_set_1"
	AllocationWeaponSet2 TreeAllocationMode = "weapon_set_2"
)

func (mode TreeAllocationMode) valid() bool {
	return mode == AllocationCommon || mode == AllocationWeaponSet1 || mode == AllocationWeaponSet2
}

type PassiveAllocationChange struct {
	NodeID int                `json:"node_id"`
	Before TreeAllocationMode `json:"before,omitempty"`
	After  TreeAllocationMode `json:"after,omitempty"`
}

type PassivePoints struct {
	Passives   int `json:"passives"`
	WeaponSet1 int `json:"weapon_set_1"`
	WeaponSet2 int `json:"weapon_set_2"`
}

type PassivePointsChange struct {
	Before PassivePoints `json:"before"`
	After  PassivePoints `json:"after"`
}

var weaponAllocationFields = []struct {
	name string
	mode TreeAllocationMode
}{{"WeaponSet1", AllocationWeaponSet1}, {"WeaponSet2", AllocationWeaponSet2}}

func readTreeAllocations(spec *element, nodes []int) (map[string]TreeAllocationMode, error) {
	values := make(map[string]TreeAllocationMode)
	for _, field := range weaponAllocationFields {
		ids, err := readWeaponNodes(spec, field.name)
		if err != nil {
			return nil, err
		}
		if err := storeWeaponAllocations(values, nodes, ids, field.mode); err != nil {
			return nil, err
		}
	}
	if len(values) == 0 {
		return nil, nil
	}
	return values, nil
}

func readWeaponNodes(spec *element, name string) ([]int, error) {
	branch, err := spec.child(name)
	if err != nil || branch == nil {
		return nil, err
	}
	return parsePassiveNodes(branch.attr("nodes", ""))
}

func storeWeaponAllocations(values map[string]TreeAllocationMode, nodes, ids []int, mode TreeAllocationMode) error {
	for _, id := range ids {
		if !slices.Contains(nodes, id) {
			continue
		}
		key := strconv.Itoa(id)
		if values[key] != "" {
			return fmt.Errorf("пассив %d указан в обеих ветвях оружия", id)
		}
		values[key] = mode
	}
	return nil
}

func updateTreeAllocations(spec *element, previous, nodes []int, change PassiveChange) error {
	if len(change.Allocations) == 0 {
		return nil
	}
	before, err := readTreeAllocations(spec, previous)
	if err != nil {
		return err
	}
	if err := validateAllocationChanges(before, previous, nodes, change); err != nil {
		return err
	}
	for _, field := range weaponAllocationFields {
		if err := writeWeaponBranch(spec, field.name, field.mode, change.Allocations); err != nil {
			return err
		}
	}
	return nil
}

func validateAllocationChanges(before map[string]TreeAllocationMode, previous, nodes []int, change PassiveChange) error {
	seen := make(map[int]bool, len(change.Allocations))
	for _, allocation := range change.Allocations {
		if seen[allocation.NodeID] || !validAllocationChange(before, previous, nodes, allocation) {
			return fmt.Errorf("PoB вернул неверный режим узла %d", allocation.NodeID)
		}
		seen[allocation.NodeID] = true
	}
	for _, id := range slices.Concat(change.Added, change.Removed) {
		if !seen[id] {
			return fmt.Errorf("PoB не вернул режим затронутого узла %d", id)
		}
	}
	return nil
}

func validAllocationChange(before map[string]TreeAllocationMode, previous, nodes []int, change PassiveAllocationChange) bool {
	expected := before[strconv.Itoa(change.NodeID)]
	if expected == "" && slices.Contains(previous, change.NodeID) {
		expected = AllocationCommon
	}
	if expected != change.Before || change.Before == change.After {
		return false
	}
	if !slices.Contains(nodes, change.NodeID) {
		return change.After == ""
	}
	return change.After.valid()
}

func writeWeaponBranch(spec *element, name string, mode TreeAllocationMode, changes []PassiveAllocationChange) error {
	old, err := readWeaponNodes(spec, name)
	if err != nil {
		return err
	}
	nodes := editedWeaponNodes(old, mode, changes)
	if slices.Equal(old, nodes) {
		return nil
	}
	branch, _ := spec.child(name)
	if branch == nil {
		branch = newElement(name)
		spec.parts = append(spec.parts, part{element: branch})
	}
	branch.set("nodes", formatPassiveNodes(nodes))
	return nil
}

func editedWeaponNodes(old []int, mode TreeAllocationMode, changes []PassiveAllocationChange) []int {
	nodes := slices.Clone(old)
	for _, change := range changes {
		nodes = slices.DeleteFunc(nodes, func(id int) bool { return id == change.NodeID })
		if change.After == mode {
			nodes = append(nodes, change.NodeID)
		}
	}
	return nodes
}
