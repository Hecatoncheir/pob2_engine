package grpcapi

import (
	pb "github.com/Hecatoncheir/pob2_engine/gen/go/pob2/v1"
	"github.com/Hecatoncheir/pob2_engine/internal/build"
)

func wirePassiveChange(change *build.PassiveChange) *pb.PassiveChange {
	if change == nil {
		return nil
	}
	result := &pb.PassiveChange{}
	for _, id := range change.Added {
		result.Added = append(result.Added, uint32(id))
	}
	for _, id := range change.Removed {
		result.Removed = append(result.Removed, uint32(id))
	}
	return result
}

func localPassiveChange(change *pb.PassiveChange) *build.PassiveChange {
	if change == nil {
		return nil
	}
	result := &build.PassiveChange{}
	for _, id := range change.GetAdded() {
		result.Added = append(result.Added, int(id))
	}
	for _, id := range change.GetRemoved() {
		result.Removed = append(result.Removed, int(id))
	}
	return result
}
