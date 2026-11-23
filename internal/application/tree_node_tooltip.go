package application

import (
	"context"

	"github.com/Hecatoncheir/pob2_engine/internal/build"
)

type TreeNodeTooltipView struct {
	Revision uint64 `json:"revision"`
	build.TreeNodeTooltip
}

type TreeTooltipCommands interface {
	TreeNodeTooltip(ctx context.Context, id string, expected uint64, node uint32, compare, scene bool) (TreeNodeTooltipView, error)
}

type treeTooltipReader interface {
	ReadTreeNodeTooltip(ctx context.Context, doc *build.Document, node uint32, compare, scene bool) (*build.TreeNodeTooltip, error)
}

func (s *Service) SupportsTreeTooltip() bool {
	_, ok := s.calculator.(treeTooltipReader)
	return ok
}

func (s *Service) TreeNodeTooltip(ctx context.Context, id string, expected uint64, node uint32, compare, scene bool) (TreeNodeTooltipView, error) {
	doc, snapshot, err := s.buildDocumentAtRevision(ctx, id, expected)
	if err != nil {
		return TreeNodeTooltipView{}, err
	}
	reader, ok := s.calculator.(treeTooltipReader)
	if !ok {
		return TreeNodeTooltipView{}, failure(InvalidArgument, "подсказки дерева доступны с исходным Lua PoB")
	}
	tip, err := reader.ReadTreeNodeTooltip(ctx, doc, node, compare, scene)
	if err != nil {
		return TreeNodeTooltipView{}, failure(CalculationFailure, err.Error())
	}
	if tip == nil {
		return TreeNodeTooltipView{}, failure(InvalidArgument, "узел отсутствует в активном дереве")
	}
	return TreeNodeTooltipView{Revision: snapshot.Revision, TreeNodeTooltip: *tip}, nil
}
