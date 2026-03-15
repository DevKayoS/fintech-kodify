package models

type CreateRevenueRequest struct {
	Amount      float64 `json:"amount" binding:"required,gt=0"` // em reais
	Description string  `json:"description"`
	ReceivedAt  string  `json:"received_at"` // RFC3339 ou YYYY-MM-DD, opcional (default: now)
}

type RevenueResponse struct {
	ID          int64   `json:"id"`
	Amount      float64 `json:"amount"` // em reais
	Description string  `json:"description,omitempty"`
	ReceivedAt  string  `json:"received_at"`
	CreatedAt   string  `json:"created_at"`
}

type ListRevenuesQuery struct {
	Month string `form:"month"` // YYYY-MM, opcional
}
