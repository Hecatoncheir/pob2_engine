package build

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

type Issue struct {
	Code    string `json:"code"`
	Path    string `json:"path"`
	Message string `json:"message"`
}

type Gem struct {
	VariantID string `json:"variant_id"`
	ID        string `json:"id"`
	Name      string `json:"name"`
	SkillID   string `json:"skill_id"`
	Level     int    `json:"level"`
	Quality   int    `json:"quality"`
	Enabled   bool   `json:"enabled"`
	StatSet   int    `json:"stat_set"`
	Corrupted bool   `json:"corrupted"`
}

type Group struct {
	ID        int    `json:"id"`
	Label     string `json:"label"`
	Enabled   bool   `json:"enabled"`
	MainSkill int    `json:"main_skill"`
	Gems      []Gem  `json:"gems"`
}

type Item struct {
	ID  int    `json:"id"`
	Raw string `json:"raw"`
}

type ConfigValue struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type ActiveSets struct {
	Skills int `json:"skills"`
	Items  int `json:"items"`
	Tree   int `json:"tree"`
	Config int `json:"config"`
}

// Model содержит выбранные входы; остальные наборы сохраняются в Document.
type Model struct {
	TargetVersion string                 `json:"target_version"`
	TreeVersion   string                 `json:"tree_version"`
	Class         string                 `json:"class"`
	ClassID       int                    `json:"class_id"`
	Ascendancy    string                 `json:"ascendancy"`
	AscendancyID  int                    `json:"ascendancy_id"`
	Level         int                    `json:"level"`
	MainGroup     int                    `json:"main_group"`
	WeaponSet     int                    `json:"weapon_set"`
	ActiveSets    ActiveSets             `json:"active_sets"`
	Groups        []Group                `json:"groups"`
	Items         []Item                 `json:"items"`
	Slots         map[string]int         `json:"slots"`
	Nodes         []int                  `json:"nodes"`
	Config        map[string]ConfigValue `json:"config"`
	Issues        []Issue                `json:"issues"`
}

func integer(value, path string, minimum, maximum int) (int, error) {
	number, err := strconv.Atoi(value)
	if err != nil || number < minimum || number > maximum {
		return 0, fmt.Errorf("неверное целое число %s", path)
	}
	return number, nil
}

func boolean(value, path string) (bool, error) {
	if value == "true" {
		return true, nil
	}
	if value == "false" || value == "nil" {
		return false, nil
	}
	return false, fmt.Errorf("неверное булево значение %s", path)
}

func (m *Model) issue(path, message string) {
	m.Issues = append(m.Issues, Issue{Code: "unsupported_xml", Path: path, Message: message})
}

func (d *Document) Model() (Model, error) {
	m := Model{WeaponSet: 1, Slots: make(map[string]int), Config: make(map[string]ConfigValue),
		Groups: []Group{}, Items: []Item{}, Nodes: []int{}, Issues: []Issue{}}
	info, err := d.root.child("Build")
	if err != nil {
		return m, err
	}
	if info == nil {
		return m, fmt.Errorf("в XML нет Build")
	}
	m.Class, m.Ascendancy, m.TargetVersion = info.attr("className", ""), info.attr("ascendClassName", "None"), info.attr("targetVersion", "")
	if m.Class == "" {
		return m, fmt.Errorf("в XML не указан класс")
	}
	m.Level, err = integer(info.attr("level", ""), "Build/level", 1, 100)
	if err != nil {
		return m, err
	}
	m.MainGroup, err = integer(info.attr("mainSocketGroup", "1"), "Build/mainSocketGroup", 1, 100)
	if err != nil {
		return m, err
	}
	for _, section := range []struct {
		name  string
		parse func(*element) error
	}{
		{"Tree", m.readTree}, {"Items", m.readItems}, {"Skills", m.readSkills}, {"Config", m.readConfig},
	} {
		node, err := d.root.child(section.name)
		if err != nil {
			return m, err
		}
		if node == nil {
			m.issue(section.name, "отсутствует секция")
			continue
		}
		if err := section.parse(node); err != nil {
			return m, err
		}
	}
	d.findExtensions(&m)
	return m, nil
}

