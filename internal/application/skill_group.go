package application

import (
	"context"
	"fmt"

	"github.com/Hecatoncheir/pob2_engine/internal/build"
)

type skillGroupEditor interface {
	ResolveSkillGroup(context.Context, *build.Document, build.SkillGroupEdit) (*build.Document, error)
}

func (s *Service) SupportsSkillGroup() bool {
	_, ok := s.calculator.(skillGroupEditor)
	return ok
}

func (s *Service) editSkillGroup(ctx context.Context, doc *build.Document, edit build.Edit) (*build.Document, error) {
	if err := build.ValidateEdit(edit); err != nil {
		return nil, err
	}
	editor, ok := s.calculator.(skillGroupEditor)
	if !ok {
		return nil, fmt.Errorf("штатная правка группы доступна с исходным Lua PoB")
	}
	result, err := editor.ResolveSkillGroup(ctx, doc, *edit.SkillGroup)
	if err != nil {
		return nil, err
	}
	return doc.WithResolvedSkillGroup(result)
}
