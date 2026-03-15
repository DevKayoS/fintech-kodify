package investment

import (
	"context"
	"time"

	"github.com/DevKayoS/fintech-kodify/internal/errors"
	"github.com/DevKayoS/fintech-kodify/internal/infrastructure/pgstore"
	"github.com/DevKayoS/fintech-kodify/internal/utils"
	"github.com/jackc/pgx/v5/pgtype"
)

type InvestmentItem struct {
	ID          int64
	Amount      float64
	TypeSlug    string
	TypeName    string
	Description string
	InvestedAt  time.Time
	CreatedAt   time.Time
}

func (uc *InvestmentUseCase) Create(ctx context.Context, userID int64, amountReais float64, typeSlug, description string, investedAt time.Time) (*CreateInvestmentResult, error) {
	invType, err := uc.repository.GetInvestmentTypeBySlug(ctx, typeSlug)
	if err != nil {
		return nil, errors.NotFound("tipo de investimento '" + typeSlug + "' não encontrado")
	}

	if investedAt.IsZero() {
		investedAt = time.Now().UTC()
	}

	inv, err := uc.repository.InsertInvestment(ctx, pgstore.InsertInvestmentParams{
		UserID:           userID,
		InvestmentTypeID: invType.ID,
		Amount:           utils.ToCentavos(amountReais),
		Description:      pgtype.Text{String: description, Valid: description != ""},
		InvestedAt:       pgtype.Timestamptz{Time: investedAt, Valid: true},
	})
	if err != nil {
		return nil, errors.Internal("erro ao registrar investimento", err)
	}

	return &CreateInvestmentResult{
		ID:          inv.ID,
		Amount:      amountReais,
		TypeName:    invType.Name,
		Description: description,
		InvestedAt:  investedAt,
	}, nil
}

func (uc *InvestmentUseCase) GetByID(ctx context.Context, userID, id int64) (*InvestmentItem, error) {
	row, err := uc.repository.GetInvestmentByID(ctx, pgstore.GetInvestmentByIDParams{ID: id, UserID: userID})
	if err != nil {
		return nil, errors.NotFound("investimento não encontrado")
	}

	return &InvestmentItem{
		ID:          row.ID,
		Amount:      utils.ToReais(row.Amount),
		TypeSlug:    row.TypeSlug,
		TypeName:    row.TypeName,
		Description: row.Description.String,
		InvestedAt:  row.InvestedAt.Time,
		CreatedAt:   row.CreatedAt.Time,
	}, nil
}

func (uc *InvestmentUseCase) Delete(ctx context.Context, userID, id int64) error {
	_, err := uc.repository.GetInvestmentByID(ctx, pgstore.GetInvestmentByIDParams{ID: id, UserID: userID})
	if err != nil {
		return errors.NotFound("investimento não encontrado")
	}

	return uc.repository.DeleteInvestment(ctx, pgstore.DeleteInvestmentParams{ID: id, UserID: userID})
}

func (uc *InvestmentUseCase) List(ctx context.Context, userID int64, month, typeSlug string) ([]InvestmentItem, error) {
	var items []InvestmentItem

	if month != "" {
		start, end, err := utils.ParseMonthRange(month)
		if err != nil {
			return nil, errors.BadRequest("formato de mês inválido, use YYYY-MM")
		}

		rows, err := uc.repository.ListInvestmentsByUserAndPeriod(ctx, pgstore.ListInvestmentsByUserAndPeriodParams{
			UserID:       userID,
			InvestedAt:   pgtype.Timestamptz{Time: start, Valid: true},
			InvestedAt_2: pgtype.Timestamptz{Time: end, Valid: true},
		})
		if err != nil {
			return nil, errors.Internal("erro ao buscar investimentos", err)
		}

		for _, r := range rows {
			items = append(items, InvestmentItem{
				ID:          r.ID,
				Amount:      utils.ToReais(r.Amount),
				TypeSlug:    r.TypeSlug,
				TypeName:    r.TypeName,
				Description: r.Description.String,
				InvestedAt:  r.InvestedAt.Time,
				CreatedAt:   r.CreatedAt.Time,
			})
		}
	} else {
		rows, err := uc.repository.ListInvestmentsByUser(ctx, userID)
		if err != nil {
			return nil, errors.Internal("erro ao buscar investimentos", err)
		}

		for _, r := range rows {
			items = append(items, InvestmentItem{
				ID:          r.ID,
				Amount:      utils.ToReais(r.Amount),
				TypeSlug:    r.TypeSlug,
				TypeName:    r.TypeName,
				Description: r.Description.String,
				InvestedAt:  r.InvestedAt.Time,
				CreatedAt:   r.CreatedAt.Time,
			})
		}
	}

	// Filtro por tipo em memória
	if typeSlug != "" {
		filtered := make([]InvestmentItem, 0, len(items))
		for _, item := range items {
			if item.TypeSlug == typeSlug {
				filtered = append(filtered, item)
			}
		}
		items = filtered
	}

	if items == nil {
		items = []InvestmentItem{}
	}

	return items, nil
}
