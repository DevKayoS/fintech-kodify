package user

import (
	"context"

	"github.com/DevKayoS/fintech-kodify/internal/infrastructure/pgstore"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	InsertLinkToken(ctx context.Context, arg pgstore.InsertLinkTokenParams) (int64, error)
	GetValidLinkToken(ctx context.Context, token string) (pgstore.TelegramLinkToken, error)
	MarkLinkTokenUsed(ctx context.Context, id int64) error
	UpdateUserTelegramChatID(ctx context.Context, arg pgstore.UpdateUserTelegramChatIDParams) error
}

type UserUseCase struct {
	repository UserRepository
}

func NewUserUseCase(pool *pgxpool.Pool) *UserUseCase {
	return &UserUseCase{
		repository: pgstore.New(pool),
	}
}
