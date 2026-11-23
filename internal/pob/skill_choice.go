package pob

import (
	"context"
	"fmt"

	"github.com/Hecatoncheir/pob2_engine/internal/build"
)

func (c *Calculator) ResolveSkillChoice(ctx context.Context, doc *build.Document, edit build.SkillChoiceEdit) (build.SkillChoiceChange, error) {
	return c.skillChoice(ctx, doc, edit, false)
}

func (c *Calculator) ValidateSkillChoice(ctx context.Context, doc *build.Document, edit build.SkillChoiceEdit) error {
	_, err := c.skillChoice(ctx, doc, edit, true)
	return err
}

func (c *Calculator) skillChoice(ctx context.Context, doc *build.Document, edit build.SkillChoiceEdit, validate bool) (build.SkillChoiceChange, error) {
	xml, err := doc.ExportWithoutProgression()
	if err != nil {
		return build.SkillChoiceChange{}, err
	}
	answer, err := c.run(ctx, &request{
		Operation: "skill_choice", XML: pobXML.Replace(string(xml)), SkillChoice: &edit, Validate: validate,
	})
	if err != nil {
		return build.SkillChoiceChange{}, err
	}
	if answer.Error != "" {
		return build.SkillChoiceChange{}, fmt.Errorf("PoB Lua: %s", answer.Error)
	}
	if answer.SkillChoice == nil {
		return build.SkillChoiceChange{}, fmt.Errorf("PoB не подтвердил выбор умения")
	}
	return *answer.SkillChoice, nil
}
