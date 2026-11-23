package build

// TreeScene описывает один кадр исходного интерфейса Lua PoB в порядке отрисовки.
type TreeScene struct {
	Width    float64           `json:"width"`
	Height   float64           `json:"height"`
	Scale    float64           `json:"scale"`
	OffsetX  float64           `json:"offsetX"`
	OffsetY  float64           `json:"offsetY"`
	Commands []TreeDrawCommand `json:"commands"`
}

type TreeDrawCommand struct {
	File      string    `json:"file"`
	Index     uint32    `json:"index"`
	Positions []float64 `json:"positions"`
	UV        []float64 `json:"uv"`
	Color     []float64 `json:"color"`
	Blend     string    `json:"blend"`
	Clamp     bool      `json:"clamp"`
}