func selectedByID(parent *element, child, attribute string) (*element, int, error) {
	id, err := integer(parent.attr(attribute, "1"), attribute, 1, 1000)
	if err != nil {
		return nil, 0, err
	}
	var selected *element
	seen := make(map[int]bool)
	for _, node := range parent.children(child) {
		value, err := integer(node.attr("id", "1"), child+"/id", 1, 1000)
		if err != nil {
			return nil, 0, err
		}
		if seen[value] {
			return nil, 0, fmt.Errorf("повтор id набора %s", child)
		}
		seen[value] = true
		if value == id {
			selected = node
		}
	}
	if selected == nil {
		return nil, 0, fmt.Errorf("не найден выбранный набор %s %d", child, id)
	}
	return selected, id, nil
}

func (m *Model) readTree(tree *element) error {
	id, err := integer(tree.attr("activeSpec", "1"), "Tree/activeSpec", 1, 1000)
	if err != nil {
		return err
	}
	specs := tree.children("Spec")
	if id > len(specs) {
		return fmt.Errorf("не найдена выбранная спецификация дерева")
	}
	node := specs[id-1]
	m.ActiveSets.Tree, m.TreeVersion = id, node.attr("treeVersion", "")
	m.ClassID, err = integer(node.attr("classId", ""), "Spec/classId", 1, 100)
	if err != nil {
		return err
	}
	if err := m.checkClassIdentity(node); err != nil {
		return err
	}
	m.AscendancyID, err = integer(node.attr("ascendClassId", "0"), "Spec/ascendClassId", 0, 100)
	if err != nil {
		return err
	}
	m.Nodes, err = readPassiveNodes(node.attr("nodes", ""))
	return err
}

func readPassiveNodes(raw string) ([]int, error) {
	seen := make(map[int]bool)
	var nodes []int
	for _, value := range strings.Split(raw, ",") {
		if strings.TrimSpace(value) == "" {
			continue
		}
		id, err := integer(strings.TrimSpace(value), "Spec/nodes", 1, 1000000)
		if err != nil {
			return nil, err
		}
		if seen[id] {
			return nil, fmt.Errorf("повтор пассива %d", id)
		}
		seen[id] = true
		nodes = append(nodes, id)
	}
	sort.Ints(nodes)
	return nodes, nil
}

func (m *Model) checkClassIdentity(spec *element) error {
	value := spec.attr("classInternalId", "")
	if value == "" {
		return nil
	}
	id, err := integer(value, "Spec/classInternalId", 1, 100)
	if err != nil {
		return err
	}
	if id != m.ClassID {
		m.issue("Tree/Spec/classInternalId", "идентификаторы класса противоречат друг другу")
	}
	return nil
}

func (m *Model) readItems(items *element) error {
	set, id, err := selectedByID(items, "ItemSet", "activeItemSet")
	if err != nil {
		return err
	}
	m.ActiveSets.Items = id
	second, err := boolean(set.attr("useSecondWeaponSet", items.attr("useSecondWeaponSet", "false")), "ItemSet/useSecondWeaponSet")
	if err != nil {
		return err
	}
	if second {
		m.WeaponSet = 2
	}
	seen := make(map[int]bool)
	for _, node := range items.children("Item") {
		id, err := integer(node.attr("id", ""), "Item/id", 1, 1000000)
		if err != nil {
			return err
		}
		if seen[id] {
			return fmt.Errorf("повтор предмета %d", id)
		}
		seen[id] = true
		m.Items = append(m.Items, Item{ID: id, Raw: strings.TrimSpace(node.content())})
	}
	sort.Slice(m.Items, func(i, j int) bool { return m.Items[i].ID < m.Items[j].ID })
	for _, node := range set.children("Slot") {
		name := node.attr("name", "")
		if _, exists := m.Slots[name]; exists || name == "" {
			return fmt.Errorf("пустой или повторный слот")
		}
		id, err := integer(node.attr("itemId", "0"), "Slot/itemId", 0, 1000000)
		if err != nil {
			return err
		}
		if id > 0 && !seen[id] {
			return fmt.Errorf("слот %s ссылается на неизвестный предмет", name)
		}
		m.Slots[name] = id
	}
	return nil
}

