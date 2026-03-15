package routes

import (
	"github.com/DevKayoS/fintech-kodify/internal/adapters/rest/controllers"
	"github.com/gin-gonic/gin"
)

func SetupDashboardRoutes(rg *gin.RouterGroup, dc *controllers.DashboardController) {
	dashboard := rg.Group("/dashboard")
	{
		dashboard.GET("/summary", dc.Summary)
		dashboard.GET("/statement", dc.Statement)
	}
}
