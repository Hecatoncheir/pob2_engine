package build

import (
	"encoding/xml"
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/Hecatoncheir/pob2_engine/internal/gamedata"
)

type LevelEdit struct {
	Level int `json:"level"`
}
type ConfigEdit struct {
	Name  string       `json:"name"`
	Value *ConfigValue `json:"value"`
}
type GemInput struct {
	ID      string `json:"id"`
	Level   int    `json:"level"`
	Quality int    `json:"quality"`
	Enabled bool   `json:"enabled"`
}
type GemsEdit struct {
	GroupID int        `json:"group_id,omitempty"`
	Label   string     `json:"label,omitempty"`
	Enabled bool       `json:"enabled,omitempty"`
	Delete  bool       `json:"delete,omitempty"`
	Gems    []GemInput `json:"gems,omitempty"`
}
type EquipmentEdit struct {
	Slot string `json:"slot"`
	Raw  string `json:"raw"`
}
type PassivesEdit struct {
	Nodes []int `json:"nodes"`
}
type SelectionEdit struct {
	Sets      *ActiveSets `json:"sets,omitempty"`
	MainGroup int         `json:"main_group,omitempty"`
	WeaponSet int         `json:"weapon_set,omitempty"`
}

// Edit задаёт ровно одну правку. Пустой Raw снимает предмет, nil Value удаляет условие.
type Edit struct {
	Level      *LevelEdit      `json:"level,omitempty"`
	Config     *ConfigEdit     `json:"config,omitempty"`
	Gems       *GemsEdit       `json:"gems,omitempty"`
	Equipment  *EquipmentEdit  `json:"equipment,omitempty"`
	Passives   *PassivesEdit   `json:"passives,omitempty"`
	Selection  *SelectionEdit  `json:"selection,omitempty"`
	AddSupport *AddSupportEdit `json:"add_support,omitempty"`
	TreeNode   *TreeNodeEdit   `json:"tree_node,omitempty"`
}

func newElement(name string) *element { return &element{name: xml.Name{Local: name}} }

func (n *element) remove(target *element) {
	for i, p := range n.parts {
		if p.element == target {
			n.parts = append(n.parts[:i], n.parts[i+1:]...)
			return
		}
	}
}

func (d *Document) Edited(edit Edit, data *gamedata.Dataset) (*Document, error) {
	if err := validateEdit(edit, data); err != nil {
		return nil, err
	}
	copy, err := d.copyForEdit()
	if err != nil {
		return nil, err
	}
	if err := copy.applyEdit(edit, data); err != nil {
		return nil, err
	}
	copy.clearCachedResults()
	model, err := copy.Model()
	if err != nil {
		return nil, err
	}
	if selectsSkills(edit.Selection) && model.MainGroup > len(model.Groups) {
		return nil, fmt.Errorf("основная группа отсутствует")
	}
	if _, err := copy.Export(); err != nil {
		return nil, err
	}
	return copy, nil
}

func (d *Document) copyForEdit() (*Document, error) {
	xml, err := d.Export()
	if err != nil {
		return nil, err
	}
	return Parse(xml)
}

func validateEdit(edit Edit, data *gamedata.Dataset) error {
	if data == nil {
		return fmt.Errorf("нет набора данных для правки")
	}
	return ValidateEdit(edit)
}

// ValidateEdit проверяет форму команды независимо от изменяемого каталога PoB.
func ValidateEdit(edit Edit) error {
	operations := 0
	for _, present := range []bool{edit.Level != nil, edit.Config != nil, edit.Gems != nil, edit.Equipment != nil, edit.Passives != nil, edit.Selection != nil, edit.AddSupport != nil, edit.TreeNode != nil} {
		if present {
			operations++
		}
	}
	if operations != 1 {
		return fmt.Errorf("нужна ровно одна правка сборки")
	}
	if edit.AddSupport != nil {
		return edit.AddSupport.ValidateID()
	}
	if edit.TreeNode != nil {
		return edit.TreeNode.validate()
	}
	return nil
}

func (d *Document) clearCachedResults() {
	info, _ := d.root.child("Build")
	for _, name := range []string{"PlayerStat", "MinionStat", "FullDPSSkill", "Buffs"} {
		for _, cached := range info.children(name) {
			info.remove(cached)
		}
	}
}

func selectsSkills(selection *SelectionEdit) bool {
	if selection == nil {
		return false
	}
	if selection.MainGroup != 0 {
		return true
	}
	return selection.Sets != nil && selection.Sets.Skills != 0
}

