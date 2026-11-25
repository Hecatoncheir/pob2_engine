package build

import (
	"fmt"
	"strconv"
)

type SkillChoiceEdit struct {
	Control  string `json:"control"`
	OptionID int    `json:"option_id"`
}

type SkillChoiceChange struct {
	Group       int    `json:"group"`
	ActiveSkill int    `json:"active_skill"`
	Gem         int    `json:"gem"`
	Effect      string `json:"effect"`
}

func (e SkillChoiceEdit) validate() error {
	if e.OptionID < 1 || e.OptionID > 100 {
		return fmt.Errorf("неверный номер варианта умения")
	}
	switch e.Control {
	case "groups", "skills", "parts", "stat_sets":
		return nil
	default:
		return fmt.Errorf("неизвестный режим выбора умения")
	}
}

// WithResolvedSkillChoice сохраняет подтверждённый Lua выбор в исходном XML.
func (d *Document) WithResolvedSkillChoice(edit SkillChoiceEdit, change SkillChoiceChange) (*Document, error) {
	updated, err := d.copyForEdit()
	if err != nil {
		return nil, err
	}
	if err := updated.applySkillChoice(edit, change); err != nil {
		return nil, err
	}
	updated.clearCachedResults()
	return updated, nil
}

func (d *Document) applySkillChoice(edit SkillChoiceEdit, change SkillChoiceChange) error {
	if edit.Control == "groups" {
		if change.Group != edit.OptionID {
			return fmt.Errorf("PoB не подтвердил выбранную группу")
		}
		info, err := d.root.child("Build")
		if err != nil || info == nil {
			return fmt.Errorf("в XML нет раздела Build")
		}
		setMainGroup(info, change.Group)
		return nil
	}
	set, err := d.active("Skills", "SkillSet", "activeSkillSet")
	if err != nil {
		return err
	}
	group, err := indexedSkillElement(set, "Skill", change.Group)
	if err != nil {
		return err
	}
	if edit.Control == "skills" {
		group.set("mainActiveSkill", strconv.Itoa(change.ActiveSkill))
		return nil
	}
	gem, err := indexedSkillElement(group, "Gem", change.Gem)
	if err != nil {
		return err
	}
	return applyGemChoice(gem, edit, change)
}

func indexedSkillElement(parent *element, name string, index int) (*element, error) {
	nodes := parent.children(name)
	if index < 1 || index > len(nodes) {
		return nil, fmt.Errorf("элемент %s с номером %d отсутствует в XML", name, index)
	}
	return nodes[index-1], nil
}

func applyGemChoice(gem *element, edit SkillChoiceEdit, change SkillChoiceChange) error {
	value := strconv.Itoa(edit.OptionID)
	if edit.Control == "parts" {
		gem.set("skillPart", value)
		return nil
	}
	if change.Effect == "" {
		return fmt.Errorf("PoB не вернул эффект выбранного варианта")
	}
	for _, child := range gem.children("StatSetIndex") {
		if child.attr("grantedEffect", "") == change.Effect {
			child.set("index", value)
			return nil
		}
	}
	child := newElement("StatSetIndex")
	child.set("grantedEffect", change.Effect)
	child.set("index", value)
	gem.parts = append(gem.parts, part{element: child})
	return nil
}
