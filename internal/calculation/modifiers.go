package calculation

type Modifier struct {
	Stat   string  `json:"stat"`
	Kind   string  `json:"kind"`
	Value  float64 `json:"value"`
	Source string  `json:"source"`
}
