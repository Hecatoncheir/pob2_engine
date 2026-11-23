package grpcapi

import (
	pb "github.com/Hecatoncheir/pob2_engine/gen/go/pob2/v1"
	"github.com/Hecatoncheir/pob2_engine/internal/build"
)

func wireTreeScene(scene *build.TreeScene) *pb.TreeScene {
	if scene == nil {
		return nil
	}
	result := &pb.TreeScene{
		Width: scene.Width, Height: scene.Height, Scale: scene.Scale, OffsetX: scene.OffsetX, OffsetY: scene.OffsetY,
		Commands: make([]*pb.TreeDrawCommand, 0, len(scene.Commands)),
	}
	for _, command := range scene.Commands {
		result.Commands = append(result.Commands, &pb.TreeDrawCommand{
			File: command.File, Index: command.Index, Positions: command.Positions, Uv: command.UV,
			Color: command.Color, Blend: command.Blend, Clamp: command.Clamp,
		})
	}
	return result
}

func localTreeScene(scene *pb.TreeScene) *build.TreeScene {
	if scene == nil {
		return nil
	}
	result := &build.TreeScene{
		Width: scene.GetWidth(), Height: scene.GetHeight(), Scale: scene.GetScale(), OffsetX: scene.GetOffsetX(), OffsetY: scene.GetOffsetY(),
		Commands: make([]build.TreeDrawCommand, 0, len(scene.GetCommands())),
	}
	for _, command := range scene.GetCommands() {
		result.Commands = append(result.Commands, build.TreeDrawCommand{
			File: command.GetFile(), Index: command.GetIndex(), Positions: command.GetPositions(), UV: command.GetUv(),
			Color: command.GetColor(), Blend: command.GetBlend(), Clamp: command.GetClamp(),
		})
	}
	return result
}
