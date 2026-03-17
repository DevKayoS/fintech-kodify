package routes

import (
	"github.com/DevKayoS/fintech-kodify/internal/adapters/rest/controllers"
	"github.com/gin-gonic/gin"
)

func SetupInvestmentRoutes(rg *gin.RouterGroup, ic *controllers.InvestmentController) {
	investments := rg.Group("/investments")
	{
		investments.POST("", ic.Create)
		investments.POST("/withdraw", ic.Withdraw)
		investments.GET("", ic.List)
		investments.GET("/types", ic.ListTypes)
		investments.GET("/:id", ic.GetByID)
		investments.DELETE("/:id", ic.Delete)
	}
}
