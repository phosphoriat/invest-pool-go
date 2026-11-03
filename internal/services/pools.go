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

// View Build role-filtered read models and the public catalogue of open opportunities.
// Works on freshly decoded aggregates, so filtering never touches shared state.
func (s *PoolsService) View(ctx context.Context, u models.User) ([]models.ProblemView, []models.Opportunity, error) {
	all, err := s.actions.repo.Problems(ctx)
	if err != nil {
		return nil, nil, err
	}

	visible := []models.ProblemView{}
	opps := []models.Opportunity{}

	for _, p := range all {
		if u.Role == "investor" && p.Pool != nil && p.Pool.Status == "open" {
			opps = append(opps, models.Opportunity{ProblemID: p.ID, Title: p.Title, PoolID: p.Pool.ID, Target: p.Pool.Target, Total: p.Pool.Total})
		}

		v := models.NewProblemView(p)

		switch u.Role {
		case "region":
			if p.RegionID != u.ID {
				continue
			}

		case "expert":
		case "technologist":
			own := []models.Solution{}
			selected := false

			for _, sol := range v.Solutions {
				if sol.TechnologistID == u.ID {
					own = append(own, sol)
					if p.Match != nil && p.Match.SolutionID == sol.ID {
						selected = true
					}
				}
			}

			v.Solutions = own

			// Matched stays true: a losing technologist learns the problem is closed, but not who won
			if !selected {
				v.Pool = nil
				v.Match = nil
			} else if v.Pool != nil {
				v.Pool.Investments = nil

				if d := v.Pool.Distribution; d != nil {
					v.Pool.Distribution = &models.Distribution{ID: d.ID, Profit: d.Profit}
				}
			}

		case "investor":
			if v.Pool == nil {
				continue
			}

			own := []models.InvestmentView{}
			for _, inv := range v.Pool.Investments {
				if inv.InvestorID == u.ID {
					own = append(own, inv)
				}
			}
			if len(own) == 0 {
				continue
			}

			v.Pool.Investments = own

			if d := v.Pool.Distribution; d != nil {
				pays := []models.Payout{}

				for _, pay := range d.Payouts {
					if pay.InvestorID == u.ID {
						pays = append(pays, pay)
					}
				}

				v.Pool.Distribution = &models.Distribution{ID: d.ID, Profit: d.Profit, Payouts: pays}
			}

			selected := []models.Solution{}
			for _, sol := range v.Solutions {
				if p.Match != nil && sol.ID == p.Match.SolutionID {
					selected = append(selected, sol)
				}
			}

			v.Solutions = selected

		default:
			continue
		}

		visible = append(visible, v)
	}

	return visible, opps, nil
}
