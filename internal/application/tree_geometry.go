package application

import (
	"context"

	"github.com/Hecatoncheir/pob2_engine/internal/build"
)

type TreeGeometryView struct {
	Revision uint64 `json:"revision"`
	build.TreeGeometry
}

type TreeGeometryCommands interface {
	TreeGeometry(context.Context, string, uint64) (TreeGeometryView, error)
}

type treeGeometryReader interface {
	ReadTreeGeometry(context.Context, *build.Document) (build.TreeGeometry, error)
}

func (s *Service) SupportsTreeGeometry() bool {
	_, ok := s.calculator.(treeGeometryReader)
	return ok
}

func (s *Service) TreeGeometry(ctx context.Context, id string, expected uint64) (TreeGeometryView, error) {
	snapshot, err := s.Snapshot(ctx, id)
	if err != nil {
		return TreeGeometryView{}, err
	}
	if expected != 0 && snapshot.Revision != expected {
		return TreeGeometryView{}, failure(RevisionConflict, "сборка изменилась, запросите дерево новой ревизии")
	}
	reader, ok := s.calculator.(treeGeometryReader)
	if !ok {
		return TreeGeometryView{}, failure(InvalidArgument, "геометрия дерева доступна с исходным Lua PoB")
	}
	doc, err := build.Parse([]byte(snapshot.BuildXML))
	if err != nil {
		return TreeGeometryView{}, failure(InvalidArgument, err.Error())
	}
	tree, err := reader.ReadTreeGeometry(ctx, doc)
	if err != nil {
		return TreeGeometryView{}, failure(CalculationFailure, err.Error())
	}
	return TreeGeometryView{Revision: snapshot.Revision, TreeGeometry: tree}, nil
}
