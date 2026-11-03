package services

import (
	"context"

	"github.com/phosphoriat/invest-pool-go/internal/models"
	"github.com/phosphoriat/invest-pool-go/internal/repository"
)

// actions shares the transaction boundary; domain services own individual rules
type actions struct {
	repo *repository.ProblemRepository
	bus  *Bus
}

// commit Persist aggregate changes and their event atomically, then wake the dispatcher
func (a actions) commit(ctx context.Context, id string, fn func(*models.Problem) (*models.Event, error)) (string, error) {
	if err := a.repo.Change(ctx, id, fn); err != nil {
		return "", err
	}

	if a.bus != nil {
		a.bus.Notify()
	}

	return id, nil
}

// update Check role, aggregate existence and optional ownership within the transaction
func (a actions) update(ctx context.Context, u models.User, id, role string, ownerOnly bool, fn func(*models.Problem) (*models.Event, error)) (string, error) {
	if u.Role != role {
		return "", fail(403, "Действие недоступно этой роли")
	}

	if id == "" {
		return "", fail(400, "Нужен ID проблемы")
	}

	return a.commit(ctx, id, func(p *models.Problem) (*models.Event, error) {
		if p.RegionID == "" {
			return nil, fail(404, "Проблема не найдена")
		}

		if ownerOnly && p.RegionID != u.ID {
			return nil, fail(403, "Чужой регион")
		}

		return fn(p)
	})
}
