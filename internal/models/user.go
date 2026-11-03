package models

type User struct {
	ID   string `json:"id"`
	Role string `json:"role"`
	Name string `json:"name"`
}

// Users seeder
var Users = []User{
	{"region-tomsk", "region", "Томская область"},
	{"region-other", "region", "Другой регион"},
	{"tech-water", "technologist", "Технолог очистки воды"},
	{"tech-other", "technologist", "Другой технолог"},
	{"expert-1", "expert", "Эксперт"},
	{"investor-a", "investor", "Инвестор A"},
	{"investor-b", "investor", "Инвестор B"},
	{"investor-c", "investor", "Инвестор C"},
}
