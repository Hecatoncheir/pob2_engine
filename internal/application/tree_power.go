package application

import (
	"context"

	"github.com/Hecatoncheir/pob2_engine/internal/build"
)

type TreePowerView struct {
	Revision uint64 `json:"revision"`
	build.TreePower
}

type TreePowerCommands interface {
	TreePower(context.Context, string, uint64, build.TreePowerQuery) (TreePowerView, error)
}

type TreePowerProgressCommands interface {
	TreePowerWithProgress(context.Context, string, uint64, build.TreePowerQuery, func(uint32)) (TreePowerView, error)
}

type treePowerReader interface {
	ReadTreePower(context.Context, *build.Document, build.TreePowerQuery, func(uint32)) (*build.TreePower, error)
}

func (s *Service) SupportsTreePower() bool {
	_, ok := s.calculator.(treePowerReader)
	return ok
}

func (s *Service) TreePower(ctx context.Context, id string, expected uint64, query build.TreePowerQuery) (TreePowerView, error) {
	return s.TreePowerWithProgress(ctx, id, expected, query, nil)
}

func (s *Service) TreePowerWithProgress(ctx context.Context, id string, expected uint64, query build.TreePowerQuery, progress func(uint32)) (TreePowerView, error) {
	if err := query.Validate(); err != nil {
		return TreePowerView{}, failure(InvalidArgument, err.Error())
	}
	doc, snapshot, err := s.buildDocumentAtRevision(ctx, id, expected)
	if err != nil {
		return TreePowerView{}, err
	}
	reader, ok := s.calculator.(treePowerReader)
	if !ok {
		return TreePowerView{}, failure(InvalidArgument, "влияние узлов доступно с исходным Lua PoB")
	}
	if progress != nil {
		progress(0)
	}
	report, err := reader.ReadTreePower(ctx, doc, query, progress)
	if err != nil {
		if ctx.Err() != nil {
			return TreePowerView{}, ctx.Err()
		}
		return TreePowerView{}, failure(CalculationFailure, err.Error())
	}
	if report == nil {
		return TreePowerView{}, failure(InvalidArgument, "характеристика отсутствует в каталоге Lua PoB")
	}
	return TreePowerView{Revision: snapshot.Revision, TreePower: *report}, nil
}
