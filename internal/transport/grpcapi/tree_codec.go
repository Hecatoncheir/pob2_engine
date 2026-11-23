package grpcapi

import (
	"strconv"

	pb "github.com/Hecatoncheir/pob2_engine/gen/go/pob2/v1"
	"github.com/Hecatoncheir/pob2_engine/internal/build"
)

func wirePassiveChoices[T ~string](nodes []int, values map[string]T) map[uint32]string {
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
	for _, allocation := range change.Allocations {
		result.Allocations = append(result.Allocations, &pb.PassiveAllocationChange{
			NodeId: uint32(allocation.NodeID), Before: string(allocation.Before), After: string(allocation.After),
		})
	}
	if change.Points != nil {
		result.Points = &pb.PassivePointsChange{
			Before: wirePassivePoints(change.Points.Before), After: wirePassivePoints(change.Points.After),
		}
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
	for _, allocation := range change.GetAllocations() {
		result.Allocations = append(result.Allocations, build.PassiveAllocationChange{
			NodeID: int(allocation.GetNodeId()), Before: build.TreeAllocationMode(allocation.GetBefore()),
			After: build.TreeAllocationMode(allocation.GetAfter()),
		})
	}
	if points := change.GetPoints(); points != nil {
		result.Points = &build.PassivePointsChange{
			Before: localPassivePoints(points.GetBefore()), After: localPassivePoints(points.GetAfter()),
		}
	}
	return result
}

func wirePassivePoints(points build.PassivePoints) *pb.PassivePoints {
	return &pb.PassivePoints{
		Passives: uint32(points.Passives), WeaponSet_1: uint32(points.WeaponSet1), WeaponSet_2: uint32(points.WeaponSet2),
	}
}

func localPassivePoints(points *pb.PassivePoints) build.PassivePoints {
	return build.PassivePoints{
		Passives: int(points.GetPassives()), WeaponSet1: int(points.GetWeaponSet_1()), WeaponSet2: int(points.GetWeaponSet_2()),
	}
}
