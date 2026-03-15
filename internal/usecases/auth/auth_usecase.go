package auth

import (
	"context"

	"github.com/DevKayoS/fintech-kodify/internal/infrastructure/pgstore"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthRepository interface {
	GetUserByEmail(ctx context.Context, email string) (pgstore.User, error)
	InsertUser(ctx context.Context, arg pgstore.InsertUserParams) (int64, error)
	GetUserWithRole(ctx context.Context, email string) (pgstore.GetUserWithRoleRow, error)
	GetUserPermissions(ctx context.Context, email string) ([]string, error)
	GetUserByID(ctx context.Context, id int64) (pgstore.User, error)
}

type AuthUseCase struct {
	repository AuthRepository
}

func NewAuthUseCase(pool *pgxpool.Pool) *AuthUseCase {
	return &AuthUseCase{repository: pgstore.New(pool)}
}
