package services

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/phosphoriat/invest-pool-go/internal/models"
	"github.com/phosphoriat/invest-pool-go/internal/repository"
	"github.com/phosphoriat/invest-pool-go/internal/utils"
)

type ProblemsService struct {
	actions actions
}

// NewProblemsService Create a new ProblemsService instance
func NewProblemsService(repo *repository.ProblemRepository, bus *Bus) *ProblemsService {
	return &ProblemsService{actions: actions{repo: repo, bus: bus}}
}

// Publish regional problem and record ProblemPublished
func (s *ProblemsService) Publish(ctx context.Context, u models.User, text string) (string, error) {
	text = strings.TrimSpace(text)
	if u.Role != "region" {
		return "", fail(403, "Действие недоступно этой роли")
	}

	return s.actions.commit(ctx, utils.NewID(), func(p *models.Problem) (*models.Event, error) {
		if n := utf8.RuneCountInString(text); n < 3 || n > 1000 {
			return nil, fail(422, "Название: 3–1000 символов")
		}

		p.RegionID = u.ID
		p.Title = text
		return newEvent(p.ID, "ProblemPublished", map[string]any{"region_id": u.ID, "title": p.Title})
	})
}
