package statement

import (
	"context"
	"fmt"
	"time"

	"github.com/DevKayoS/fintech-kodify/internal/infrastructure/pgstore"
	"github.com/DevKayoS/fintech-kodify/internal/utils"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StatementRepository interface {
	GetUserByTelegramChatID(ctx context.Context, telegramChatID pgtype.Int8) (pgstore.User, error)
	ListExpensesByUserAndPeriod(ctx context.Context, arg pgstore.ListExpensesByUserAndPeriodParams) ([]pgstore.ListExpensesByUserAndPeriodRow, error)
	ListRevenuesByUserAndPeriod(ctx context.Context, arg pgstore.ListRevenuesByUserAndPeriodParams) ([]pgstore.Revenue, error)
	ListInvestmentsByUserAndPeriod(ctx context.Context, arg pgstore.ListInvestmentsByUserAndPeriodParams) ([]pgstore.ListInvestmentsByUserAndPeriodRow, error)
}

type StatementUseCase struct {
	repository StatementRepository
}

func NewStatementUseCase(pool *pgxpool.Pool) *StatementUseCase {
	return &StatementUseCase{repository: pgstore.New(pool)}
}

type StatementEntry struct {
	Kind        string // "expense", "revenue", "investment"
	Date        time.Time
	Amount      float64
	Label       string // categoria ou tipo
	Description string
}

type StatementResult struct {
	Month   time.Time
	Entries []StatementEntry
}

func (uc *StatementUseCase) GetStatement(ctx context.Context, chatID int64, monthStart time.Time) (*StatementResult, error) {
	user, err := uc.repository.GetUserByTelegramChatID(ctx, pgtype.Int8{Int64: chatID, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("conta não vinculada. Use /start para se cadastrar")
	}

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
		UserID:       user.ID,
		OccurredAt:   startTS,
		OccurredAt_2: endTS,
	})
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar gastos")
	}

	revenues, err := uc.repository.ListRevenuesByUserAndPeriod(ctx, pgstore.ListRevenuesByUserAndPeriodParams{
		UserID:       user.ID,
		ReceivedAt:   startTS,
		ReceivedAt_2: endTS,
	})
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar receitas")
	}

	investments, err := uc.repository.ListInvestmentsByUserAndPeriod(ctx, pgstore.ListInvestmentsByUserAndPeriodParams{
		UserID:       user.ID,
		InvestedAt:   startTS,
		InvestedAt_2: endTS,
	})
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar investimentos")
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

	// ordenar por data decrescente
	sortEntriesByDate(entries)

	return &StatementResult{
		Month:   start,
		Entries: entries,
	}, nil
}

func sortEntriesByDate(entries []StatementEntry) {
	for i := 1; i < len(entries); i++ {
		for j := i; j > 0 && entries[j].Date.After(entries[j-1].Date); j-- {
			entries[j], entries[j-1] = entries[j-1], entries[j]
		}
	}
}
