package grpcapi

import (
	"context"

	pb "github.com/Hecatoncheir/pob2_engine/gen/go/pob2/v1"
	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/build"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) GetTreeTexture(ctx context.Context, req *pb.GetTreeTextureRequest) (*pb.GetTreeTextureResponse, error) {
	reader, ok := s.commands.(application.TreeTextureCommands)
	if !ok {
		return nil, status.Error(codes.Unimplemented, "сервис не поддерживает чтение текстур дерева")
	}
	query := build.TreeTextureQuery{File: req.GetFile(), Index: req.GetIndex(), TargetSize: req.GetTargetSize()}
	texture, err := reader.TreeTexture(ctx, query)
	if err != nil {
		return nil, rpcError(err)
	}
	return treeTextureReply(texture), nil
}

func (s *Server) GetTreeTextures(req *pb.GetTreeTexturesRequest, stream grpc.ServerStreamingServer[pb.GetTreeTextureResponse]) error {
	reader, ok := s.commands.(application.TreeTextureBatchCommands)
	if !ok {
		return status.Error(codes.Unimplemented, "сервис не поддерживает пакетное чтение текстур")
	}
	queries := make([]build.TreeTextureQuery, 0, len(req.GetLayers()))
	for _, layer := range req.GetLayers() {
		queries = append(queries, build.TreeTextureQuery{File: req.GetFile(), Index: layer.GetIndex(), TargetSize: layer.GetTargetSize()})
	}
	err := reader.TreeTextures(stream.Context(), queries, func(texture build.TreeTexture) error {
		return stream.Send(treeTextureReply(texture))
	})
	return rpcError(err)
}

func treeTextureReply(texture build.TreeTexture) *pb.GetTreeTextureResponse {
	return &pb.GetTreeTextureResponse{Width: texture.Width, Height: texture.Height, MipLevel: texture.MipLevel, Png: texture.PNG}
}

func (c *Client) TreeTextures(ctx context.Context, file string, layers []*pb.TreeTextureLayer) (grpc.ServerStreamingClient[pb.GetTreeTextureResponse], error) {
	return c.rpc.GetTreeTextures(ctx, &pb.GetTreeTexturesRequest{File: file, Layers: layers})
}

func (c *Client) TreeTexture(ctx context.Context, query build.TreeTextureQuery) (build.TreeTexture, error) {
	reply, err := c.rpc.GetTreeTexture(ctx, &pb.GetTreeTextureRequest{
		File: query.File, Index: query.Index, TargetSize: query.TargetSize,
	})
	if err != nil {
		return build.TreeTexture{}, err
	}
	return build.TreeTexture{
		Width: reply.GetWidth(), Height: reply.GetHeight(), MipLevel: reply.GetMipLevel(), PNG: reply.GetPng(),
	}, nil
}
