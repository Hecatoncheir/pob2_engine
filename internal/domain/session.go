package domain

// Snapshot — независимый снимок сессии; ревизия увеличивается после каждой правки.
type Snapshot struct {
	SessionID string `json:"session_id"`
	BuildName string `json:"build_name"`
	Revision  uint64 `json:"revision"`
}
