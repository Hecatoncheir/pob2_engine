package grpcapi

import (
	"context"

	pb "github.com/Hecatoncheir/pob2_engine/gen/go/pob2/v1"
	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/domain"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const maxMessageSize = 16 << 20

func (s *Server) buildCommands() (application.BuildCommands, error) {
	commands, ok := s.commands.(application.BuildCommands)
	if !ok {
		return nil, status.Error(codes.Unimplemented, "сервис не поддерживает полные сборки")
	}
	return commands, nil
}

func (s *Server) ApplyBuild(ctx context.Context, req *pb.ApplyBuildRequest) (*pb.ApplyBuildResponse, error) {
	commands, err := s.buildCommands()
	if err != nil {
		return nil, err
	}
	edit, err := localEdit(req.GetEdit())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	value, err := commands.Apply(ctx, application.ApplyBuild{SessionID: req.GetSessionId(), ExpectedRevision: req.GetExpectedRevision(), RequestID: req.GetRequestId(), XML: req.GetXml(), Edit: edit})
	if err != nil {
		return nil, rpcError(err)
	}
	message, err := snapshot(value)
	return &pb.ApplyBuildResponse{Snapshot: message}, rpcError(err)
}

func (s *Server) ExportBuild(ctx context.Context, req *pb.ExportBuildRequest) (*pb.ExportBuildResponse, error) {
	commands, err := s.buildCommands()
	if err != nil {
		return nil, err
	}
	value, err := commands.Export(ctx, req.GetSessionId())
	if err != nil {
		return nil, rpcError(err)
	}
	return &pb.ExportBuildResponse{SessionId: value.SessionID, Revision: value.Revision, Xml: value.XML, InputSha256: value.InputSHA256}, nil
}

func (s *Server) CompareBuild(ctx context.Context, req *pb.CompareBuildRequest) (*pb.CompareBuildResponse, error) {
	commands, err := s.buildCommands()
	if err != nil {
		return nil, err
	}
	edit, err := localEdit(req.GetEdit())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	value, err := commands.Compare(ctx, application.CompareBuild{SessionID: req.GetSessionId(), ExpectedRevision: req.GetExpectedRevision(), XML: req.GetXml(), Edit: edit})
	if err != nil {
		return nil, rpcError(err)
	}
	comparison := &pb.BuildComparison{Before: wireResult(value.Comparison.Before), After: wireResult(value.Comparison.After)}
	for _, delta := range value.Comparison.Deltas {
		comparison.Deltas = append(comparison.Deltas, &pb.StatDelta{Stat: delta.Stat, Before: delta.Before, After: delta.After, Difference: delta.Difference})
	}
	return &pb.CompareBuildResponse{SessionId: value.SessionID, Revision: value.Revision, Comparison: comparison}, nil
}

type calculationAnswer struct {
	value domain.Snapshot
	err   error
}

func (s *Server) CalculateBuild(req *pb.CalculateBuildRequest, stream grpc.ServerStreamingServer[pb.CalculateBuildResponse]) error {
	commands, err := s.buildCommands()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(stream.Context())
	ctx, updates := application.WithProgress(ctx)
	md, _ := metadata.FromIncomingContext(ctx)
	go s.watchLease(ctx, cancel, md.Get("pob2-lease")[0], true)
	answers := make(chan calculationAnswer, 1)
	finished := make(chan struct{})
	go func() {
		defer close(finished)
		value, err := commands.Calculate(ctx, application.CalculateBuild{SessionID: req.GetSessionId(), ExpectedRevision: req.GetExpectedRevision(), RequestID: req.GetRequestId()})
		answers <- calculationAnswer{value: value, err: err}
	}()
	// Счётчик lifecycle остаётся занят до фактического окончания работы, включая запись.
	defer func() { cancel(); <-finished }()
	for {
		select {
		case <-ctx.Done():
			return rpcError(ctx.Err())
		case update := <-updates:
			if update.Stage != "completed" {
				if err := stream.Send(&pb.CalculateBuildResponse{Stage: update.Stage}); err != nil {
					return err
				}
			}
		case answer := <-answers:
			if answer.err != nil {
				return rpcError(answer.err)
			}
			value, err := snapshot(answer.value)
			if err != nil {
				return rpcError(err)
			}
			return stream.Send(&pb.CalculateBuildResponse{Stage: "completed", Snapshot: value})
		}
	}
}
