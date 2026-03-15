package auth

import (
	"context"
	"time"

	"github.com/DevKayoS/fintech-kodify/internal/errors"
	"github.com/DevKayoS/fintech-kodify/internal/models"
	"github.com/DevKayoS/fintech-kodify/internal/utils"
	"github.com/golang-jwt/jwt"
)

type LoginResult struct {
	Token     string
	ExpiresAt time.Time
}

func (uc *AuthUseCase) Login(ctx context.Context, email, password string) (*LoginResult, error) {
	u, err := uc.repository.GetUserWithRole(ctx, email)
	if err != nil {
		// Mensagem genérica para não vazar se o email existe
		return nil, errors.Unauthorized("credenciais inválidas")
	}

	if !u.Password.Valid {
		return nil, errors.Unauthorized("credenciais inválidas")
	}

	if !utils.CheckPasswordHash(password, u.Password.String) {
		return nil, errors.Unauthorized("credenciais inválidas")
	}

	permissions, err := uc.repository.GetUserPermissions(ctx, email)
	if err != nil {
		return nil, errors.Internal("erro ao buscar permissões", err)
	}

	expiresAt := time.Now().UTC().Add(24 * time.Hour)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":     u.ID,
		"email":       u.Email,
		"role":        u.RoleName.String,
		"permissions": permissions,
		"exp":         expiresAt.Unix(),
	})

	tokenString, err := token.SignedString(models.SecretKey)
	if err != nil {
		return nil, errors.Internal("erro ao gerar token", err)
	}

	return &LoginResult{Token: tokenString, ExpiresAt: expiresAt}, nil
}
