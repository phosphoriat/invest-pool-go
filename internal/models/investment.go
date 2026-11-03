package models

type Investment struct {
	ID         string `json:"id"`
	InvestorID string `json:"investor_id"`
	Amount     int64  `json:"amount"`
	Key        string `json:"key"`
}
