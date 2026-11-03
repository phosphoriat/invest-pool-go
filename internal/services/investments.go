package services

import (
	"context"
	"strings"

	"github.com/phosphoriat/invest-pool-go/internal/models"
	"github.com/phosphoriat/invest-pool-go/internal/repository"
	"github.com/phosphoriat/invest-pool-go/internal/utils"
	"github.com/phosphoriat/invest-pool-go/internal/validation"
)

type InvestmentsService struct {
	actions actions
}

// NewInvestmentsService Create a new InvestmentsService instance
func NewInvestmentsService(repo *repository.ProblemRepository, bus *Bus) *InvestmentsService {
	return &InvestmentsService{actions: actions{repo: repo, bus: bus}}
}

// Invest Accept a contribution within the target and deduplicate by investor and operation key
func (s *InvestmentsService) Invest(ctx context.Context, u models.User, id string, amount int64, key string) (string, error) {
	key = strings.TrimSpace(key)

	return s.actions.update(ctx, u, id, "investor", false, func(p *models.Problem) (*models.Event, error) {
		pool, err := fundingPool(p, amount)
		if err != nil {
			return nil, err
		}
		if !validation.OperationKey(key) {
			return nil, fail(422, "Нужен ключ операции (1–128 байт)")
		}

		for _, i := range pool.Investments {
			if i.InvestorID == u.ID && i.Key == key {
				if i.Amount != amount {
					return nil, fail(409, "Ключ уже использован с другой суммой")
				}
				return nil, nil
			}
		}

		if pool.Status != "open" || amount > pool.Target-pool.Total {
			return nil, fail(409, "Набор закрыт или вклад превышает остаток")
		}

		inv := models.Investment{ID: utils.NewID(), InvestorID: u.ID, Amount: amount, Key: key}
		pool.Investments = append(pool.Investments, inv)
		pool.Total += amount
		if pool.Total == pool.Target {
			pool.Status = "funded"
		}

		return newEvent(p.ID, "InvestmentAccepted", map[string]any{
			"pool_id":       pool.ID,
			"investment_id": inv.ID,
			"investor_id":   u.ID,
			"amount":        amount,
			"total":         pool.Total,
		})
	})
}
