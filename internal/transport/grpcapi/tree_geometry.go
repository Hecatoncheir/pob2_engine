package grpcapi

import (
	"context"

	pb "github.com/Hecatoncheir/pob2_engine/gen/go/pob2/v1"
	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/build"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) GetTreeGeometry(ctx context.Context, req *pb.GetTreeGeometryRequest) (*pb.GetTreeGeometryResponse, error) {
	reader, ok := s.commands.(application.TreeGeometryCommands)
	if !ok {
		return nil, status.Error(codes.Unimplemented, "сервис не поддерживает чтение дерева")
	}
	tree, err := reader.TreeGeometry(ctx, req.GetSessionId(), req.GetExpectedRevision())
	if err != nil {
		return nil, rpcError(err)
	}
	result := &pb.GetTreeGeometryResponse{Revision: tree.Revision, Version: tree.Version}
	for _, node := range tree.Nodes {
		result.Nodes = append(result.Nodes, &pb.TreeGeometryNode{
			Id: uint32(node.ID), X: node.X, Y: node.Y, Name: node.Name, Kind: node.Kind, Icon: node.Icon,
			Allocated: node.Allocated, Attribute: node.Attribute, AllocationMode: uint32(node.AllocationMode), DrawSize: node.DrawSize,
		})
	}
	for _, edge := range tree.Edges {
		result.Edges = append(result.Edges, &pb.TreeGeometryEdge{
			FromNode: uint32(edge.FromNode), ToNode: uint32(edge.ToNode),
		})
	}
	return result, nil
}

func (c *Client) TreeGeometry(ctx context.Context, id string, expected uint64) (application.TreeGeometryView, error) {
	reply, err := c.rpc.GetTreeGeometry(ctx, &pb.GetTreeGeometryRequest{SessionId: id, ExpectedRevision: expected})
	if err != nil {
		return application.TreeGeometryView{}, err
	}
	tree := application.TreeGeometryView{Revision: reply.GetRevision()}
	tree.Version = reply.GetVersion()
	tree.Nodes = make([]build.TreeGeometryNode, 0, len(reply.GetNodes()))
	for _, node := range reply.GetNodes() {
		tree.Nodes = append(tree.Nodes, build.TreeGeometryNode{
			ID: int(node.GetId()), X: node.GetX(), Y: node.GetY(), Name: node.GetName(), Kind: node.GetKind(), Icon: node.GetIcon(),
			Allocated: node.GetAllocated(), Attribute: node.GetAttribute(), AllocationMode: int(node.GetAllocationMode()), DrawSize: node.GetDrawSize(),
		})
	}
	tree.Edges = make([]build.TreeGeometryEdge, 0, len(reply.GetEdges()))
	for _, edge := range reply.GetEdges() {
		tree.Edges = append(tree.Edges, build.TreeGeometryEdge{
			FromNode: int(edge.GetFromNode()), ToNode: int(edge.GetToNode()),
		})
	}
	return tree, nil
}
