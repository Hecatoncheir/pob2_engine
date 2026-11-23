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
	TreeNodeTooltip(context.Context, string, uint64, uint32, bool) (TreeNodeTooltipView, error)
}

type treeTooltipReader interface {
	ReadTreeNodeTooltip(context.Context, *build.Document, uint32, bool) (*build.TreeNodeTooltip, error)
}

func (s *Service) TreeNodeTooltip(ctx context.Context, id string, expected uint64, node uint32, compare bool) (TreeNodeTooltipView, error) {
	doc, revision, err := s.treeDocument(ctx, id, expected)
	if err != nil {
		return TreeNodeTooltipView{}, err
	}
	reader, ok := s.calculator.(treeTooltipReader)
	if !ok {
		return TreeNodeTooltipView{}, failure(InvalidArgument, "подсказки дерева доступны с исходным Lua PoB")
	}
	tip, err := reader.ReadTreeNodeTooltip(ctx, doc, node, compare)
	if err != nil {
		return TreeNodeTooltipView{}, failure(CalculationFailure, err.Error())
	}
	if tip == nil {
		return TreeNodeTooltipView{}, failure(InvalidArgument, "узел отсутствует в активном дереве")
	}
	return TreeNodeTooltipView{Revision: revision, TreeNodeTooltip: *tip}, nil
}
