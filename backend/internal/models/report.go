package models

type ReportSummary struct {
	TotalIncome   float64           `json:"total_income"`
	TotalExpenses float64           `json:"total_expenses"`
	Net           float64           `json:"net"`
	ByCategory    []CategorySummary `json:"by_category"`
}

type CategorySummary struct {
	CategoryID   string  `json:"category_id"`
	CategoryName string  `json:"category_name"`
	Color        string  `json:"color"`
	Icon         string  `json:"icon"`
	Total        float64 `json:"total"`
	Count        int     `json:"count"`
}

type TrendPoint struct {
	Month         string  `json:"month"`
	TotalIncome   float64 `json:"total_income"`
	TotalExpenses float64 `json:"total_expenses"`
	Net           float64 `json:"net"`
}
