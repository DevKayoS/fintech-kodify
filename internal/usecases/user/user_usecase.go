package user

import (
	"context"

	"github.com/DevKayoS/fintech-kodify/internal/infrastructure/pgstore"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	// link token (REST API flow)
	InsertLinkToken(ctx context.Context, arg pgstore.InsertLinkTokenParams) (int64, error)
	GetValidLinkToken(ctx context.Context, token string) (pgstore.TelegramLinkToken, error)
	MarkLinkTokenUsed(ctx context.Context, id int64) error
	UpdateUserTelegramChatID(ctx context.Context, arg pgstore.UpdateUserTelegramChatIDParams) error

	// user lookup
	GetUserByTelegramChatID(ctx context.Context, telegramChatID pgtype.Int8) (pgstore.User, error)

	// telegram onboarding
	InsertUserFromTelegram(ctx context.Context, arg pgstore.InsertUserFromTelegramParams) (int64, error)
	GetConversationState(ctx context.Context, chatID int64) (pgstore.TelegramConversation, error)
	UpsertConversationState(ctx context.Context, arg pgstore.UpsertConversationStateParams) error
	DeleteConversationState(ctx context.Context, chatID int64) error
}

type UserUseCase struct {
	repository UserRepository
}

func NewUserUseCase(pool *pgxpool.Pool) *UserUseCase {
	return &UserUseCase{
		repository: pgstore.New(pool),
	}
}
