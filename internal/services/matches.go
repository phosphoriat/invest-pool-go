package services

import (
	"context"

	"github.com/phosphoriat/invest-pool-go/internal/models"
	"github.com/phosphoriat/invest-pool-go/internal/repository"
)

type MatchesService struct {
	actions actions
}

// NewMatchesService Create a new MatchesService instance
func NewMatchesService(repo *repository.ProblemRepository, bus *Bus) *MatchesService {
	return &MatchesService{actions: actions{repo: repo, bus: bus}}
}

// Confirm the expert match and make retries of the same solution idempotent
func (s *MatchesService) Confirm(ctx context.Context, u models.User, id, solutionID string) (string, error) {
	return s.actions.update(ctx, u, id, "expert", false, func(p *models.Problem) (*models.Event, error) {
		if p.Match != nil {
			if p.Match.SolutionID == solutionID {
				return nil, nil
			}
			return nil, fail(409, "Другой матч уже подтверждён")
		}

		found := false
		for _, v := range p.Solutions {
			if v.ID == solutionID {
				found = true
			}
		}
		if !found {
			return nil, fail(422, "Решение не принадлежит проблеме")
		}

		p.Match = &models.Match{SolutionID: solutionID, ExpertID: u.ID}
		return newEvent(p.ID, "SolutionMatched", p.Match)
	})
}
