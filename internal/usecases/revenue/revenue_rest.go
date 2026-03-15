package revenue

import (
	"context"
	"time"

	"github.com/DevKayoS/fintech-kodify/internal/errors"
	"github.com/DevKayoS/fintech-kodify/internal/infrastructure/pgstore"
	"github.com/DevKayoS/fintech-kodify/internal/utils"
	"github.com/jackc/pgx/v5/pgtype"
)

type RevenueItem struct {
	ID          int64
	Amount      float64
	Description string
	ReceivedAt  time.Time
	CreatedAt   time.Time
}

func (uc *RevenueUseCase) Create(ctx context.Context, userID int64, amountReais float64, description string, receivedAt time.Time) (*CreateRevenueResult, error) {
	if receivedAt.IsZero() {
		receivedAt = time.Now().UTC()
	}

	rev, err := uc.repository.InsertRevenue(ctx, pgstore.InsertRevenueParams{
		UserID:      userID,
		Amount:      utils.ToCentavos(amountReais),
		Description: pgtype.Text{String: description, Valid: description != ""},
		ReceivedAt:  pgtype.Timestamptz{Time: receivedAt, Valid: true},
	})
	if err != nil {
		return nil, errors.Internal("erro ao registrar receita", err)
	}

	return &CreateRevenueResult{
		ID:          rev.ID,
		Amount:      amountReais,
		Description: description,
		ReceivedAt:  receivedAt,
	}, nil
}

func (uc *RevenueUseCase) List(ctx context.Context, userID int64, month string) ([]RevenueItem, error) {
	var items []RevenueItem

	if month != "" {
		start, end, err := utils.ParseMonthRange(month)
		if err != nil {
			return nil, errors.BadRequest("formato de mês inválido, use YYYY-MM")
		}

		rows, err := uc.repository.ListRevenuesByUserAndPeriod(ctx, pgstore.ListRevenuesByUserAndPeriodParams{
			UserID:       userID,
			ReceivedAt:   pgtype.Timestamptz{Time: start, Valid: true},
			ReceivedAt_2: pgtype.Timestamptz{Time: end, Valid: true},
		})
		if err != nil {
			return nil, errors.Internal("erro ao buscar receitas", err)
		}

		for _, r := range rows {
			items = append(items, RevenueItem{
				ID:          r.ID,
				Amount:      utils.ToReais(r.Amount),
				Description: r.Description.String,
				ReceivedAt:  r.ReceivedAt.Time,
				CreatedAt:   r.CreatedAt.Time,
			})
		}
	} else {
		rows, err := uc.repository.ListRevenuesByUser(ctx, userID)
		if err != nil {
			return nil, errors.Internal("erro ao buscar receitas", err)
		}

		for _, r := range rows {
			items = append(items, RevenueItem{
				ID:          r.ID,
				Amount:      utils.ToReais(r.Amount),
				Description: r.Description.String,
				ReceivedAt:  r.ReceivedAt.Time,
				CreatedAt:   r.CreatedAt.Time,
			})
		}
	}

	if items == nil {
		items = []RevenueItem{}
	}

	return items, nil
}
