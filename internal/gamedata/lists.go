package gamedata

import "slices"

func (d *Dataset) Gems() []Gem {
	result := slices.Clone(d.catalog.Gems)
	for i := range result {
		result[i].Levels = slices.Clone(result[i].Levels)
		result[i].ConstantStats = slices.Clone(result[i].ConstantStats)
	}
	return result
}

func (d *Dataset) Items() []Item { return slices.Clone(d.catalog.Items) }

func (d *Dataset) Nodes() []Node {
	result := slices.Clone(d.catalog.Nodes)
	for i := range result {
		result[i].Stats = slices.Clone(result[i].Stats)
		result[i].Neighbors = slices.Clone(result[i].Neighbors)
	}
	return result
}
