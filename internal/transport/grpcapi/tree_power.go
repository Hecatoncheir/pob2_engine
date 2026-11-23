package grpcapi

import (
	"context"

	pb "github.com/Hecatoncheir/pob2_engine/gen/go/pob2/v1"
	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/build"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func (s *Server) GetTreePower(ctx context.Context, req *pb.GetTreePowerRequest) (*pb.GetTreePowerResponse, error) {
	reader, ok := s.commands.(application.TreePowerCommands)
	if !ok {
		return nil, status.Error(codes.Unimplemented, "сервис не поддерживает расчёт влияния узлов")
	}
	view, err := reader.TreePower(ctx, req.GetSessionId(), req.GetExpectedRevision(), localPowerQuery(req))
	return treePowerReply(ctx, view, err)
}

func localPowerQuery(req *pb.GetTreePowerRequest) build.TreePowerQuery {
	return build.TreePowerQuery{
		Metric: req.GetMetric(), MaxDistance: req.GetMaxDistance(), Allocation: req.GetAllocation(), IncludeScene: req.GetIncludeScene(),
	}
}

func treePowerReply(ctx context.Context, view application.TreePowerView, err error) (*pb.GetTreePowerResponse, error) {
	if err != nil {
		if ctx.Err() != nil {
			return nil, status.FromContextError(ctx.Err()).Err()
		}
		return nil, rpcError(err)
	}
	return wireTreePower(view), nil
}

type treePowerResult struct {
	reply *pb.GetTreePowerResponse
	err   error
}

func (s *Server) WatchTreePower(req *pb.WatchTreePowerRequest, stream grpc.ServerStreamingServer[pb.WatchTreePowerResponse]) error {
	reader, ok := s.commands.(application.TreePowerProgressCommands)
	if !ok {
		return status.Error(codes.Unimplemented, "сервис не поддерживает подписку на расчёт влияния узлов")
	}
	query := req.GetQuery()
	if query.GetMetric() == "" {
		return status.Error(codes.InvalidArgument, "укажите характеристику для расчёта влияния узлов")
	}
	ctx, cancel := context.WithCancel(stream.Context())
	md, _ := metadata.FromIncomingContext(ctx)
	go s.watchLease(ctx, cancel, md.Get("pob2-lease")[0], false)
	progress, done := startTreePower(ctx, reader, query)
	result := sendTreePower(ctx, stream, progress, done)
	cancel()
	for range done {
		// При досрочном закрытии дождаться завершения Lua перед снятием задачи.
	}
	return result
}

func startTreePower(ctx context.Context, reader application.TreePowerProgressCommands, req *pb.GetTreePowerRequest) (<-chan uint32, <-chan treePowerResult) {
	progress := make(chan uint32, 1)
	done := make(chan treePowerResult, 1)
	go func() {
		defer close(done)
		view, err := reader.TreePowerWithProgress(ctx, req.GetSessionId(), req.GetExpectedRevision(), localPowerQuery(req), func(percent uint32) {
			select {
			case progress <- percent:
			default:
			}
		})
		reply, err := treePowerReply(ctx, view, err)
		done <- treePowerResult{reply: reply, err: err}
	}()
	return progress, done
}

func sendTreePower(ctx context.Context, stream grpc.ServerStreamingServer[pb.WatchTreePowerResponse], progress <-chan uint32, done <-chan treePowerResult) error {
	for {
		select {
		case <-ctx.Done():
			return status.FromContextError(ctx.Err()).Err()
		case percent := <-progress:
			if err := stream.Send(&pb.WatchTreePowerResponse{Update: &pb.WatchTreePowerResponse_Progress{Progress: percent}}); err != nil {
				return err
			}
		case result := <-done:
			if result.err != nil {
				return result.err
			}
			return stream.Send(&pb.WatchTreePowerResponse{Update: &pb.WatchTreePowerResponse_Result{Result: result.reply}})
		}
	}
}

func wireTreePower(view application.TreePowerView) *pb.GetTreePowerResponse {
	result := &pb.GetTreePowerResponse{Revision: view.Revision, Scene: wireTreeScene(view.Scene)}
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
	reply, err := c.rpc.GetTreePower(ctx, treePowerRequest(id, expected, query))
	if err != nil {
		return application.TreePowerView{}, err
	}
	return localTreePower(reply), nil
}

func (c *Client) WatchTreePower(ctx context.Context, id string, expected uint64, query build.TreePowerQuery) (grpc.ServerStreamingClient[pb.WatchTreePowerResponse], error) {
	return c.rpc.WatchTreePower(ctx, &pb.WatchTreePowerRequest{Query: treePowerRequest(id, expected, query)})
}

func treePowerRequest(id string, expected uint64, query build.TreePowerQuery) *pb.GetTreePowerRequest {
	return &pb.GetTreePowerRequest{SessionId: id, ExpectedRevision: expected, Metric: query.Metric,
		MaxDistance: query.MaxDistance, Allocation: query.Allocation, IncludeScene: query.IncludeScene}
}

func localTreePower(reply *pb.GetTreePowerResponse) application.TreePowerView {
	view := application.TreePowerView{Revision: reply.GetRevision()}
	view.Scene = localTreeScene(reply.GetScene())
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
