package routes

import (
	"github.com/DevKayoS/fintech-kodify/internal/controllers"
	"github.com/gin-gonic/gin"
)

func SetupHealthRoutes(rg *gin.RouterGroup, hc *controllers.HealthController) {
	health := rg.Group("/health")
	{
		health.GET("", hc.Check)
	}
}
