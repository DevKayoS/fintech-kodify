package routes

import (
	"github.com/DevKayoS/fintech-kodify/internal/adapters/rest/controllers"
	"github.com/gin-gonic/gin"
)

func SetupUserRoutes(rg *gin.RouterGroup, uc *controllers.UserController) {
	rg.GET("/me", uc.GetMe)
	rg.POST("/telegram/link", uc.GenerateTelegramLink)
}
