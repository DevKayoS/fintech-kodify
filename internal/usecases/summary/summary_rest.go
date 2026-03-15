package summary

import (
	"context"
	"time"

	"github.com/DevKayoS/fintech-kodify/internal/errors"
	"github.com/DevKayoS/fintech-kodify/internal/infrastructure/pgstore"
	"github.com/DevKayoS/fintech-kodify/internal/utils"
	"github.com/jackc/pgx/v5/pgtype"
)

// GetForUser retorna o resumo financeiro do mês para o usuário identificado pelo userID (JWT).
func (uc *SummaryUseCase) GetForUser(ctx context.Context, userID int64, monthStart time.Time) (*MonthlySummary, error) {
	var start, end time.Time
	if monthStart.IsZero() {
		start, end = utils.CurrentMonthRange()
	} else {
		start = monthStart
		end = start.AddDate(0, 1, 0)
	}

	expenseRows, err := uc.repository.GetExpenseSummaryByPeriod(ctx, pgstore.GetExpenseSummaryByPeriodParams{
		UserID:       userID,
		OccurredAt:   pgtype.Timestamptz{Time: start, Valid: true},
		OccurredAt_2: pgtype.Timestamptz{Time: end, Valid: true},
	})
	if err != nil {
		return nil, errors.Internal("erro ao buscar gastos", err)
	}

	totalRevenues, err := uc.repository.GetRevenueSummaryByPeriod(ctx, pgstore.GetRevenueSummaryByPeriodParams{
		UserID:       userID,
		ReceivedAt:   pgtype.Timestamptz{Time: start, Valid: true},
		ReceivedAt_2: pgtype.Timestamptz{Time: end, Valid: true},
	})
	if err != nil {
		return nil, errors.Internal("erro ao buscar receitas", err)
	}

	investmentRows, err := uc.repository.GetInvestmentSummaryByPeriod(ctx, pgstore.GetInvestmentSummaryByPeriodParams{
		UserID:       userID,
		InvestedAt:   pgtype.Timestamptz{Time: start, Valid: true},
		InvestedAt_2: pgtype.Timestamptz{Time: end, Valid: true},
	})
	if err != nil {
		return nil, errors.Internal("erro ao buscar investimentos do mês", err)
	}

	allTimeInvestments, err := uc.repository.GetTotalInvestmentsByUser(ctx, userID)
	if err != nil {
		return nil, errors.Internal("erro ao buscar total de investimentos", err)
	}

	var totalExpenses, totalInvestments int64
	categories := make([]CategorySummary, 0, len(expenseRows))
	for _, row := range expenseRows {
		totalExpenses += row.Total
		categories = append(categories, CategorySummary{
			Slug:  row.CategorySlug,
			Name:  row.CategoryName,
			Total: row.Total,
		})
	}

	investmentTypes := make([]InvestmentTypeSummary, 0, len(investmentRows))
	for _, row := range investmentRows {
		totalInvestments += row.Total
		investmentTypes = append(investmentTypes, InvestmentTypeSummary{
			Slug:  row.TypeSlug,
			Name:  row.TypeName,
			Total: row.Total,
		})
	}

	return &MonthlySummary{
		Month:              start,
		TotalExpenses:      totalExpenses,
		TotalRevenues:      totalRevenues,
		Balance:            totalRevenues - totalExpenses,
		TotalInvestments:   totalInvestments,
		AllTimeInvestments: allTimeInvestments,
		ExpensesByCategory: categories,
		InvestmentsByType:  investmentTypes,
	}, nil
}
