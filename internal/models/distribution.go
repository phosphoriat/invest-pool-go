package models

type Distribution struct {
	ID      string   `json:"id"`
	Profit  int64    `json:"profit"`
	Payouts []Payout `json:"payouts"`
}

type Payout struct {
	InvestorID string `json:"investor_id"`
	Amount     int64  `json:"amount"`
}
