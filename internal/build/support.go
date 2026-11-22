package build

import (
	"fmt"
	"strconv"

	"github.com/Hecatoncheir/pob2_engine/internal/gamedata"
)

// WithSupport добавляет камень в копию выбранной группы, сохраняя остальные входы.
// Совместимость и повтор семейства проверяются расчётчиком полного документа.
func (d *Document) WithSupport(gem gamedata.Gem) (*Document, int, error) {
	return d.WithSupportParameters(gem, 1, 0)
}

func (d *Document) WithSupportParameters(gem gamedata.Gem, level, quality int) (*Document, int, error) {
	if !gem.Support {
		return nil, 0, fmt.Errorf("нужен камень поддержки")
	}
	xml, err := d.Export()
	if err != nil {
		return nil, 0, err
	}
	copy, err := Parse(xml)
	if err != nil {
		return nil, 0, err
	}
	model, err := copy.Model()
	if err != nil {
		return nil, 0, err
	}
	skills, err := copy.root.child("Skills")
	if err != nil {
		return nil, 0, err
	}
	if skills == nil {
		return nil, 0, fmt.Errorf("нет Skills для предпросмотра")
	}
	set, _, err := selectedByID(skills, "SkillSet", "activeSkillSet")
	if err != nil {
		return nil, 0, err
	}
	groups := set.children("Skill")
	if model.MainGroup > len(groups) {
		return nil, 0, fmt.Errorf("нет выбранной группы")
	}
	group := groups[model.MainGroup-1]
	node := supportElement(gem)
	node.set("level", strconv.Itoa(level))
	node.set("quality", strconv.Itoa(quality))
	group.parts = append(group.parts, part{element: node})
	copy.clearCachedResults()
	if _, err := copy.Export(); err != nil {
		return nil, 0, err
	}
	return copy, model.MainGroup, nil
}

func supportElement(gem gamedata.Gem) *element {
	node := newElement("Gem")
	for name, value := range map[string]string{"gemId": gem.GameID, "variantId": gem.VariantID, "skillId": gem.SkillID, "nameSpec": gem.Name,
		"level": "1", "quality": "0", "enabled": "true"} {
		node.set(name, value)
	}
	return node
}
