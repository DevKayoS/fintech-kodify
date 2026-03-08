package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthController struct{}

func NewHealthController() *HealthController {
	return &HealthController{}
}

func (h *HealthController) Check(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "API is running",
	})
}
