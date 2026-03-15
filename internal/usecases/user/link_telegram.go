package user

import (
	"context"
	"fmt"

	"github.com/DevKayoS/fintech-kodify/internal/infrastructure/pgstore"
	"github.com/jackc/pgx/v5/pgtype"
)

func (uc *UserUseCase) LinkTelegramChatID(ctx context.Context, token string, chatID int64) error {
	linkToken, err := uc.repository.GetValidLinkToken(ctx, token)
	if err != nil {
		return fmt.Errorf("token inválido ou expirado")
	}

	if err := uc.repository.UpdateUserTelegramChatID(ctx, pgstore.UpdateUserTelegramChatIDParams{
		ID:             linkToken.UserID,
		TelegramChatID: pgtype.Int8{Int64: chatID, Valid: true},
	}); err != nil {
		return fmt.Errorf("erro ao vincular conta")
	}

	if err := uc.repository.MarkLinkTokenUsed(ctx, linkToken.ID); err != nil {
		return fmt.Errorf("erro ao finalizar vinculação")
	}

	return nil
}
