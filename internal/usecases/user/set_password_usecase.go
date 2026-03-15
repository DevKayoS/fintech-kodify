package user

import (
	"context"
	"fmt"

	"github.com/DevKayoS/fintech-kodify/internal/utils"
	"github.com/jackc/pgx/v5/pgtype"
)

// SetPasswordFromTelegram define (ou redefine) a senha do usuário identificado pelo chatID.
func (uc *UserUseCase) SetPasswordFromTelegram(ctx context.Context, chatID int64, password string) error {
	u, err := uc.repository.GetUserByTelegramChatID(ctx, pgtype.Int8{Int64: chatID, Valid: true})
	if err != nil {
		return fmt.Errorf("conta não encontrada")
	}

	hash, err := utils.HashPassword(password)
	if err != nil {
		return fmt.Errorf("erro ao processar senha")
	}

	return uc.repository.UpdateUserPassword(ctx, u.ID, hash)
}
