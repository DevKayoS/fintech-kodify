package expense

import (
	"context"
	"time"

	"github.com/DevKayoS/fintech-kodify/internal/errors"
	"github.com/DevKayoS/fintech-kodify/internal/infrastructure/pgstore"
	"github.com/DevKayoS/fintech-kodify/internal/utils"
	"github.com/jackc/pgx/v5/pgtype"
)

type ExpenseItem struct {
	ID           int64
	Amount       float64
	CategorySlug string
	CategoryName string
	Description  string
	OccurredAt   time.Time
	CreatedAt    time.Time
}

func (uc *ExpenseUseCase) Create(ctx context.Context, userID int64, amountReais float64, categorySlug, description string, occurredAt time.Time) (*CreateExpenseResult, error) {
	if categorySlug == "" {
		categorySlug = "outros"
	}

	category, err := uc.repository.GetExpenseCategoryBySlug(ctx, categorySlug)
	if err != nil {
		return nil, errors.NotFound("categoria '" + categorySlug + "' não encontrada")
	}

	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}

	exp, err := uc.repository.InsertExpense(ctx, pgstore.InsertExpenseParams{
		UserID:      userID,
		CategoryID:  pgtype.Int4{Int32: category.ID, Valid: true},
		Amount:      utils.ToCentavos(amountReais),
		Description: pgtype.Text{String: description, Valid: description != ""},
		OccurredAt:  pgtype.Timestamptz{Time: occurredAt, Valid: true},
	})
	if err != nil {
		return nil, errors.Internal("erro ao registrar gasto", err)
	}

	return &CreateExpenseResult{
		ID:           exp.ID,
		Amount:       amountReais,
		CategoryName: category.Name,
		Description:  description,
		OccurredAt:   occurredAt,
	}, nil
}

func (uc *ExpenseUseCase) GetByID(ctx context.Context, userID, id int64) (*ExpenseItem, error) {
	row, err := uc.repository.GetExpenseByID(ctx, pgstore.GetExpenseByIDParams{ID: id, UserID: userID})
	if err != nil {
		return nil, errors.NotFound("gasto não encontrado")
	}

	return &ExpenseItem{
		ID:           row.ID,
		Amount:       utils.ToReais(row.Amount),
		CategorySlug: row.CategorySlug.String,
		CategoryName: row.CategoryName.String,
		Description:  row.Description.String,
		OccurredAt:   row.OccurredAt.Time,
		CreatedAt:    row.CreatedAt.Time,
	}, nil
}

func (uc *ExpenseUseCase) Delete(ctx context.Context, userID, id int64) error {
	// Verifica existência e ownership antes de deletar
	_, err := uc.repository.GetExpenseByID(ctx, pgstore.GetExpenseByIDParams{ID: id, UserID: userID})
	if err != nil {
		return errors.NotFound("gasto não encontrado")
	}

	return uc.repository.DeleteExpense(ctx, pgstore.DeleteExpenseParams{ID: id, UserID: userID})
}

func (uc *ExpenseUseCase) List(ctx context.Context, userID int64, month, categorySlug string) ([]ExpenseItem, error) {
	var items []ExpenseItem

	if month != "" {
		start, end, err := utils.ParseMonthRange(month)
		if err != nil {
			return nil, errors.BadRequest("formato de mês inválido, use YYYY-MM")
		}

		rows, err := uc.repository.ListExpensesByUserAndPeriod(ctx, pgstore.ListExpensesByUserAndPeriodParams{
			UserID:       userID,
			OccurredAt:   pgtype.Timestamptz{Time: start, Valid: true},
			OccurredAt_2: pgtype.Timestamptz{Time: end, Valid: true},
		})
		if err != nil {
			return nil, errors.Internal("erro ao buscar gastos", err)
		}

		for _, r := range rows {
			items = append(items, ExpenseItem{
				ID:           r.ID,
				Amount:       utils.ToReais(r.Amount),
				CategorySlug: r.CategorySlug.String,
				CategoryName: r.CategoryName.String,
				Description:  r.Description.String,
				OccurredAt:   r.OccurredAt.Time,
				CreatedAt:    r.CreatedAt.Time,
			})
		}
	} else {
		rows, err := uc.repository.ListExpensesByUser(ctx, userID)
		if err != nil {
			return nil, errors.Internal("erro ao buscar gastos", err)
		}

		for _, r := range rows {
			items = append(items, ExpenseItem{
				ID:           r.ID,
				Amount:       utils.ToReais(r.Amount),
				CategorySlug: r.CategorySlug.String,
				CategoryName: r.CategoryName.String,
				Description:  r.Description.String,
				OccurredAt:   r.OccurredAt.Time,
				CreatedAt:    r.CreatedAt.Time,
			})
		}
	}

	// Filtro por categoria (em memória)
	if categorySlug != "" {
		filtered := make([]ExpenseItem, 0, len(items))
		for _, item := range items {
			if item.CategorySlug == categorySlug {
				filtered = append(filtered, item)
			}
		}
		items = filtered
	}

	if items == nil {
		items = []ExpenseItem{}
	}

	return items, nil
}
