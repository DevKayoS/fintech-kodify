package routes

import (
	"github.com/DevKayoS/fintech-kodify/internal/adapters/rest/controllers"
	"github.com/gin-gonic/gin"
)

func SetupRevenueRoutes(rg *gin.RouterGroup, rc *controllers.RevenueController) {
	revenues := rg.Group("/revenues")
	{
		revenues.POST("", rc.Create)
		revenues.GET("", rc.List)
	}
}
