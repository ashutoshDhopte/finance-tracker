package models

type ReportSummary struct {
	TotalIncome    float64           `json:"total_income"`
	TotalExpenses  float64           `json:"total_expenses"`
	TotalTransfers float64           `json:"total_transfers"`
	Net            float64           `json:"net"`
	ByCategory     []CategorySummary `json:"by_category"`
	ByAccount      []AccountSummary  `json:"by_account"`
}

type AccountSummary struct {
	AccountID   string  `json:"account_id"`
	AccountName string  `json:"account_name"`
	Institution string  `json:"institution"`
	AccountType string  `json:"account_type"`
	LastFour    *string `json:"last_four,omitempty"`
	Income      float64 `json:"income"`
	Expenses    float64 `json:"expenses"`
	Net         float64 `json:"net"`
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
