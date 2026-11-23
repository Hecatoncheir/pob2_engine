package grpcapi

import (
	"context"

	pb "github.com/Hecatoncheir/pob2_engine/gen/go/pob2/v1"
	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/build"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) GetTreeNodeTooltip(ctx context.Context, req *pb.GetTreeNodeTooltipRequest) (*pb.GetTreeNodeTooltipResponse, error) {
	reader, ok := s.commands.(application.TreeTooltipCommands)
	if !ok {
		return nil, status.Error(codes.Unimplemented, "сервис не поддерживает подсказки дерева")
	}
	tip, err := reader.TreeNodeTooltip(ctx, req.GetSessionId(), req.GetExpectedRevision(), req.GetNodeId(), req.GetCompare())
	if err != nil {
		return nil, rpcError(err)
	}
	result := &pb.GetTreeNodeTooltipResponse{
		Revision: tip.Revision, NodeId: tip.NodeID, Path: tip.Path, Points: tip.Points, Note: tip.Note, Recipe: tip.Recipe,
	}
	for _, line := range tip.Lines {
		result.Lines = append(result.Lines, &pb.TreeTooltipLine{Size: line.Size, Text: line.Text, Separator: line.Separator})
	}
	return result, nil
}

func (c *Client) TreeNodeTooltip(ctx context.Context, id string, expected uint64, node uint32, compare bool) (application.TreeNodeTooltipView, error) {
	reply, err := c.rpc.GetTreeNodeTooltip(ctx, &pb.GetTreeNodeTooltipRequest{
		SessionId: id, ExpectedRevision: expected, NodeId: node, Compare: compare,
	})
	if err != nil {
		return application.TreeNodeTooltipView{}, err
	}
	tip := application.TreeNodeTooltipView{
		Revision: reply.GetRevision(),
		TreeNodeTooltip: build.TreeNodeTooltip{
			NodeID: reply.GetNodeId(), Path: reply.GetPath(), Points: reply.GetPoints(), Note: reply.GetNote(), Recipe: reply.GetRecipe(),
			Lines: make([]build.TreeTooltipLine, 0, len(reply.GetLines())),
		},
	}
	for _, line := range reply.GetLines() {
		tip.Lines = append(tip.Lines, build.TreeTooltipLine{Size: line.GetSize(), Text: line.GetText(), Separator: line.GetSeparator()})
	}
	return tip, nil
}
