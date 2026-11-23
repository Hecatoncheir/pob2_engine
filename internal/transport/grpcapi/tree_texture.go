package grpcapi

import (
	"context"

	pb "github.com/Hecatoncheir/pob2_engine/gen/go/pob2/v1"
	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/build"
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
	return &pb.GetTreeTextureResponse{
		Width: texture.Width, Height: texture.Height, MipLevel: texture.MipLevel, Png: texture.PNG,
	}, nil
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
