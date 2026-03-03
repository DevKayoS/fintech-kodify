package models

type CreateExpenseRequest struct {
	CategorySlug string  `json:"category_slug"`
	Amount       float64 `json:"amount" binding:"required,gt=0"` // em reais
	Description  string  `json:"description"`
	OccurredAt   string  `json:"occurred_at"` // RFC3339 ou YYYY-MM-DD, opcional (default: now)
}

type ExpenseResponse struct {
	ID           int64   `json:"id"`
	CategorySlug string  `json:"category_slug,omitempty"`
	CategoryName string  `json:"category_name,omitempty"`
	Amount       float64 `json:"amount"` // em reais
	Description  string  `json:"description,omitempty"`
	OccurredAt   string  `json:"occurred_at"`
	CreatedAt    string  `json:"created_at"`
}

type ListExpensesQuery struct {
	Month    string `form:"month"`    // YYYY-MM, opcional
	Category string `form:"category"` // slug, opcional
}
