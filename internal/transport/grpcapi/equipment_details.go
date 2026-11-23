package grpcapi

import (
	"context"

	pb "github.com/Hecatoncheir/pob2_engine/gen/go/pob2/v1"
	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/build"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) GetEquipmentDetails(ctx context.Context, req *pb.GetEquipmentDetailsRequest) (*pb.GetEquipmentDetailsResponse, error) {
	reader, ok := s.commands.(application.EquipmentDetailsCommands)
	if !ok {
		return nil, status.Error(codes.Unimplemented, "сервис не поддерживает подробное описание предмета")
	}
	query := build.EquipmentDetailsQuery{
		Slot: req.GetSlot(), ItemID: req.GetItemId(), Raw: req.GetRaw(), Compare: req.GetCompare(),
	}
	view, err := reader.EquipmentDetails(ctx, req.GetSessionId(), req.GetExpectedRevision(), query)
	if err != nil {
		return nil, rpcError(err)
	}
	return &pb.GetEquipmentDetailsResponse{Revision: view.Revision, DetailsJson: view.Details}, nil
}

func (c *Client) EquipmentDetails(ctx context.Context, id string, expected uint64, query build.EquipmentDetailsQuery) (application.EquipmentDetailsView, error) {
	if err := query.Validate(); err != nil {
		return application.EquipmentDetailsView{}, err
	}
	request := &pb.GetEquipmentDetailsRequest{SessionId: id, ExpectedRevision: expected, Slot: query.Slot, Compare: query.Compare}
	if query.ItemID != 0 {
		request.Item = &pb.GetEquipmentDetailsRequest_ItemId{ItemId: query.ItemID}
	} else {
		request.Item = &pb.GetEquipmentDetailsRequest_Raw{Raw: query.Raw}
	}
	reply, err := c.rpc.GetEquipmentDetails(ctx, request)
	if err != nil {
		return application.EquipmentDetailsView{}, err
	}
	return application.EquipmentDetailsView{Revision: reply.GetRevision(), Details: reply.GetDetailsJson()}, nil
}
