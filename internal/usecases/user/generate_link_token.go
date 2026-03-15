package user

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/DevKayoS/fintech-kodify/internal/infrastructure/pgstore"
	"github.com/jackc/pgx/v5/pgtype"
)

type GenerateLinkTokenResult struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

func (uc *UserUseCase) GenerateLinkToken(ctx context.Context, userID int64) (*GenerateLinkTokenResult, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("erro ao gerar token")
	}
	token := hex.EncodeToString(b)
	expiresAt := time.Now().UTC().Add(15 * time.Minute)

	_, err := uc.repository.InsertLinkToken(ctx, pgstore.InsertLinkTokenParams{
		UserID:    userID,
		Token:     token,
		ExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("erro ao salvar token de vinculação")
	}

	return &GenerateLinkTokenResult{
		Token:     token,
		ExpiresAt: expiresAt,
	}, nil
}
