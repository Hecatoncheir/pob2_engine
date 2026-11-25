package application

import (
	"context"
	"fmt"

	"github.com/Hecatoncheir/pob2_engine/internal/build"
)

type skillChoiceEditor interface {
	ResolveSkillChoice(context.Context, *build.Document, build.SkillChoiceEdit) (build.SkillChoiceChange, error)
	ValidateSkillChoice(context.Context, *build.Document, build.SkillChoiceEdit) error
}

func (s *Service) SupportsSkillChoice() bool {
	_, ok := s.calculator.(skillChoiceEditor)
	return ok
}

func (s *Service) editSkillChoice(ctx context.Context, doc *build.Document, edit build.Edit) (*build.Document, error) {
	if err := build.ValidateEdit(edit); err != nil {
		return nil, err
	}
	editor, ok := s.calculator.(skillChoiceEditor)
	if !ok {
		return nil, fmt.Errorf("выбор умения доступен с исходным Lua PoB")
	}
	choice := *edit.SkillChoice
	change, err := editor.ResolveSkillChoice(ctx, doc, choice)
	if err != nil {
		return nil, err
	}
	updated, err := doc.WithResolvedSkillChoice(choice, change)
	if err != nil {
		return nil, err
	}
	if choice.Control == "groups" {
		// PoB уже подтвердил выбор после кадра, а XML хранит номер группы напрямую.
		return updated, nil
	}
	if err := editor.ValidateSkillChoice(ctx, updated, choice); err != nil {
		return nil, err
	}
	return updated, nil
}
