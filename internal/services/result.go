package services

import "github.com/phosphoriat/invest-pool-go/internal/models"

// Result Describe what a command produced. Replayed marks an idempotent retry that changed nothing
// and returns the entity created by the original call.
type Result struct {
	ProblemID    string               `json:"problem_id"`
	SolutionID   string               `json:"solution_id,omitempty"`
	PoolID       string               `json:"pool_id,omitempty"`
	InvestmentID string               `json:"investment_id,omitempty"`
	Total        int64                `json:"total,omitempty"`
	Distribution *models.Distribution `json:"distribution,omitempty"`
	Replayed     bool                 `json:"replayed"`
}
