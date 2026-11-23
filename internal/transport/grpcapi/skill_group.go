package grpcapi

import (
	pb "github.com/Hecatoncheir/pob2_engine/gen/go/pob2/v1"
	"github.com/Hecatoncheir/pob2_engine/internal/build"
)

func wireSkillGroup(value build.SkillGroupEdit) *pb.SkillGroupEdit {
	result := &pb.SkillGroupEdit{GroupId: uint32(value.GroupID), Label: value.Label, Delete: value.Delete}
	for _, gem := range value.Gems {
		result.Gems = append(result.Gems, &pb.SkillGroupGem{Id: gem.ID, Level: uint32(gem.Level), Quality: uint32(gem.Quality),
			Enabled: gem.Enabled, SourceIndex: uint32(gem.SourceIndex)})
	}
	return result
}

func localSkillGroup(value *pb.SkillGroupEdit) *build.SkillGroupEdit {
	if value == nil {
		return &build.SkillGroupEdit{}
	}
	result := &build.SkillGroupEdit{GroupID: int(value.GetGroupId()), Label: value.Label, Delete: value.GetDelete()}
	for _, gem := range value.GetGems() {
		result.Gems = append(result.Gems, build.SkillGroupGem{ID: gem.GetId(), Level: int(gem.GetLevel()), Quality: int(gem.GetQuality()),
			Enabled: gem.Enabled, SourceIndex: int(gem.GetSourceIndex())})
	}
	return result
}
