package controllers

import (
	"net/http"

	"github.com/DevKayoS/fintech-kodify/internal/errors"
	"github.com/DevKayoS/fintech-kodify/internal/models"
	"github.com/DevKayoS/fintech-kodify/internal/usecases/user"
	"github.com/gin-gonic/gin"
)

type UserController struct {
	useCase *user.UserUseCase
}

func NewUserController(uc *user.UserUseCase) *UserController {
	return &UserController{useCase: uc}
}

// GetMe retorna o perfil do usuário autenticado.
// GET /api/v1/me
func (c *UserController) GetMe(ctx *gin.Context) {
	userID := ctx.GetInt64("user_id")

	result, err := c.useCase.GetMe(ctx.Request.Context(), userID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, models.UserProfileResponse{
		ID:        result.ID,
		Name:      result.Name,
		Email:     result.Email,
		Role:      result.Role,
		CreatedAt: result.CreatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

// GenerateTelegramLink gera um token de vinculação do Telegram para o usuário autenticado.
// POST /api/v1/telegram/link
func (c *UserController) GenerateTelegramLink(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.Error(errors.Unauthorized("user not found in context"))
		return
	}

	result, err := c.useCase.GenerateLinkToken(ctx.Request.Context(), userID.(int64))
	if err != nil {
		ctx.Error(errors.Internal("failed to generate link token", err))
		return
	}

	ctx.JSON(http.StatusOK, result)
}
