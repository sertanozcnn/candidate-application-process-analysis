package interaction

type Session struct {
	ID           string
	PositionCode string
}

type Event struct {
	EventID   string         `json:"event_id"`
	Sequence  int            `json:"sequence"`
	ElapsedMS int64          `json:"elapsed_ms"`
	Type      string         `json:"type"`
	FieldCode string         `json:"field_code,omitempty"`
	Metadata  map[string]any `json:"metadata"`
}

type BatchResult struct {
	Accepted  int `json:"accepted"`
	Duplicate int `json:"duplicate"`
	Rejected  int `json:"rejected"`
}
