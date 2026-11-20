package build

import (
	"encoding/xml"
	"fmt"
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
	GroupID int        `json:"group_id"`
	Label   string     `json:"label"`
	Enabled bool       `json:"enabled"`
	Delete  bool       `json:"delete"`
	Gems    []GemInput `json:"gems"`
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
	MainGroup int         `json:"main_group"`
	WeaponSet int         `json:"weapon_set"`
}

// Edit задаёт ровно одну правку. Пустой Raw снимает предмет, nil Value удаляет условие.
type Edit struct {
	Level     *LevelEdit     `json:"level,omitempty"`
	Config    *ConfigEdit    `json:"config,omitempty"`
	Gems      *GemsEdit      `json:"gems,omitempty"`
	Equipment *EquipmentEdit `json:"equipment,omitempty"`
	Passives  *PassivesEdit  `json:"passives,omitempty"`
	Selection *SelectionEdit `json:"selection,omitempty"`
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
	if data == nil {
		return nil, fmt.Errorf("нет набора данных для правки")
	}
	operations := 0
	for _, present := range []bool{edit.Level != nil, edit.Config != nil, edit.Gems != nil, edit.Equipment != nil, edit.Passives != nil, edit.Selection != nil} {
		if present {
			operations++
		}
	}
	if operations != 1 {
		return nil, fmt.Errorf("нужна ровно одна правка сборки")
	}
	xml, err := d.Export()
	if err != nil {
		return nil, err
	}
	copy, err := Parse(xml)
	if err != nil {
		return nil, err
	}
	if err := copy.applyEdit(edit, data); err != nil {
		return nil, err
	}
	info, _ := copy.root.child("Build")
	for _, name := range []string{"PlayerStat", "MinionStat", "FullDPSSkill", "Buffs"} {
		for _, cached := range info.children(name) {
			info.remove(cached)
		}
	}
	model, err := copy.Model()
	if err != nil {
		return nil, err
	}
	if edit.Selection != nil && edit.Selection.MainGroup > len(model.Groups) {
		return nil, fmt.Errorf("основная группа отсутствует")
	}
	if _, err := copy.Export(); err != nil {
		return nil, err
	}
	return copy, nil
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
		return d.editEquipment(*edit.Equipment)
	case edit.Passives != nil:
		return d.editPassives(*edit.Passives, data)
	case edit.Selection != nil:
		return d.editSelection(*edit.Selection)
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
	id := edit.GroupID
	if id == 0 {
		id = len(groups) + 1
	}
	if id < 1 || id > len(groups)+1 || id > 100 || len(edit.Label) > 512 || !utf8.ValidString(edit.Label) || len(edit.Gems) > 12 {
		return fmt.Errorf("неверная группа камней")
	}
	if edit.Delete {
		if id > len(groups) || len(edit.Gems) > 0 {
			return fmt.Errorf("нельзя удалить эту группу")
		}
		set.remove(groups[id-1])
		info, _ := d.root.child("Build")
		main, err := integer(info.attr("mainSocketGroup", "1"), "mainSocketGroup", 1, 100)
		if err != nil {
			return err
		}
		if main > id {
			main--
		} else if main == id {
			main = 1
		}
		info.set("mainSocketGroup", strconv.Itoa(main))
		return nil
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
	for _, input := range edit.Gems {
		gem, ok := data.Gem(input.ID)
		if !ok || input.Level < 1 || input.Level > 40 || input.Quality < 0 || input.Quality > 100 {
			return fmt.Errorf("неверный камень %s", input.ID)
		}
		node := newElement("Gem")
		for key, value := range map[string]string{"nameSpec": gem.Name, "skillId": gem.SkillID, "gemId": gem.GameID, "variantId": gem.VariantID,
			"level": strconv.Itoa(input.Level), "quality": strconv.Itoa(input.Quality), "enabled": strconv.FormatBool(input.Enabled), "statSetIndex": "1"} {
			node.set(key, value)
		}
		group.parts = append(group.parts, part{element: node})
	}
	return nil
}

func (d *Document) editEquipment(edit EquipmentEdit) error {
	if edit.Slot == "" || len(edit.Slot) > 128 || len(edit.Raw) > 65536 || !utf8.ValidString(edit.Raw) {
		return fmt.Errorf("неверный слот или текст предмета")
	}
	set, err := d.active("Items", "ItemSet", "activeItemSet")
	if err != nil {
		return err
	}
	var slot *element
	for _, node := range set.children("Slot") {
		if node.attr("name", "") == edit.Slot {
			slot = node
			break
		}
	}
	if slot == nil {
		return fmt.Errorf("слот %s отсутствует в выбранном наборе", edit.Slot)
	}
	if strings.TrimSpace(edit.Raw) == "" {
		slot.set("itemId", "0")
		return nil
	}
	items, _ := d.root.child("Items")
	maximum := 0
	for _, node := range items.children("Item") {
		id, err := integer(node.attr("id", ""), "Item/id", 1, 1000000)
		if err != nil {
			return err
		}
		maximum = max(maximum, id)
	}
	if maximum >= 1000000 || len(items.children("Item")) >= 512 {
		return fmt.Errorf("слишком много предметов")
	}
	node := newElement("Item")
	node.set("id", strconv.Itoa(maximum+1))
	node.parts = []part{{text: strings.TrimSpace(edit.Raw)}}
	items.parts = append(items.parts, part{element: node})
	slot.set("itemId", strconv.Itoa(maximum+1))
	return nil
}

func (d *Document) editPassives(edit PassivesEdit, data *gamedata.Dataset) error {
	model, err := d.Model()
	if err != nil {
		return err
	}
	class, ok := data.Class(model.ClassID)
	if !ok {
		return fmt.Errorf("неизвестный класс дерева")
	}
	seen := map[int]bool{class.StartNode: true}
	for _, id := range edit.Nodes {
		if id < 1 || id > 1000000 {
			return fmt.Errorf("неверный идентификатор пассива")
		}
		if seen[id] && id != class.StartNode {
			return fmt.Errorf("повтор пассива")
		}
		seen[id] = true
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

func (d *Document) editSelection(edit SelectionEdit) error {
	if edit.Sets == nil && edit.MainGroup == 0 && edit.WeaponSet == 0 {
		return fmt.Errorf("не указан новый выбор")
	}
	if edit.Sets != nil {
		for _, selection := range []struct {
			name, attribute string
			id              int
		}{
			{"Tree", "activeSpec", edit.Sets.Tree}, {"Items", "activeItemSet", edit.Sets.Items},
			{"Skills", "activeSkillSet", edit.Sets.Skills}, {"Config", "activeConfigSet", edit.Sets.Config},
		} {
			if selection.id == 0 {
				continue
			}
			if selection.id < 1 {
				return fmt.Errorf("неверный идентификатор набора")
			}
			section, _ := d.root.child(selection.name)
			if section == nil {
				return fmt.Errorf("нет секции %s", selection.name)
			}
			section.set(selection.attribute, strconv.Itoa(selection.id))
		}
	}
	if edit.MainGroup != 0 {
		if edit.MainGroup < 1 || edit.MainGroup > 100 {
			return fmt.Errorf("неверная основная группа")
		}
		info, _ := d.root.child("Build")
		info.set("mainSocketGroup", strconv.Itoa(edit.MainGroup))
	}
	if edit.WeaponSet != 0 {
		if edit.WeaponSet != 1 && edit.WeaponSet != 2 {
			return fmt.Errorf("вариант оружия: 1 или 2")
		}
		set, err := d.active("Items", "ItemSet", "activeItemSet")
		if err != nil {
			return err
		}
		set.set("useSecondWeaponSet", strconv.FormatBool(edit.WeaponSet == 2))
	}
	return nil
}
