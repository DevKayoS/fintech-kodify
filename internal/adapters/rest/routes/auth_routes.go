package routes

import (
	"github.com/DevKayoS/fintech-kodify/internal/adapters/rest/controllers"
	"github.com/gin-gonic/gin"
)

func SetupAuthRoutes(rg *gin.RouterGroup, ac *controllers.AuthController) {
	auth := rg.Group("/auth")
	{
		auth.POST("/register", ac.Register)
		auth.POST("/login", ac.Login)
	}
}
