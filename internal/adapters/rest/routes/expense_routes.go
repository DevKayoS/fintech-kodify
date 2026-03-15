package routes

import (
	"github.com/DevKayoS/fintech-kodify/internal/adapters/rest/controllers"
	"github.com/gin-gonic/gin"
)

func SetupExpenseRoutes(rg *gin.RouterGroup, ec *controllers.ExpenseController) {
	expenses := rg.Group("/expenses")
	{
		expenses.POST("", ec.Create)
		expenses.GET("", ec.List)
		expenses.GET("/categories", ec.ListCategories)
		expenses.GET("/:id", ec.GetByID)
		expenses.DELETE("/:id", ec.Delete)
	}
}
