package user

import (
	"context"
	"time"

	"github.com/DevKayoS/fintech-kodify/internal/errors"
)

type MeResult struct {
	ID        int64
	Name      string
	Email     string
	Role      string
	CreatedAt time.Time
}

func (uc *UserUseCase) GetMe(ctx context.Context, userID int64) (*MeResult, error) {
	u, err := uc.repository.GetUserByID(ctx, userID)
	if err != nil {
		return nil, errors.NotFound("usuário não encontrado")
	}

	role := ""
	withRole, err := uc.repository.GetUserWithRole(ctx, u.Email)
	if err == nil {
		role = withRole.RoleName.String
	}

	return &MeResult{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Role:      role,
		CreatedAt: u.CreatedAt.Time,
	}, nil
}
