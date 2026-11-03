package services

import (
	"context"
	"fmt"
	"strings"

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

// Distribute Calculate and persist all payouts for a funded pool in one idempotent transaction
func (s *DistributionsService) Distribute(ctx context.Context, u models.User, id string, amount int64, key string) (string, error) {
	key = strings.TrimSpace(key)

	return s.actions.update(ctx, u, id, "region", true, func(p *models.Problem) (*models.Event, error) {
		pool, err := fundingPool(p, amount, key)
		if err != nil {
			return nil, err
		}

		if pool.Distribution != nil {
			d := pool.Distribution
			if d.Key == key && d.Profit == amount {
				return nil, nil
			}

			return nil, fail(409, "Прибыль этого пилота уже зафиксирована")
		}

		if pool.Status != "funded" || pool.Total != pool.Target {
			return nil, fail(409, "Пул ещё не набран")
		}

		payouts, err := Allocate(pool.Investments, amount)
		if err != nil {
			return nil, fmt.Errorf("allocate: %w", err)
		}

		pool.Distribution = &models.Distribution{ID: utils.NewID(), Key: key, Profit: amount, Payouts: payouts}
		pool.Status = "distributed"

		return newEvent(p.ID, "ProfitDistributed", map[string]any{"pool_id": pool.ID, "distribution": pool.Distribution})
	})
}
