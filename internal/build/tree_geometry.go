package build

// TreeGeometry содержит координаты и связи активной спецификации Lua PoB.
type TreeGeometry struct {
	Version string             `json:"version"`
	Nodes   []TreeGeometryNode `json:"nodes"`
	Edges   []TreeGeometryEdge `json:"edges"`
	Scene   *TreeScene         `json:"scene,omitempty"`
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
	DrawSize       float64 `json:"draw_size"`
}

type TreeGeometryEdge struct {
	FromNode int `json:"from_node"`
	ToNode   int `json:"to_node"`
}

// TreeNodeTooltip содержит строки оригинальной подсказки и список узлов пути.
type TreeNodeTooltip struct {
	NodeID uint32            `json:"node_id"`
	Lines  []TreeTooltipLine `json:"lines"`
	Path   []uint32          `json:"path,omitempty"`
	Points uint32            `json:"points,omitempty"`
	Note   string            `json:"note,omitempty"`
	Recipe []string          `json:"recipe,omitempty"`
	Scene  *TreeScene        `json:"scene,omitempty"`
}

type TreeTooltipLine struct {
	Size      float64 `json:"size,omitempty"`
	Text      string  `json:"text,omitempty"`
	Separator bool    `json:"separator,omitempty"`
}
