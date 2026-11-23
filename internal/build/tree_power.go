package build

import "fmt"

type TreePowerQuery struct {
	Metric       string `json:"metric,omitempty"`
	MaxDistance  uint32 `json:"max_distance,omitempty"`
	Allocation   string `json:"allocation,omitempty"`
	IncludeScene bool   `json:"include_scene,omitempty"`
}

func (query TreePowerQuery) Validate() error {
	if len(query.Metric) > 100 || query.MaxDistance > 200 {
		return fmt.Errorf("неверная характеристика или расстояние до узла")
	}
	switch query.Allocation {
	case "", "all", "allocated", "unallocated":
		return nil
	default:
		return fmt.Errorf("режим фильтра: all, allocated или unallocated")
	}
}

type TreePower struct {
	Metric  *TreePowerMetric  `json:"metric,omitempty"`
	Metrics []TreePowerMetric `json:"metrics,omitempty"`
	Entries []TreePowerEntry  `json:"entries,omitempty"`
	Scene   *TreeScene        `json:"scene,omitempty"`
}

type TreePowerMetric struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

type TreePowerEntry struct {
	ID            uint32  `json:"id"`
	Name          string  `json:"name"`
	Kind          string  `json:"kind"`
	Allocated     bool    `json:"allocated"`
	Distance      uint32  `json:"distance"`
	Power         float64 `json:"power"`
	PathPower     float64 `json:"path_power"`
	PowerText     string  `json:"power_text"`
	PathPowerText string  `json:"path_power_text"`
}