func (d *Document) applyEdit(edit Edit, data *gamedata.Dataset) error {
	switch {
	case edit.Level != nil:
		if edit.Level.Level < 1 || edit.Level.Level > 100 {
			return fmt.Errorf("уровень: 1–100")
		}
		info, _ := d.root.child("Build")
		info.set("level", strconv.Itoa(edit.Level.Level))
		info.set("characterLevelAutoMode", "false")
	case edit.Config != nil:
		return d.editConfig(*edit.Config)
	case edit.Gems != nil:
		return d.editGems(*edit.Gems, data)
	case edit.Equipment != nil:
		return d.editEquipment(*edit.Equipment, data.IsRuntime())
	case edit.Passives != nil:
		return d.editPassives(*edit.Passives, data)
	case edit.Selection != nil:
		return d.editSelection(*edit.Selection)
	case edit.AddSupport != nil:
		return d.editSupport(*edit.AddSupport, data)
	case edit.TreeNode != nil:
		return fmt.Errorf("для правки узла требуется путь из оригинального PoB")
	}
	return nil
}

func (d *Document) active(name, child, attribute string) (*element, error) {
	parent, err := d.root.child(name)
	if err != nil {
		return nil, err
	}
	if parent == nil {
		return nil, fmt.Errorf("нет секции %s", name)
	}
	selected, _, err := selectedByID(parent, child, attribute)
	return selected, err
}

func (d *Document) editConfig(edit ConfigEdit) error {
	if len(edit.Name) == 0 || len(edit.Name) > 256 || !utf8.ValidString(edit.Name) {
		return fmt.Errorf("неверное имя условия")
	}
	set, err := d.active("Config", "ConfigSet", "activeConfigSet")
	if err != nil {
		return err
	}
	var input *element
	for _, node := range set.children("Input") {
		if node.attr("name", "") == edit.Name {
			input = node
			break
		}
	}
	if edit.Value == nil {
		if input != nil {
			set.remove(input)
		}
		return nil
	}
	if !containsWord("boolean number string", edit.Value.Type) {
		return fmt.Errorf("тип условия: boolean, number или string")
	}
	if !utf8.ValidString(edit.Value.Value) || len(edit.Value.Value) > 65536 {
		return fmt.Errorf("значение условия: до 65536 байт UTF-8")
	}
	replacement := newElement("Input")
	replacement.set("name", edit.Name)
	replacement.set(edit.Value.Type, edit.Value.Value)
	if _, err := configValue(replacement); err != nil {
		return err
	}
	if input == nil {
		set.parts = append(set.parts, part{element: replacement})
	} else {
		replacement.parts = input.parts
		// Неизвестные атрибуты исходного условия сохраняются при смене его значения.
		for _, attr := range input.attrs {
			if !containsWord("name boolean number string", attr.Name.Local) {
				replacement.attrs = append(replacement.attrs, attr)
			}
		}
		for i, p := range set.parts {
			if p.element == input {
				set.parts[i].element = replacement
				break
			}
		}
	}
	return nil
}

func (d *Document) editGems(edit GemsEdit, data *gamedata.Dataset) error {
	set, err := d.active("Skills", "SkillSet", "activeSkillSet")
	if err != nil {
		return err
	}
	groups := set.children("Skill")
	id, err := validateGemGroup(edit, groups)
	if err != nil {
		return err
	}
	if edit.Delete {
		return d.removeGemGroup(set, groups, id)
	}
	nodes, err := gemElements(edit.Gems, data)
	if err != nil {
		return err
	}
	group := newElement("Skill")
	if id <= len(groups) {
		group = groups[id-1]
	} else {
		set.parts = append(set.parts, part{element: group})
	}
	group.set("label", edit.Label)
	group.set("enabled", strconv.FormatBool(edit.Enabled))
	group.set("mainActiveSkill", "1")
	group.set("includeInFullDPS", "true")
	for _, gem := range group.children("Gem") {
		group.remove(gem)
	}
	for _, node := range nodes {
		group.parts = append(group.parts, part{element: node})
	}
	return nil
}

func validateGemGroup(edit GemsEdit, groups []*element) (int, error) {
	id := edit.GroupID
	if id == 0 {
		id = len(groups) + 1
	}
	if id < 1 || id > len(groups)+1 || id > 100 || len(edit.Label) > 512 || !utf8.ValidString(edit.Label) || len(edit.Gems) > 12 {
		return 0, fmt.Errorf("неверная группа камней")
	}
	if edit.Delete && (id > len(groups) || len(edit.Gems) > 0) {
		return 0, fmt.Errorf("нельзя удалить эту группу")
	}
	return id, nil
}

func (d *Document) removeGemGroup(set *element, groups []*element, id int) error {
	set.remove(groups[id-1])
	info, _ := d.root.child("Build")
	main, err := mainGroup(info)
	if err != nil {
		return err
	}
	if main > id {
		main--
	} else if main == id {
		main = 1
	}
	setMainGroup(info, main)
	return nil
}

func setMainGroup(info *element, id int) {
	value := strconv.Itoa(id)
	info.set("mainSocketGroup", value)
	if info.attr("mainSkillIndex", "") != "" {
		info.set("mainSkillIndex", value)
	}
}

func gemElements(inputs []GemInput, data *gamedata.Dataset) ([]*element, error) {
	nodes := make([]*element, 0, len(inputs))
	for _, input := range inputs {
		gem, ok := data.Gem(input.ID)
		if !ok || !validGemInput(input, gem, data.IsRuntime()) {
			return nil, fmt.Errorf("неверный камень %s", input.ID)
		}
		nodes = append(nodes, gemElement(input, gem))
	}
	return nodes, nil
}

