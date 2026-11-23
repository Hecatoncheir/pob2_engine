package grpcapi

import (
	"context"

	pb "github.com/Hecatoncheir/pob2_engine/gen/go/pob2/v1"
	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) GetBuildReport(ctx context.Context, req *pb.GetBuildReportRequest) (*pb.GetBuildReportResponse, error) {
	reader, ok := s.commands.(application.BuildReportCommands)
	if !ok {
		return nil, status.Error(codes.Unimplemented, "сервис не поддерживает подробный отчёт")
	}
	view, err := reader.BuildReport(ctx, req.GetSessionId(), req.GetExpectedRevision())
	if err != nil {
		return nil, rpcError(err)
	}
	return &pb.GetBuildReportResponse{Revision: view.Revision, ReportJson: view.Report}, nil
}

func (c *Client) BuildReport(ctx context.Context, id string, expected uint64) (application.BuildReportView, error) {
	reply, err := c.rpc.GetBuildReport(ctx, &pb.GetBuildReportRequest{SessionId: id, ExpectedRevision: expected})
	if err != nil {
		return application.BuildReportView{}, err
	}
	return application.BuildReportView{Revision: reply.GetRevision(), Report: reply.GetReportJson()}, nil
}
