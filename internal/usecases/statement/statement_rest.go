package statement

import (
	"context"
	"time"

	"github.com/DevKayoS/fintech-kodify/internal/errors"
	"github.com/DevKayoS/fintech-kodify/internal/infrastructure/pgstore"
	"github.com/DevKayoS/fintech-kodify/internal/utils"
	"github.com/jackc/pgx/v5/pgtype"
)

// GetForUser retorna o extrato do mês para o usuário identificado pelo userID (JWT).
func (uc *StatementUseCase) GetForUser(ctx context.Context, userID int64, monthStart time.Time) (*StatementResult, error) {
	var start, end time.Time
	if monthStart.IsZero() {
		start, end = utils.CurrentMonthRange()
	} else {
		start = monthStart
		end = start.AddDate(0, 1, 0)
	}

	startTS := pgtype.Timestamptz{Time: start, Valid: true}
	endTS := pgtype.Timestamptz{Time: end, Valid: true}

	expenses, err := uc.repository.ListExpensesByUserAndPeriod(ctx, pgstore.ListExpensesByUserAndPeriodParams{
		UserID:       userID,
		OccurredAt:   startTS,
		OccurredAt_2: endTS,
	})
	if err != nil {
		return nil, errors.Internal("erro ao buscar gastos", err)
	}

	revenues, err := uc.repository.ListRevenuesByUserAndPeriod(ctx, pgstore.ListRevenuesByUserAndPeriodParams{
		UserID:       userID,
		ReceivedAt:   startTS,
		ReceivedAt_2: endTS,
	})
	if err != nil {
		return nil, errors.Internal("erro ao buscar receitas", err)
	}

	investments, err := uc.repository.ListInvestmentsByUserAndPeriod(ctx, pgstore.ListInvestmentsByUserAndPeriodParams{
		UserID:       userID,
		InvestedAt:   startTS,
		InvestedAt_2: endTS,
	})
	if err != nil {
		return nil, errors.Internal("erro ao buscar investimentos", err)
	}

	entries := make([]StatementEntry, 0, len(expenses)+len(revenues)+len(investments))

	for _, e := range expenses {
		entries = append(entries, StatementEntry{
			Kind:        "expense",
			Date:        e.OccurredAt.Time,
			Amount:      utils.ToReais(e.Amount),
			Label:       e.CategoryName.String,
			Description: e.Description.String,
		})
	}

	for _, r := range revenues {
		entries = append(entries, StatementEntry{
			Kind:        "revenue",
			Date:        r.ReceivedAt.Time,
			Amount:      utils.ToReais(r.Amount),
			Description: r.Description.String,
		})
	}

	for _, i := range investments {
		entries = append(entries, StatementEntry{
			Kind:        "investment",
			Date:        i.InvestedAt.Time,
			Amount:      utils.ToReais(i.Amount),
			Label:       i.TypeName,
			Description: i.Description.String,
		})
	}

	sortEntriesByDate(entries)

	return &StatementResult{
		Month:   start,
		Entries: entries,
	}, nil
}
