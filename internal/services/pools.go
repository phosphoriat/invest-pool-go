package services

import (
	"context"

	"github.com/phosphoriat/invest-pool-go/internal/models"
	"github.com/phosphoriat/invest-pool-go/internal/repository"
	"github.com/phosphoriat/invest-pool-go/internal/utils"
	"github.com/phosphoriat/invest-pool-go/internal/validation"
)

type PoolsService struct {
	actions actions
}

// NewPoolsService Create a new PoolsService instance
func NewPoolsService(repo *repository.ProblemRepository, bus *Bus) *PoolsService {
	return &PoolsService{actions: actions{repo: repo, bus: bus}}
}

// Create Create a pool for the owning region after expert confirmation
func (s *PoolsService) Create(ctx context.Context, u models.User, id string, amount int64) (string, error) {
	return s.actions.update(ctx, u, id, "region", true, func(p *models.Problem) (*models.Event, error) {
		if p.Match == nil {
			return nil, fail(409, "Сначала эксперт подтверждает матч")
		}
		if p.Pool != nil {
			return nil, fail(409, "Пул уже существует")
		}
		if !validation.Money(amount) {
			return nil, fail(422, "Недопустимая цель")
		}

		p.Pool = &models.Pool{
			ID:          utils.NewID(),
			Target:      amount,
			Status:      "open",
			Investments: []models.Investment{},
		}

		return newEvent(p.ID, "PoolCreated", map[string]any{"pool_id": p.Pool.ID, "target": amount})
	})
}

// fundingPool Validate match, pool, amount and operation key before a financial operation.
func fundingPool(p *models.Problem, amount int64, key string) (*models.Pool, error) {
	if p.Match == nil || p.Pool == nil {
		return nil, fail(409, "Нет подтверждённого матча и пула")
	}
	if !validation.Money(amount) {
		return nil, fail(422, "Сумма должна быть положительным целым в допустимом диапазоне")
	}
	if !validation.OperationKey(key) {
		return nil, fail(422, "Нужен ключ операции (1–128 байт)")
	}

	return p.Pool, nil
}

// View Build role-filtered pool views and the public catalogue of open opportunities.
// Filters freshly decoded aggregates without mutating shared state.
func (s *PoolsService) View(ctx context.Context, u models.User) ([]models.Problem, []models.Opportunity, error) {
	all, err := s.actions.repo.Problems(ctx)
	if err != nil {
		return nil, nil, err
	}

	visible := []models.Problem{}
	opps := []models.Opportunity{}

	for _, p := range all {
		if u.Role == "investor" && p.Pool != nil && p.Pool.Status == "open" {
			opps = append(opps, models.Opportunity{ProblemID: p.ID, Title: p.Title, PoolID: p.Pool.ID, Target: p.Pool.Target, Total: p.Pool.Total})
		}

		switch u.Role {
		case "region":
			if p.RegionID != u.ID {
				continue
			}

		case "expert":
		case "technologist":
			own := []models.Solution{}
			matched := false

			for _, sol := range p.Solutions {
				if sol.TechnologistID == u.ID {
					own = append(own, sol)
					if p.Match != nil && p.Match.SolutionID == sol.ID {
						matched = true
					}
				}
			}

			p.Solutions = own

			if !matched {
				p.Pool = nil
				p.Match = nil
			} else if p.Pool != nil {
				p.Pool.Investments = nil

				if p.Pool.Distribution != nil {
					p.Pool.Distribution.Payouts = nil
					p.Pool.Distribution.Key = ""
				}
			}

		case "investor":
			if p.Pool == nil {
				continue
			}

			own := []models.Investment{}
			for _, inv := range p.Pool.Investments {
				if inv.InvestorID == u.ID {
					own = append(own, inv)
				}
			}
			if len(own) == 0 {
				continue
			}

			p.Pool.Investments = own

			if p.Pool.Distribution != nil {
				pays := []models.Payout{}

				for _, pay := range p.Pool.Distribution.Payouts {
					if pay.InvestorID == u.ID {
						pays = append(pays, pay)
					}
				}

				p.Pool.Distribution.Payouts = pays
				p.Pool.Distribution.Key = ""
			}

			selected := []models.Solution{}
			for _, sol := range p.Solutions {
				if p.Match != nil && sol.ID == p.Match.SolutionID {
					selected = append(selected, sol)
				}
			}

			p.Solutions = selected

		default:
			continue
		}

		visible = append(visible, p)
	}

	return visible, opps, nil
}
