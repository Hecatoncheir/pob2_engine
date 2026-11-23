package application

import (
	"context"
	"fmt"

	"github.com/Hecatoncheir/pob2_engine/internal/build"
	"github.com/Hecatoncheir/pob2_engine/internal/gamedata"
)

type catalogDataSource interface {
	CatalogData(context.Context) (*gamedata.Dataset, error)
}

type equipmentValidator interface {
	ValidateEquipment(context.Context, *build.Document, string) error
}

type jewelValidator interface {
	ValidateJewel(context.Context, *build.Document, int) error
}

type treeEditor interface {
	ResolveTreeEdit(context.Context, *build.Document, build.TreeNodeEdit) (build.PassiveChange, error)
}

type treeSpecResolver interface {
	ResolveTreeSpec(context.Context, *build.Document) (build.TreeIdentity, error)
}

type treeClassEditor interface {
	ResolveTreeClass(context.Context, *build.Document, build.TreeClassEdit) (build.TreeClassChange, error)
}

// SupportsTreeClass сообщает, что класс можно сменить средствами Lua PoB.
func (s *Service) SupportsTreeClass() bool {
	_, ok := s.calculator.(treeClassEditor)
	return ok
}

// SupportsTreeSpecs сообщает, что выбор спецификации проверяет оригинальный PoB.
func (s *Service) SupportsTreeSpecs() bool {
	_, ok := s.calculator.(treeSpecResolver)
	return ok
}

// SupportsTreePreview сообщает, что движок может править дерево через PoB.
func (s *Service) SupportsTreePreview() bool {
	_, ok := s.calculator.(treeEditor)
	return ok
}

// SupportsEquipmentPreview сообщает, что движок проверяет экипировку через PoB.
func (s *Service) SupportsEquipmentPreview() bool {
	_, ok := s.calculator.(equipmentValidator)
	return ok
}

// SupportsJewelPreview сообщает, что оригинальный PoB проверяет самоцветы в гнёздах.
func (s *Service) SupportsJewelPreview() bool {
	_, ok := s.calculator.(jewelValidator)
	return ok
}

func (s *Service) editDocument(ctx context.Context, doc *build.Document, edit build.Edit) (*build.Document, *build.PassiveChange, error) {
	if edit.TreeNode != nil {
		return s.editTree(ctx, doc, edit)
	}
	if edit.TreeClass != nil {
		updated, err := s.editTreeClass(ctx, doc, edit)
		return updated, nil, err
	}
	if edit.Jewel != nil {
		updated, err := s.editJewel(ctx, doc, edit)
		return updated, nil, err
	}
	data, err := s.editData(ctx)
	if err != nil {
		return nil, nil, err
	}
	updated, err := doc.Edited(edit, data)
	if err != nil {
		return nil, nil, err
	}
	if validator, ok := s.calculator.(equipmentValidator); ok && edit.Equipment != nil {
		if err := validator.ValidateEquipment(ctx, updated, edit.Equipment.Slot); err != nil {
			return nil, nil, err
		}
	}
	if selectsTreeSpec(edit.Selection) {
		updated, err = s.resolveTreeSpec(ctx, updated)
	}
	return updated, nil, err
}

func (s *Service) editTreeClass(ctx context.Context, doc *build.Document, edit build.Edit) (*build.Document, error) {
	if err := build.ValidateEdit(edit); err != nil {
		return nil, err
	}
	editor, ok := s.calculator.(treeClassEditor)
	if !ok {
		return nil, fmt.Errorf("смена класса доступна с оригинальным Lua PoB")
	}
	change, err := editor.ResolveTreeClass(ctx, doc, *edit.TreeClass)
	if err != nil {
		return nil, err
	}
	return doc.WithResolvedTreeClass(change)
}

func selectsTreeSpec(edit *build.SelectionEdit) bool {
	return edit != nil && edit.Sets != nil && edit.Sets.Tree != 0
}

func (s *Service) resolveTreeSpec(ctx context.Context, doc *build.Document) (*build.Document, error) {
	resolver, ok := s.calculator.(treeSpecResolver)
	if !ok {
		return doc, nil
	}
	identity, err := resolver.ResolveTreeSpec(ctx, doc)
	if err != nil {
		return nil, err
	}
	return doc.WithTreeIdentity(identity)
}

func (s *Service) editJewel(ctx context.Context, doc *build.Document, edit build.Edit) (*build.Document, error) {
	validator, ok := s.calculator.(jewelValidator)
	if !ok {
		return nil, fmt.Errorf("правка самоцвета доступна с оригинальным Lua PoB")
	}
	data, err := s.editData(ctx)
	if err != nil {
		return nil, err
	}
	updated, err := doc.Edited(edit, data)
	if err != nil {
		return nil, err
	}
	if err := validator.ValidateJewel(ctx, updated, edit.Jewel.NodeID); err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *Service) editTree(ctx context.Context, doc *build.Document, edit build.Edit) (*build.Document, *build.PassiveChange, error) {
	if err := build.ValidateEdit(edit); err != nil {
		return nil, nil, err
	}
	editor, ok := s.calculator.(treeEditor)
	if !ok {
		return nil, nil, fmt.Errorf("правка узла доступна с оригинальным Lua PoB")
	}
	change, err := editor.ResolveTreeEdit(ctx, doc, *edit.TreeNode)
	if err != nil {
		return nil, nil, err
	}
	updated, err := doc.WithResolvedTreeEdit(change)
	if edit.TreeNode.Attribute == "" {
		change.Attributes = nil
	}
	return updated, &change, err
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
