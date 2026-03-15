package expense

import (
	"context"
	"fmt"
	"time"

	"github.com/DevKayoS/fintech-kodify/internal/infrastructure/pgstore"
	"github.com/DevKayoS/fintech-kodify/internal/utils"
	"github.com/jackc/pgx/v5/pgtype"
)

type CreateExpenseResult struct {
	ID           int64
	Amount       float64
	CategoryName string
	Description  string
	OccurredAt   time.Time
}

func (uc *ExpenseUseCase) CreateFromTelegram(ctx context.Context, chatID int64, amountReais float64, categorySlug string, description string) (*CreateExpenseResult, error) {
	user, err := uc.repository.GetUserByTelegramChatID(ctx, pgtype.Int8{Int64: chatID, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("conta não vinculada. Use /start <token> para vincular sua conta")
	}

	category, err := uc.repository.GetExpenseCategoryBySlug(ctx, categorySlug)
	if err != nil {
		return nil, fmt.Errorf("categoria '%s' não encontrada. Use /categorias para ver as disponíveis", categorySlug)
	}

	now := time.Now().UTC()
	expense, err := uc.repository.InsertExpense(ctx, pgstore.InsertExpenseParams{
		UserID:      user.ID,
		CategoryID:  pgtype.Int4{Int32: category.ID, Valid: true},
		Amount:      utils.ToCentavos(amountReais),
		Description: pgtype.Text{String: description, Valid: description != ""},
		OccurredAt:  pgtype.Timestamptz{Time: now, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("erro ao registrar gasto, tente novamente")
	}

	return &CreateExpenseResult{
		ID:           expense.ID,
		Amount:       amountReais,
		CategoryName: category.Name,
		Description:  description,
		OccurredAt:   now,
	}, nil
}