func gemElement(input GemInput, gem gamedata.Gem) *element {
	node := newElement("Gem")
	for key, value := range map[string]string{"nameSpec": gem.Name, "skillId": gem.SkillID, "gemId": gem.GameID, "variantId": gem.VariantID,
		"level": strconv.Itoa(input.Level), "quality": strconv.Itoa(input.Quality), "enabled": strconv.FormatBool(input.Enabled), "statSetIndex": "1"} {
		node.set(key, value)
	}
	return node
}

func validGemInput(input GemInput, gem gamedata.Gem, runtime bool) bool {
	if input.Level < 1 || input.Level > 40 || input.Quality < 0 || input.Quality > 100 {
		return false
	}
	if !runtime {
		return true
	}
	return slices.ContainsFunc(gem.Levels, func(level gamedata.GemLevel) bool { return level.Level == input.Level }) &&
		slices.Contains(gem.Qualities, input.Quality)
}

func (d *Document) editPassives(edit PassivesEdit, data *gamedata.Dataset) error {
	model, err := d.Model()
	if err != nil {
		return err
	}
	if data.IsRuntime() && model.TreeVersion != data.TreeVersion() {
		return fmt.Errorf("версия дерева сборки %s не соответствует каталогу %s", model.TreeVersion, data.TreeVersion())
	}
	class, ok := data.Class(model.ClassID)
	if !ok {
		return fmt.Errorf("неизвестный класс дерева")
	}
	seen, err := selectedPassiveNodes(class.StartNode, edit.Nodes, data)
	if err != nil {
		return err
	}
	if len(seen) > 500 {
		return fmt.Errorf("слишком много пассивов")
	}
	nodes := make([]int, 0, len(seen))
	for id := range seen {
		nodes = append(nodes, id)
	}
	sort.Ints(nodes)
	values := make([]string, 0, len(nodes))
	for _, id := range nodes {
		values = append(values, strconv.Itoa(id))
	}
	tree, _ := d.root.child("Tree")
	tree.children("Spec")[model.ActiveSets.Tree-1].set("nodes", strings.Join(values, ","))
	return nil
}

func selectedPassiveNodes(start int, nodes []int, data *gamedata.Dataset) (map[int]bool, error) {
	seen := map[int]bool{start: true}
	for _, id := range nodes {
		if err := validatePassiveID(id, data); err != nil {
			return nil, err
		}
		if seen[id] && id != start {
			return nil, fmt.Errorf("повтор пассива")
		}
		seen[id] = true
	}
	return seen, nil
}

func validatePassiveID(id int, data *gamedata.Dataset) error {
	if id < 1 || id > 1000000 {
		return fmt.Errorf("неверный идентификатор пассива")
	}
	if data.IsRuntime() {
		if _, ok := data.Node(id); !ok {
			return fmt.Errorf("пассив %d отсутствует в активном PoB", id)
		}
	}
	return nil
}

func (d *Document) editSelection(edit SelectionEdit) error {
	if edit.Sets == nil && edit.MainGroup == 0 && edit.WeaponSet == 0 {
		return fmt.Errorf("не указан новый выбор")
	}
	if err := d.selectSets(edit.Sets); err != nil {
		return err
	}
	if edit.MainGroup != 0 {
		if edit.MainGroup < 1 || edit.MainGroup > 100 {
			return fmt.Errorf("неверная основная группа")
		}
		info, _ := d.root.child("Build")
		setMainGroup(info, edit.MainGroup)
	}
	return d.selectWeaponSet(edit.WeaponSet)
}

func (d *Document) selectSets(sets *ActiveSets) error {
	if sets == nil {
		return nil
	}
	selections := []struct {
		section, attribute string
		id                 int
	}{
		{"Tree", "activeSpec", sets.Tree},
		{"Items", "activeItemSet", sets.Items},
		{"Skills", "activeSkillSet", sets.Skills},
		{"Config", "activeConfigSet", sets.Config},
	}
	for _, selection := range selections {
		if err := d.selectSet(selection.section, selection.attribute, selection.id); err != nil {
			return err
		}
	}
	return nil
}

func (d *Document) selectSet(name, attribute string, id int) error {
	if id == 0 {
		return nil
	}
	if id < 1 {
		return fmt.Errorf("неверный идентификатор набора")
	}
	section, err := d.root.child(name)
	if err != nil {
		return err
	}
	if section == nil {
		return fmt.Errorf("нет секции %s", name)
	}
	section.set(attribute, strconv.Itoa(id))
	return nil
}

func (d *Document) selectWeaponSet(id int) error {
	if id == 0 {
		return nil
	}
	if id != 1 && id != 2 {
		return fmt.Errorf("вариант оружия: 1 или 2")
	}
	set, err := d.active("Items", "ItemSet", "activeItemSet")
	if err != nil {
		return err
	}
	set.set("useSecondWeaponSet", strconv.FormatBool(id == 2))
	return nil
}
