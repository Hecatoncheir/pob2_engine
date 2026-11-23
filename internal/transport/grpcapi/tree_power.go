package grpcapi

import (
	"context"

	pb "github.com/Hecatoncheir/pob2_engine/gen/go/pob2/v1"
	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/build"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) GetTreePower(ctx context.Context, req *pb.GetTreePowerRequest) (*pb.GetTreePowerResponse, error) {
	reader, ok := s.commands.(application.TreePowerCommands)
	if !ok {
		return nil, status.Error(codes.Unimplemented, "сервис не поддерживает расчёт влияния узлов")
	}
	query := build.TreePowerQuery{Metric: req.GetMetric(), MaxDistance: req.GetMaxDistance(), Allocation: req.GetAllocation()}
	view, err := reader.TreePower(ctx, req.GetSessionId(), req.GetExpectedRevision(), query)
	if err != nil {
		if ctx.Err() != nil {
			return nil, status.FromContextError(ctx.Err()).Err()
		}
		return nil, rpcError(err)
	}
	return wireTreePower(view), nil
}

func wireTreePower(view application.TreePowerView) *pb.GetTreePowerResponse {
	result := &pb.GetTreePowerResponse{Revision: view.Revision}
	if view.Metric != nil {
		result.Metric = &pb.TreePowerMetric{Id: view.Metric.ID, Label: view.Metric.Label}
	}
	for _, metric := range view.Metrics {
		result.Metrics = append(result.Metrics, &pb.TreePowerMetric{Id: metric.ID, Label: metric.Label})
	}
	for _, entry := range view.Entries {
		result.Entries = append(result.Entries, &pb.TreePowerEntry{
			Id: entry.ID, Name: entry.Name, Kind: entry.Kind, Allocated: entry.Allocated, Distance: entry.Distance,
			Power: entry.Power, PathPower: entry.PathPower, PowerText: entry.PowerText, PathPowerText: entry.PathPowerText,
		})
	}
	return result
}

func (c *Client) TreePower(ctx context.Context, id string, expected uint64, query build.TreePowerQuery) (application.TreePowerView, error) {
	reply, err := c.rpc.GetTreePower(ctx, &pb.GetTreePowerRequest{
		SessionId: id, ExpectedRevision: expected, Metric: query.Metric, MaxDistance: query.MaxDistance, Allocation: query.Allocation,
	})
	if err != nil {
		return application.TreePowerView{}, err
	}
	return localTreePower(reply), nil
}

func localTreePower(reply *pb.GetTreePowerResponse) application.TreePowerView {
	view := application.TreePowerView{Revision: reply.GetRevision()}
	if metric := reply.GetMetric(); metric != nil {
		view.Metric = &build.TreePowerMetric{ID: metric.GetId(), Label: metric.GetLabel()}
	}
	for _, metric := range reply.GetMetrics() {
		view.Metrics = append(view.Metrics, build.TreePowerMetric{ID: metric.GetId(), Label: metric.GetLabel()})
	}
	for _, entry := range reply.GetEntries() {
		view.Entries = append(view.Entries, build.TreePowerEntry{
			ID: entry.GetId(), Name: entry.GetName(), Kind: entry.GetKind(), Allocated: entry.GetAllocated(), Distance: entry.GetDistance(),
			Power: entry.GetPower(), PathPower: entry.GetPathPower(), PowerText: entry.GetPowerText(), PathPowerText: entry.GetPathPowerText(),
		})
	}
	return view
}
