package services

import (
	"context"
	"fmt"

	"github.com/phosphoriat/invest-pool-go/internal/models"
	"github.com/phosphoriat/invest-pool-go/internal/repository"
	"github.com/phosphoriat/invest-pool-go/internal/utils"
)

type DistributionsService struct {
	actions actions
}

// NewDistributionsService Create a new DistributionsService instance
func NewDistributionsService(repo *repository.ProblemRepository, bus *Bus) *DistributionsService {
	return &DistributionsService{actions: actions{repo: repo, bus: bus}}
}

// Distribute Calculate and persist all payouts for a funded pool in one transaction.
// A pool has exactly one distribution, so the pool itself is the idempotency key:
// repeating the same profit is a no-op, a different profit is a conflict.
func (s *DistributionsService) Distribute(ctx context.Context, u models.User, id string, amount int64) (string, error) {
	return s.actions.update(ctx, u, id, "region", true, func(p *models.Problem) (*models.Event, error) {
		pool, err := fundingPool(p, amount)
		if err != nil {
			return nil, err
		}

		if pool.Distribution != nil {
			d := pool.Distribution
			if d.Profit == amount {
				return nil, nil
			}

			return nil, fail(409, fmt.Sprintf("Прибыль этого пилота уже зафиксирована: %d", d.Profit))
		}

		if pool.Status != "funded" || pool.Total != pool.Target {
			return nil, fail(409, "Пул ещё не набран")
		}

		payouts, err := Allocate(pool.Investments, amount)
		if err != nil {
			return nil, fmt.Errorf("allocate: %w", err)
		}

		pool.Distribution = &models.Distribution{ID: utils.NewID(), Profit: amount, Payouts: payouts}
		pool.Status = "distributed"

		return newEvent(p.ID, "ProfitDistributed", map[string]any{"pool_id": pool.ID, "distribution": pool.Distribution})
	})
}
