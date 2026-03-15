package revenue

import (
	"context"
	"fmt"
	"time"

	"github.com/DevKayoS/fintech-kodify/internal/infrastructure/pgstore"
	"github.com/DevKayoS/fintech-kodify/internal/utils"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RevenueRepository interface {
	GetUserByTelegramChatID(ctx context.Context, telegramChatID pgtype.Int8) (pgstore.User, error)
	InsertRevenue(ctx context.Context, arg pgstore.InsertRevenueParams) (pgstore.Revenue, error)
}

type RevenueUseCase struct {
	repository RevenueRepository
}

func NewRevenueUseCase(pool *pgxpool.Pool) *RevenueUseCase {
	return &RevenueUseCase{
		repository: pgstore.New(pool),
	}
}

type CreateRevenueResult struct {
	ID          int64
	Amount      float64
	Description string
	ReceivedAt  time.Time
}

func (uc *RevenueUseCase) CreateFromTelegram(ctx context.Context, chatID int64, amountReais float64, description string) (*CreateRevenueResult, error) {
	u, err := uc.repository.GetUserByTelegramChatID(ctx, pgtype.Int8{Int64: chatID, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("conta não vinculada. Use /start para se cadastrar")
	}

	now := time.Now().UTC()
	rev, err := uc.repository.InsertRevenue(ctx, pgstore.InsertRevenueParams{
		UserID:      u.ID,
		Amount:      utils.ToCentavos(amountReais),
		Description: pgtype.Text{String: description, Valid: description != ""},
		ReceivedAt:  pgtype.Timestamptz{Time: now, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("erro ao registrar receita, tente novamente")
	}

	return &CreateRevenueResult{
		ID:          rev.ID,
		Amount:      amountReais,
		Description: description,
		ReceivedAt:  now,
	}, nil
}
