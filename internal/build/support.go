package build

import (
	"fmt"
	"slices"
	"strconv"
	"unicode/utf8"

	"github.com/Hecatoncheir/pob2_engine/internal/gamedata"
)

type AddSupportEdit struct {
	SupportID string `json:"support_id"`
	Level     int    `json:"level,omitempty"`
	Quality   int    `json:"quality,omitempty"`
}

func (edit AddSupportEdit) ValidateID() error {
	if !utf8.ValidString(edit.SupportID) || len(edit.SupportID) > 256 {
		return fmt.Errorf("support_id: нужен UTF-8 до 256 байт")
	}
	return nil
}

func (edit AddSupportEdit) Resolve(data *gamedata.Dataset) (gamedata.Gem, int, error) {
	gem, ok := data.Gem(edit.SupportID)
	if !ok || !gem.Support {
		return gamedata.Gem{}, 0, fmt.Errorf("неизвестная или неподдерживаемая поддержка")
	}
	level := edit.Level
	if level == 0 {
		level = 1
	}
	if level < 1 || level > 40 || edit.Quality < 0 || edit.Quality > 100 {
		return gamedata.Gem{}, 0, fmt.Errorf("поддержка: уровень 1–40, качество 0–100")
	}
	if !data.IsRuntime() && (level != 1 || edit.Quality != 0) {
		return gamedata.Gem{}, 0, fmt.Errorf("режим Go поддерживает только уровень 1 и качество 0")
	}
	if !slices.ContainsFunc(gem.Levels, func(candidate gamedata.GemLevel) bool { return candidate.Level == level }) {
		return gamedata.Gem{}, 0, fmt.Errorf("этого уровня поддержки нет в активном PoB")
	}
	return gem, level, nil
}

func (d *Document) editSupport(edit AddSupportEdit, data *gamedata.Dataset) error {
	gem, level, err := edit.Resolve(data)
	if err != nil {
		return err
	}
	_, err = d.appendSupport(gem, level, edit.Quality)
	return err
}

// WithSupport добавляет камень в копию выбранной группы, сохраняя остальные входы.
// Совместимость и повтор семейства проверяются расчётчиком полного документа.
func (d *Document) WithSupport(gem gamedata.Gem) (*Document, int, error) {
	return d.WithSupportParameters(gem, 1, 0)
}

func (d *Document) WithSupportParameters(gem gamedata.Gem, level, quality int) (*Document, int, error) {
	xml, err := d.Export()
	if err != nil {
		return nil, 0, err
	}
	copy, err := Parse(xml)
	if err != nil {
		return nil, 0, err
	}
	group, err := copy.appendSupport(gem, level, quality)
	if err != nil {
		return nil, 0, err
	}
	copy.clearCachedResults()
	if _, err := copy.Export(); err != nil {
		return nil, 0, err
	}
	return copy, group, nil
}

func (d *Document) appendSupport(gem gamedata.Gem, level, quality int) (int, error) {
	if !gem.Support {
		return 0, fmt.Errorf("нужен камень поддержки")
	}
	model, err := d.Model()
	if err != nil {
		return 0, err
	}
	set, err := d.active("Skills", "SkillSet", "activeSkillSet")
	if err != nil {
		return 0, err
	}
	groups := set.children("Skill")
	if model.MainGroup < 1 || model.MainGroup > len(groups) {
		return 0, fmt.Errorf("нет выбранной группы")
	}
	group := groups[model.MainGroup-1]
	node := supportElement(gem)
	node.set("level", strconv.Itoa(level))
	node.set("quality", strconv.Itoa(quality))
	group.parts = append(group.parts, part{element: node})
	return model.MainGroup, nil
}

func supportElement(gem gamedata.Gem) *element {
	node := newElement("Gem")
	for name, value := range map[string]string{"gemId": gem.GameID, "variantId": gem.VariantID, "skillId": gem.SkillID, "nameSpec": gem.Name,
		"level": "1", "quality": "0", "enabled": "true"} {
		node.set(name, value)
	}
	return node
}
