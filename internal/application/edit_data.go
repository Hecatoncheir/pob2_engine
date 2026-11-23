package application

import (
	"context"

	"github.com/Hecatoncheir/pob2_engine/internal/build"
	"github.com/Hecatoncheir/pob2_engine/internal/gamedata"
)

type catalogDataSource interface {
	CatalogData(context.Context) (*gamedata.Dataset, error)
}

type equipmentValidator interface {
	ValidateEquipment(context.Context, *build.Document, string) error
}

// SupportsEquipmentPreview сообщает, что движок проверяет экипировку через PoB.
func (s *Service) SupportsEquipmentPreview() bool {
	_, ok := s.calculator.(equipmentValidator)
	return ok
}

func (s *Service) editDocument(ctx context.Context, doc *build.Document, edit build.Edit) (*build.Document, error) {
	data, err := s.editData(ctx)
	if err != nil {
		return nil, err
	}
	updated, err := doc.Edited(edit, data)
	if err != nil {
		return nil, err
	}
	if validator, ok := s.calculator.(equipmentValidator); ok && edit.Equipment != nil {
		if err := validator.ValidateEquipment(ctx, updated, edit.Equipment.Slot); err != nil {
			return nil, err
		}
	}
	return updated, nil
}

func (s *Service) editData(ctx context.Context) (*gamedata.Dataset, error) {
	if err := s.loadData(); err != nil {
		return nil, err
	}
	if source, ok := s.calculator.(catalogDataSource); ok {
		return source.CatalogData(ctx)
	}
	return s.data, ctx.Err()
}
