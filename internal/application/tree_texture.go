package application

import (
	"context"

	"github.com/Hecatoncheir/pob2_engine/internal/build"
)

type TreeTextureCommands interface {
	TreeTexture(context.Context, build.TreeTextureQuery) (build.TreeTexture, error)
}

type treeTextureReader interface {
	ReadTreeTexture(context.Context, build.TreeTextureQuery) (build.TreeTexture, error)
}

func (s *Service) SupportsTreeTexture() bool {
	_, ok := s.calculator.(treeTextureReader)
	return ok
}

func (s *Service) TreeTexture(ctx context.Context, query build.TreeTextureQuery) (build.TreeTexture, error) {
	if err := query.Validate(); err != nil {
		return build.TreeTexture{}, failure(InvalidArgument, err.Error())
	}
	reader, ok := s.calculator.(treeTextureReader)
	if !ok {
		return build.TreeTexture{}, failure(InvalidArgument, "текстуры доступны с установленным Lua PoB")
	}
	texture, err := reader.ReadTreeTexture(ctx, query)
	if err != nil {
		return build.TreeTexture{}, failure(CalculationFailure, err.Error())
	}
	return texture, nil
}