func (m *Model) readSkills(skills *element) error {
	set, id, err := selectedByID(skills, "SkillSet", "activeSkillSet")
	if err != nil {
		return err
	}
	m.ActiveSets.Skills = id
	for i, node := range set.children("Skill") {
		group := Group{ID: i + 1, Label: node.attr("label", ""), Gems: []Gem{}}
		group.Enabled, err = boolean(node.attr("enabled", "true"), "Skill/enabled")
		if err != nil {
			return err
		}
		group.MainSkill, err = integer(node.attr("mainActiveSkill", "1"), "Skill/mainActiveSkill", 1, 100)
		if err != nil {
			return err
		}
		for _, child := range node.children("Gem") {
			gem, err := readGem(child)
			if err != nil {
				return err
			}
			group.Gems = append(group.Gems, gem)
		}
		m.Groups = append(m.Groups, group)
	}
	return nil
}

func readGem(node *element) (Gem, error) {
	gem := Gem{ID: node.attr("gemId", ""), VariantID: node.attr("variantId", ""), Name: node.attr("nameSpec", ""), SkillID: node.attr("skillId", "")}
	var err error
	gem.Level, err = integer(node.attr("level", "1"), "Gem/level", 1, 100)
	if err != nil {
		return gem, err
	}
	gem.Quality, err = integer(node.attr("quality", "0"), "Gem/quality", 0, 100)
	if err != nil {
		return gem, err
	}
	gem.Enabled, err = boolean(node.attr("enabled", "true"), "Gem/enabled")
	if err != nil {
		return gem, err
	}
	gem.Corrupted, err = boolean(node.attr("corrupted", "false"), "Gem/corrupted")
	if err != nil {
		return gem, err
	}
	value := node.attr("statSetIndex", "1")
	if value == "nil" {
		value = "1"
	}
	gem.StatSet, err = integer(value, "Gem/statSetIndex", 1, 100)
	return gem, err
}

func configValue(node *element) (ConfigValue, error) {
	result := ConfigValue{}
	for _, kind := range []string{"boolean", "number", "string"} {
		for _, attr := range node.attrs {
			if attr.Name.Local != kind {
				continue
			}
			if result.Type != "" {
				return result, fmt.Errorf("несколько типов значения Config/Input")
			}
			result = ConfigValue{Type: kind, Value: attr.Value}
		}
	}
	if result.Type == "" {
		return result, fmt.Errorf("не указан тип Config/Input")
	}
	if result.Type == "boolean" {
		if _, err := boolean(result.Value, "Config/Input"); err != nil {
			return result, err
		}
	}
	if result.Type == "number" {
		number, err := strconv.ParseFloat(result.Value, 64)
		if err != nil || math.IsNaN(number) || math.IsInf(number, 0) {
			return result, fmt.Errorf("неверное число Config/Input")
		}
		result.Value = strconv.FormatFloat(number, 'g', -1, 64)
	}
	return result, nil
}

func (m *Model) readConfig(config *element) error {
	set, id, err := selectedByID(config, "ConfigSet", "activeConfigSet")
	if err != nil {
		return err
	}
	m.ActiveSets.Config = id
	for _, input := range set.children("Input") {
		name := input.attr("name", "")
		if _, exists := m.Config[name]; name == "" || exists {
			return fmt.Errorf("пустой или повторный параметр Config/Input")
		}
		value, err := configValue(input)
		if err != nil {
			return err
		}
		m.Config[name] = value
	}
	return nil
}
