// Package calculation определяет результаты и интерфейс расчёта.
package calculation

import "github.com/Hecatoncheir/pob2_engine/internal/build"

// LocalPoBDataID обозначает расчёт из явно выбранного каталога PoB.
const LocalPoBDataID = "pob-lua-local"

type Result struct {
	Status      string             `json:"status"`
	DataID      string             `json:"data_id"`
	DataSHA256  string             `json:"data_sha256"`
	PoBCommit   string             `json:"pob_commit"`
	InputSHA256 string             `json:"input_sha256"`
	Skill       string             `json:"skill"`
	Mode        string             `json:"mode"`
	Stats       map[string]float64 `json:"stats"`
	Modifiers   []Modifier         `json:"modifiers"`
	Issues      []build.Issue      `json:"issues"`
}
