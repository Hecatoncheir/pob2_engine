package domain

// Snapshot — независимый снимок сессии; ревизия увеличивается после каждой правки.
type Snapshot struct {
	SessionID              string `json:"session_id"`
	BuildName              string `json:"build_name"`
	Revision               uint64 `json:"revision"`
	BuildXML               string `json:"build_xml,omitempty"`
	InputSHA256            string `json:"input_sha256,omitempty"`
	BuildRevision          uint64 `json:"build_revision,omitempty"`
	CalculationJSON        string `json:"calculation_json,omitempty"`
	CalculatedFromRevision uint64 `json:"calculated_from_revision,omitempty"`
}
