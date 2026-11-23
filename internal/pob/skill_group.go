package pob

import (
	"context"
	"fmt"

	"github.com/Hecatoncheir/pob2_engine/internal/build"
)

func (c *Calculator) ResolveSkillGroup(ctx context.Context, doc *build.Document, edit build.SkillGroupEdit) (*build.Document, error) {
	xml, err := doc.ExportWithoutProgression()
	if err != nil {
		return nil, err
	}
	answer, err := c.run(ctx, &request{Operation: "skill_group", XML: pobXML.Replace(string(xml)), SkillGroup: &edit})
	if err != nil {
		return nil, err
	}
	if answer.Error != "" {
		return nil, fmt.Errorf("PoB Lua: %s", answer.Error)
	}
	return build.Parse([]byte(answer.SkillGroupXML))
}
