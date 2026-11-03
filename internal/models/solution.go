package models

type Solution struct {
	ID             string `json:"id"`
	TechnologistID string `json:"technologist_id"`
	Description    string `json:"description"`
}
