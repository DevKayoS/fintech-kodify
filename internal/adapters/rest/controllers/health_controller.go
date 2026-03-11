package controllers

import (
	"net/http"

	"github.com/DevKayoS/fintech-kodify/internal/usecases/health"
	"github.com/gin-gonic/gin"
)

type HealthController struct {
	useCase *health.HealthUseCase
}

func NewHealthController(useCase *health.HealthUseCase) *HealthController {
	return &HealthController{useCase: useCase}
}

func (h *HealthController) Check(ctx *gin.Context) {
	status := h.useCase.GetStatus(ctx.Request.Context())
	ctx.JSON(http.StatusOK, status)
}
