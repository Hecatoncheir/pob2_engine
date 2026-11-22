package gamedata

import (
	"fmt"
	"slices"
	"strings"
)

// RuntimeCatalog содержит только справочные данные оригинального PoB, без формул Go.
type RuntimeCatalog struct {
	TreeVersion string  `json:"tree_version"`
	Classes     []Class `json:"classes"`
	Gems        []Gem   `json:"gems"`
	Items       []Item  `json:"items"`
	Nodes       []Node  `json:"nodes"`
}

func NewRuntimeCatalog(value RuntimeCatalog, id, hash, commit string) (*Dataset, error) {
	if value.TreeVersion == "" || len(value.Classes) == 0 || len(value.Gems) == 0 || len(value.Items) == 0 || len(value.Nodes) == 0 {
		return nil, fmt.Errorf("PoB вернул неполный каталог")
	}
	data := &Dataset{id: id, tree: value.TreeVersion, hash: hash, commit: commit, runtime: true,
		catalog: catalog{Classes: value.Classes, Gems: value.Gems, Items: value.Items, Nodes: value.Nodes}}
	data.catalog.Classes = data.Classes()
	data.catalog.Gems = data.Gems()
	data.catalog.Items = data.Items()
	data.catalog.Nodes = data.Nodes()
	slices.SortFunc(data.catalog.Gems, func(a, b Gem) int { return strings.Compare(a.ID, b.ID) })
	return data, nil
}

func (d *Dataset) IsRuntime() bool { return d.runtime }
