package controllers

import (
	"net/http"

	"github.com/DevKayoS/fintech-kodify/internal/errors"
	"github.com/DevKayoS/fintech-kodify/internal/models"
	"github.com/DevKayoS/fintech-kodify/internal/usecases/auth"
	"github.com/gin-gonic/gin"
)

type AuthController struct {
	useCase *auth.AuthUseCase
}

func NewAuthController(uc *auth.AuthUseCase) *AuthController {
	return &AuthController{useCase: uc}
}

// Register cria uma nova conta de usuário.
// POST /api/v1/auth/register
func (c *AuthController) Register(ctx *gin.Context) {
	var req models.RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(errors.BadRequest(err.Error()))
		return
	}

	result, err := c.useCase.Register(ctx.Request.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"id":    result.ID,
		"name":  result.Name,
		"email": result.Email,
	})
}

// Login autentica o usuário e retorna um JWT.
// POST /api/v1/auth/login
func (c *AuthController) Login(ctx *gin.Context) {
	var req models.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(errors.BadRequest(err.Error()))
		return
	}

	result, err := c.useCase.Login(ctx.Request.Context(), req.Email, req.Password)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, models.TokenResponse{
		Token:     result.Token,
		ExpiresAt: result.ExpiresAt.Format("2006-01-02T15:04:05Z"),
	})
}
