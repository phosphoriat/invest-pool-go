package models

type Pool struct {
	ID           string        `json:"id"`
	Target       int64         `json:"target"`
	Total        int64         `json:"total"`
	Status       string        `json:"status"`
	Investments  []Investment  `json:"investments"`
	Distribution *Distribution `json:"distribution"`
}

type Opportunity struct {
	ProblemID string `json:"problem_id"`
	Title     string `json:"title"`
	PoolID    string `json:"pool_id"`
	Target    int64  `json:"target"`
	Total     int64  `json:"total"`
}
