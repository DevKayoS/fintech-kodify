package models

type CreateInvestmentRequest struct {
	TypeSlug    string  `json:"type_slug" binding:"required"`
	Amount      float64 `json:"amount" binding:"required,gt=0"` // em reais
	Description string  `json:"description"`
	InvestedAt  string  `json:"invested_at"` // RFC3339 ou YYYY-MM-DD, opcional (default: now)
}

type InvestmentResponse struct {
	ID          int64   `json:"id"`
	TypeSlug    string  `json:"type_slug"`
	TypeName    string  `json:"type_name"`
	Amount      float64 `json:"amount"` // em reais
	Description string  `json:"description,omitempty"`
	InvestedAt  string  `json:"invested_at"`
	CreatedAt   string  `json:"created_at"`
}

type ListInvestmentsQuery struct {
	Month string `form:"month"` // YYYY-MM, opcional
	Type  string `form:"type"`  // slug, opcional
}
