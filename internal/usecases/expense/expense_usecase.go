package expense

import (
	"context"

	"github.com/DevKayoS/fintech-kodify/internal/infrastructure/pgstore"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ExpenseRepository interface {
	GetUserByTelegramChatID(ctx context.Context, telegramChatID pgtype.Int8) (pgstore.User, error)
	GetExpenseCategoryBySlug(ctx context.Context, slug string) (pgstore.ExpenseCategory, error)
	InsertExpense(ctx context.Context, arg pgstore.InsertExpenseParams) (pgstore.Expense, error)
	ListExpenseCategories(ctx context.Context) ([]pgstore.ExpenseCategory, error)
	GetExpenseByID(ctx context.Context, arg pgstore.GetExpenseByIDParams) (pgstore.GetExpenseByIDRow, error)
	DeleteExpense(ctx context.Context, arg pgstore.DeleteExpenseParams) error
	ListExpensesByUser(ctx context.Context, userID int64) ([]pgstore.ListExpensesByUserRow, error)
	ListExpensesByUserAndPeriod(ctx context.Context, arg pgstore.ListExpensesByUserAndPeriodParams) ([]pgstore.ListExpensesByUserAndPeriodRow, error)
}

type ExpenseUseCase struct {
	repository ExpenseRepository
}

func NewExpenseUseCase(pool *pgxpool.Pool) *ExpenseUseCase {
	return &ExpenseUseCase{
		repository: pgstore.New(pool),
	}
}
