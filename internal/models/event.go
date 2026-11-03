package models

import (
	"encoding/json"
)

type Event struct {
	ID         string          `json:"id"`
	Name       string          `json:"name"`
	ProblemID  string          `json:"problem_id"`
	OccurredAt string          `json:"occurred_at"`
	Payload    json.RawMessage `json:"payload"`
}
