package services

import (
	"errors"
	"math/big"
	"sort"

	"github.com/phosphoriat/invest-pool-go/internal/models"
	"github.com/phosphoriat/invest-pool-go/internal/validation"
)

const MaxMoney int64 = validation.MaxMoney // below JS Number.MAX_SAFE_INTEGER

// Allocate Distribute profit using largest remainders without floating-point arithmetic.
// Equal remainders are resolved by lexicographic investor ID.
func Allocate(investments []models.Investment, profit int64) ([]models.Payout, error) {
	if profit <= 0 || profit > MaxMoney {
		return nil, errors.New("invalid profit")
	}

	shares := map[string]int64{}
	var total int64
	for _, i := range investments {
		if i.Amount <= 0 || i.Amount > MaxMoney-total || i.InvestorID == "" {
			return nil, errors.New("invalid shares")
		}
		total += i.Amount
		shares[i.InvestorID] += i.Amount
	}

	if total == 0 {
		return nil, errors.New("empty pool")
	}

	type part struct {
		models.Payout
		remainder *big.Int
	}

	parts := []part{}
	var paid int64
	for id, amount := range shares {
		n := new(big.Int).Mul(big.NewInt(profit), big.NewInt(amount))
		q, rem := new(big.Int), new(big.Int)
		q.QuoRem(n, big.NewInt(total), rem)

		parts = append(parts, part{models.Payout{InvestorID: id, Amount: q.Int64()}, rem})
		paid += q.Int64()
	}
	sort.Slice(parts, func(i, j int) bool {
		cmp := parts[i].remainder.Cmp(parts[j].remainder)

		if cmp == 0 {
			return parts[i].InvestorID < parts[j].InvestorID
		}

		return cmp > 0
	})

	for i := int64(0); i < profit-paid; i++ {
		parts[i].Amount++
	}
	sort.Slice(parts, func(i, j int) bool {
		return parts[i].InvestorID < parts[j].InvestorID
	})

	out := make([]models.Payout, len(parts))
	for i, p := range parts {
		out[i] = p.Payout
	}

	return out, nil
}
