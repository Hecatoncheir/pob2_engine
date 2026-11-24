package application

import (
	"context"

	"github.com/Hecatoncheir/pob2_engine/internal/build"
)

const maxTreeTextureLayers = 256

type TreeTextureCommands interface {
	TreeTexture(context.Context, build.TreeTextureQuery) (build.TreeTexture, error)
}

type TreeTextureBatchCommands interface {
	TreeTextures(context.Context, []build.TreeTextureQuery, func(build.TreeTexture) error) error
}

type treeTextureBatchReader interface {
	ReadTreeTextures(context.Context, []build.TreeTextureQuery, func(build.TreeTexture) error) error
}

type treeTextureReader interface {
	ReadTreeTexture(context.Context, build.TreeTextureQuery) (build.TreeTexture, error)
}

func (s *Service) SupportsTreeTexture() bool {
	_, ok := s.calculator.(treeTextureReader)
	return ok
}

func (s *Service) SupportsTreeTextureBatch() bool {
	_, ok := s.calculator.(treeTextureBatchReader)
	return ok
}

func (s *Service) TreeTextures(ctx context.Context, queries []build.TreeTextureQuery, send func(build.TreeTexture) error) error {
	if err := validateTextureBatch(queries); err != nil {
		return err
	}
	reader, ok := s.calculator.(treeTextureBatchReader)
	if !ok {
		return failure(InvalidArgument, "пакет текстур доступен с установленным Lua PoB")
	}
	return reader.ReadTreeTextures(ctx, queries, send)
}

func validateTextureBatch(queries []build.TreeTextureQuery) error {
	if len(queries) == 0 || len(queries) > maxTreeTextureLayers {
		return failure(InvalidArgument, "в пакете должно быть от 1 до 256 слоёв")
	}
	for _, query := range queries {
		if query.File != queries[0].File {
			return failure(InvalidArgument, "слои пакета должны относиться к одному атласу")
		}
		if err := query.Validate(); err != nil {
			return failure(InvalidArgument, err.Error())
		}
	}
	return nil
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
