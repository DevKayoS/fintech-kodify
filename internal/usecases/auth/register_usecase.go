package auth

import (
	"context"

	"github.com/DevKayoS/fintech-kodify/internal/errors"
	"github.com/DevKayoS/fintech-kodify/internal/infrastructure/pgstore"
	"github.com/DevKayoS/fintech-kodify/internal/utils"
	"github.com/jackc/pgx/v5/pgtype"
)

type RegisterResult struct {
	ID    int64
	Name  string
	Email string
}

func (uc *AuthUseCase) Register(ctx context.Context, name, email, password string) (*RegisterResult, error) {
	// Verifica se o email já existe — não vaza detalhes para o cliente
	_, err := uc.repository.GetUserByEmail(ctx, email)
	if err == nil {
		return nil, errors.Conflict("email já cadastrado")
	}

	hash, err := utils.HashPassword(password)
	if err != nil {
		return nil, errors.Internal("erro ao processar senha", err)
	}

	id, err := uc.repository.InsertUser(ctx, pgstore.InsertUserParams{
		Name:     name,
		Email:    email,
		Password: pgtype.Text{String: hash, Valid: true},
	})
	if err != nil {
		return nil, errors.Internal("erro ao cadastrar usuário", err)
	}

	return &RegisterResult{ID: id, Name: name, Email: email}, nil
}
