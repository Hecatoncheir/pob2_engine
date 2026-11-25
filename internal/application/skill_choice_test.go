package application

import (
	"context"
	"strings"
	"testing"

	"github.com/Hecatoncheir/pob2_engine/internal/build"
	"github.com/Hecatoncheir/pob2_engine/internal/calculation"
)

type countingSkillChoiceEditor struct {
	calculation.Calculator
	group       int
	resolutions int
	validations int
}

func (e *countingSkillChoiceEditor) ResolveSkillChoice(context.Context, *build.Document, build.SkillChoiceEdit) (build.SkillChoiceChange, error) {
	e.resolutions++
	return build.SkillChoiceChange{Group: e.group}, nil
}

func (e *countingSkillChoiceEditor) ValidateSkillChoice(context.Context, *build.Document, build.SkillChoiceEdit) error {
	e.validations++
	return nil
}

func TestResolveMainSkillGroupWithOneLuaRequest(t *testing.T) {
	doc, err := build.Parse([]byte(`<PathOfBuilding2>
  <Build className="Witch" level="1" mainSkillIndex="1" mainSocketGroup="1"/>
  <Skills><SkillSet id="1"><Skill/><Skill/></SkillSet></Skills>
</PathOfBuilding2>`))
	if err != nil {
		t.Fatal(err)
	}
	editor := &countingSkillChoiceEditor{group: 2}
	service := &Service{calculator: editor}
	edit := build.Edit{SkillChoice: &build.SkillChoiceEdit{Control: "groups", OptionID: 2}}
	updated, err := service.editSkillChoice(t.Context(), doc, edit)
	if err != nil {
		t.Fatal(err)
	}
	if editor.resolutions != 1 || editor.validations != 0 {
		t.Fatalf("ожидался один запрос Lua, получено %d правок и %d повторных проверок", editor.resolutions, editor.validations)
	}
	xml, err := updated.Export()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(xml), `mainSkillIndex="2"`) || !strings.Contains(string(xml), `mainSocketGroup="2"`) {
		t.Fatal("номер группы не сохранён в обоих атрибутах XML")
	}
	editor.group = 1
	if _, err := service.editSkillChoice(t.Context(), doc, edit); err == nil {
		t.Fatal("принят номер группы, который отличается от выбранного")
	}
}
