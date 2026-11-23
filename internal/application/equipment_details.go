package application

import (
	"context"
	"encoding/json"

	"github.com/Hecatoncheir/pob2_engine/internal/build"
)

type EquipmentDetailsView struct {
	Revision uint64          `json:"revision"`
	Details  json.RawMessage `json:"details"`
}

type EquipmentDetailsCommands interface {
	EquipmentDetails(context.Context, string, uint64, build.EquipmentDetailsQuery) (EquipmentDetailsView, error)
}

type equipmentDetailsReader interface {
	ReadEquipmentDetails(context.Context, *build.Document, build.EquipmentDetailsQuery) (json.RawMessage, error)
}

func (s *Service) SupportsEquipmentDetails() bool {
	_, ok := s.calculator.(equipmentDetailsReader)
	return ok
}

func (s *Service) EquipmentDetails(ctx context.Context, id string, expected uint64, query build.EquipmentDetailsQuery) (EquipmentDetailsView, error) {
	if err := query.Validate(); err != nil {
		return EquipmentDetailsView{}, failure(InvalidArgument, err.Error())
	}
	doc, snapshot, err := s.buildDocumentAtRevision(ctx, id, expected)
	if err != nil {
		return EquipmentDetailsView{}, err
	}
	reader, ok := s.calculator.(equipmentDetailsReader)
	if !ok {
		return EquipmentDetailsView{}, failure(InvalidArgument, "описание предмета доступно с исходным Lua PoB")
	}
	details, err := reader.ReadEquipmentDetails(ctx, doc, query)
	if err != nil {
		return EquipmentDetailsView{}, failure(CalculationFailure, err.Error())
	}
	if len(details) == 0 {
		return EquipmentDetailsView{}, failure(InvalidArgument, "PoB не распознал предмет или слот")
	}
	return EquipmentDetailsView{Revision: snapshot.Revision, Details: details}, nil
}
