package grpcapi

import (
	"strconv"

	pb "github.com/Hecatoncheir/pob2_engine/gen/go/pob2/v1"
	"github.com/Hecatoncheir/pob2_engine/internal/build"
)

func wirePassiveAttributes(nodes []int, values map[string]build.PassiveAttribute) map[uint32]string {
	if len(values) == 0 {
		return nil
	}
	result := make(map[uint32]string, len(values))
	for _, id := range nodes {
		if value := values[strconv.Itoa(id)]; value != "" {
			result[uint32(id)] = string(value)
		}
	}
	return result
}

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
	for _, attribute := range change.Attributes {
		result.Attributes = append(result.Attributes, &pb.PassiveAttributeChange{
			NodeId: uint32(attribute.NodeID), Before: string(attribute.Before), After: string(attribute.After),
		})
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
	for _, attribute := range change.GetAttributes() {
		result.Attributes = append(result.Attributes, build.PassiveAttributeChange{
			NodeID: int(attribute.GetNodeId()), Before: build.PassiveAttribute(attribute.GetBefore()), After: build.PassiveAttribute(attribute.GetAfter()),
		})
	}
	return result
}
