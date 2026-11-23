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
	doc, revision, err := s.treeDocument(ctx, id, expected)
	if err != nil {
		return TreeGeometryView{}, err
	}
	reader, ok := s.calculator.(treeGeometryReader)
	if !ok {
		return TreeGeometryView{}, failure(InvalidArgument, "геометрия дерева доступна с исходным Lua PoB")
	}
	tree, err := reader.ReadTreeGeometry(ctx, doc)
	if err != nil {
		return TreeGeometryView{}, failure(CalculationFailure, err.Error())
	}
	return TreeGeometryView{Revision: revision, TreeGeometry: tree}, nil
}

func (s *Service) treeDocument(ctx context.Context, id string, expected uint64) (*build.Document, uint64, error) {
	snapshot, err := s.Snapshot(ctx, id)
	if err != nil {
		return nil, 0, err
	}
	if expected != 0 && snapshot.Revision != expected {
		return nil, 0, failure(RevisionConflict, "сборка изменилась, запросите дерево новой ревизии")
	}
	doc, err := build.Parse([]byte(snapshot.BuildXML))
	if err != nil {
		return nil, 0, failure(InvalidArgument, err.Error())
	}
	return doc, snapshot.Revision, nil
}
