package services

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/phosphoriat/invest-pool-go/internal/models"
	"github.com/phosphoriat/invest-pool-go/internal/repository"
	"github.com/phosphoriat/invest-pool-go/internal/utils"
)

type SolutionsService struct {
	actions actions
}

// NewSolutionsService Create a new SolutionsService instance
func NewSolutionsService(repo *repository.ProblemRepository, bus *Bus) *SolutionsService {
	return &SolutionsService{actions: actions{repo: repo, bus: bus}}
}

// Propose Attach a technology proposal to a problem that has no confirmed match
func (s *SolutionsService) Propose(ctx context.Context, u models.User, id, text string) (string, error) {
	text = strings.TrimSpace(text)

	return s.actions.update(ctx, u, id, "technologist", false, func(p *models.Problem) (*models.Event, error) {
		if p.Match != nil {
			return nil, fail(409, "Решение уже выбрано")
		}

		if n := utf8.RuneCountInString(text); n < 3 || n > 4000 {
			return nil, fail(422, "Название: 3–4000 символов")
		}

		sol := models.Solution{ID: utils.NewID(), TechnologistID: u.ID, Description: text}
		p.Solutions = append(p.Solutions, sol)
		return newEvent(p.ID, "SolutionProposed", sol)

	})
}
