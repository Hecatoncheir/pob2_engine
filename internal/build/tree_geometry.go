package build

// TreeGeometry содержит координаты и связи активной спецификации Lua PoB.
type TreeGeometry struct {
	Version string             `json:"version"`
	Nodes   []TreeGeometryNode `json:"nodes"`
	Edges   []TreeGeometryEdge `json:"edges"`
}

type TreeGeometryNode struct {
	ID             int     `json:"id"`
	X              float64 `json:"x"`
	Y              float64 `json:"y"`
	Name           string  `json:"name"`
	Kind           string  `json:"kind"`
	Icon           string  `json:"icon,omitempty"`
	Allocated      bool    `json:"allocated"`
	Attribute      bool    `json:"attribute"`
	AllocationMode int     `json:"allocation_mode"`
}

type TreeGeometryEdge struct {
	FromNode int `json:"from_node"`
	ToNode   int `json:"to_node"`
}
