package models

type CategorySummaryResponse struct {
	Slug  string  `json:"slug"`
	Name  string  `json:"name"`
	Total float64 `json:"total"` // em reais
}

type InvestmentTypeSummaryResponse struct {
	Slug  string  `json:"slug"`
	Name  string  `json:"name"`
	Total float64 `json:"total"` // em reais
}

type SummaryResponse struct {
	Month              string                          `json:"month"` // YYYY-MM
	TotalExpenses      float64                         `json:"total_expenses"`
	TotalRevenues      float64                         `json:"total_revenues"`
	Balance            float64                         `json:"balance"`
	TotalInvestments   float64                         `json:"total_investments"`
	AllTimeInvestments float64                         `json:"all_time_investments"`
	ExpensesByCategory []CategorySummaryResponse       `json:"expenses_by_category"`
	InvestmentsByType  []InvestmentTypeSummaryResponse `json:"investments_by_type"`
}

type StatementEntryResponse struct {
	Kind        string  `json:"kind"` // expense, revenue, investment
	Date        string  `json:"date"`
	Amount      float64 `json:"amount"`
	Label       string  `json:"label,omitempty"`
	Description string  `json:"description,omitempty"`
}

type StatementResponse struct {
	Month   string                   `json:"month"`
	Entries []StatementEntryResponse `json:"entries"`
}
