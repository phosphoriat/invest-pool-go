package services

import (
	"context"
	"encoding/json"
	"time"

	"github.com/phosphoriat/invest-pool-go/internal/models"
	"github.com/phosphoriat/invest-pool-go/internal/repository"
	"github.com/phosphoriat/invest-pool-go/internal/utils"
)

type EventsService struct {
	repo  *repository.EventRepository
	pools *PoolsService
}

// NewEventsService Create a new EventsService instance
func NewEventsService(repo *repository.EventRepository, pools *PoolsService) *EventsService {
	return &EventsService{repo: repo, pools: pools}
}

// newEvent Create an event envelope with a stable ID, UTC timestamp and JSON payload
func newEvent(problemID, name string, payload any) (*models.Event, error) {
	raw, err := json.Marshal(payload)

	return &models.Event{
		ID:         utils.NewID(),
		Name:       name,
		ProblemID:  problemID,
		OccurredAt: time.Now().UTC().Format(time.RFC3339Nano),
		Payload:    raw,
	}, err
}

// Audit Return audit events only for problems visible to the authorized region or expert
func (s *EventsService) Audit(ctx context.Context, u models.User) ([]models.Event, error) {
	if u.Role != "region" && u.Role != "expert" {
		return nil, fail(403, "Аудит доступен региону и эксперту")
	}

	projects, _, err := s.pools.View(ctx, u)
	if err != nil {
		return nil, err
	}

	ids := map[string]bool{}
	for _, p := range projects {
		ids[p.ID] = true
	}

	events, err := s.repo.Events(ctx)
	if err != nil {
		return nil, err
	}

	out := []models.Event{}
	for _, e := range events {
		if ids[e.ProblemID] {
			out = append(out, e)
		}
	}

	return out, nil
}
