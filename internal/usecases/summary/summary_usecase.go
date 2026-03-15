package summary

import (
	"context"
	"fmt"
	"time"

	"github.com/DevKayoS/fintech-kodify/internal/infrastructure/pgstore"
	"github.com/DevKayoS/fintech-kodify/internal/utils"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SummaryRepository interface {
	GetUserByTelegramChatID(ctx context.Context, telegramChatID pgtype.Int8) (pgstore.User, error)
	GetExpenseSummaryByPeriod(ctx context.Context, arg pgstore.GetExpenseSummaryByPeriodParams) ([]pgstore.GetExpenseSummaryByPeriodRow, error)
	GetRevenueSummaryByPeriod(ctx context.Context, arg pgstore.GetRevenueSummaryByPeriodParams) (int64, error)
}

type SummaryUseCase struct {
	repository SummaryRepository
}

func NewSummaryUseCase(pool *pgxpool.Pool) *SummaryUseCase {
	return &SummaryUseCase{repository: pgstore.New(pool)}
}

type CategorySummary struct {
	Slug  string
	Name  string
	Total int64 // centavos
}

type MonthlySummary struct {
	Month              time.Time
	TotalExpenses      int64 // centavos
	TotalRevenues      int64 // centavos
	Balance            int64 // centavos — pode ser negativo
	ExpensesByCategory []CategorySummary
}

// GetMonthlySummary retorna o resumo financeiro do mês atual para o usuário
// identificado pelo chatID do Telegram.
func (uc *SummaryUseCase) GetMonthlySummary(ctx context.Context, chatID int64) (*MonthlySummary, error) {
	user, err := uc.repository.GetUserByTelegramChatID(ctx, pgtype.Int8{Int64: chatID, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("conta não vinculada. Use /start para se cadastrar")
	}

	start, end := utils.CurrentMonthRange()
	period := pgstore.GetExpenseSummaryByPeriodParams{
		UserID:       user.ID,
		OccurredAt:   pgtype.Timestamptz{Time: start, Valid: true},
		OccurredAt_2: pgtype.Timestamptz{Time: end, Valid: true},
	}

	expenseRows, err := uc.repository.GetExpenseSummaryByPeriod(ctx, period)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar gastos")
	}

	totalRevenues, err := uc.repository.GetRevenueSummaryByPeriod(ctx, pgstore.GetRevenueSummaryByPeriodParams{
		UserID:       user.ID,
		ReceivedAt:   pgtype.Timestamptz{Time: start, Valid: true},
		ReceivedAt_2: pgtype.Timestamptz{Time: end, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar receitas")
	}

	var totalExpenses int64
	categories := make([]CategorySummary, 0, len(expenseRows))
	for _, row := range expenseRows {
		totalExpenses += row.Total
		categories = append(categories, CategorySummary{
			Slug:  row.CategorySlug,
			Name:  row.CategoryName,
			Total: row.Total,
		})
	}

	return &MonthlySummary{
		Month:              start,
		TotalExpenses:      totalExpenses,
		TotalRevenues:      totalRevenues,
		Balance:            totalRevenues - totalExpenses,
		ExpensesByCategory: categories,
	}, nil
}
