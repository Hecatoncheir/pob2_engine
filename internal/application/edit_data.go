package application

import (
	"context"

	"github.com/Hecatoncheir/pob2_engine/internal/gamedata"
)

type catalogDataSource interface {
	CatalogData(context.Context) (*gamedata.Dataset, error)
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
