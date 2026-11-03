package models

// ProblemView Role-filtered read model of a Problem returned by GET /api/state.
// It is built from the aggregate and never carries operation keys.
type ProblemView struct {
	ID        string     `json:"id"`
	RegionID  string     `json:"region_id"`
	Title     string     `json:"title"`
	Solutions []Solution `json:"solutions"`
	// Matched reports that an expert has confirmed some solution, even when Match itself is hidden
	Matched bool      `json:"matched"`
	Match   *Match    `json:"match"`
	Pool    *PoolView `json:"pool"`
}

type PoolView struct {
	ID           string           `json:"id"`
	Target       int64            `json:"target"`
	Total        int64            `json:"total"`
	Status       string           `json:"status"`
	Investments  []InvestmentView `json:"investments"`
	Distribution *Distribution    `json:"distribution"`
}

type InvestmentView struct {
	ID         string `json:"id"`
	InvestorID string `json:"investor_id"`
	Amount     int64  `json:"amount"`
}

// NewProblemView Copy an aggregate into its read model without operation keys
func NewProblemView(p Problem) ProblemView {
	v := ProblemView{
		ID:        p.ID,
		RegionID:  p.RegionID,
		Title:     p.Title,
		Solutions: p.Solutions,
		Matched:   p.Match != nil,
		Match:     p.Match,
	}
	if v.Solutions == nil {
		v.Solutions = []Solution{}
	}

	if p.Pool != nil {
		v.Pool = &PoolView{
			ID:           p.Pool.ID,
			Target:       p.Pool.Target,
			Total:        p.Pool.Total,
			Status:       p.Pool.Status,
			Investments:  make([]InvestmentView, 0, len(p.Pool.Investments)),
			Distribution: p.Pool.Distribution,
		}
		for _, i := range p.Pool.Investments {
			v.Pool.Investments = append(v.Pool.Investments, InvestmentView{ID: i.ID, InvestorID: i.InvestorID, Amount: i.Amount})
		}
	}

	return v
}
