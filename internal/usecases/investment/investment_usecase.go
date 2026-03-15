package investment

import (
	"context"
	"fmt"
	"time"

	"github.com/DevKayoS/fintech-kodify/internal/infrastructure/pgstore"
	"github.com/DevKayoS/fintech-kodify/internal/utils"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type InvestmentRepository interface {
	GetUserByTelegramChatID(ctx context.Context, telegramChatID pgtype.Int8) (pgstore.User, error)
	GetInvestmentTypeBySlug(ctx context.Context, slug string) (pgstore.InvestmentType, error)
	InsertInvestment(ctx context.Context, arg pgstore.InsertInvestmentParams) (pgstore.Investment, error)
	ListInvestmentTypes(ctx context.Context) ([]pgstore.InvestmentType, error)
}

type InvestmentUseCase struct {
	repository InvestmentRepository
}

func NewInvestmentUseCase(pool *pgxpool.Pool) *InvestmentUseCase {
	return &InvestmentUseCase{repository: pgstore.New(pool)}
}

type CreateInvestmentResult struct {
	ID          int64
	Amount      float64
	TypeName    string
	Description string
	InvestedAt  time.Time
}

func (uc *InvestmentUseCase) CreateFromTelegram(ctx context.Context, chatID int64, amountReais float64, typeSlug string, description string) (*CreateInvestmentResult, error) {
	user, err := uc.repository.GetUserByTelegramChatID(ctx, pgtype.Int8{Int64: chatID, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("conta não vinculada. Use /start para se cadastrar")
	}

	invType, err := uc.repository.GetInvestmentTypeBySlug(ctx, typeSlug)
	if err != nil {
		return nil, fmt.Errorf("tipo '%s' não encontrado. Use /tipos_investimento para ver os disponíveis", typeSlug)
	}

	now := time.Now().UTC()
	inv, err := uc.repository.InsertInvestment(ctx, pgstore.InsertInvestmentParams{
		UserID:           user.ID,
		InvestmentTypeID: invType.ID,
		Amount:           utils.ToCentavos(amountReais),
		Description:      pgtype.Text{String: description, Valid: description != ""},
		InvestedAt:       pgtype.Timestamptz{Time: now, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("erro ao registrar investimento, tente novamente")
	}

	return &CreateInvestmentResult{
		ID:          inv.ID,
		Amount:      amountReais,
		TypeName:    invType.Name,
		Description: description,
		InvestedAt:  now,
	}, nil
}

func (uc *InvestmentUseCase) ListTypes(ctx context.Context) ([]pgstore.InvestmentType, error) {
	return uc.repository.ListInvestmentTypes(ctx)
}
