package controllers

import (
	"net/http"

	"github.com/DevKayoS/fintech-kodify/internal/services"
	"github.com/gin-gonic/gin"
)

type HealthController struct {
	service *services.HealthService
}

func NewHealthController(service *services.HealthService) *HealthController {
	return &HealthController{service: service}
}

func (h *HealthController) Check(ctx *gin.Context) {
	status := h.service.GetStatus(ctx.Request.Context())
	ctx.JSON(http.StatusOK, status)
}
