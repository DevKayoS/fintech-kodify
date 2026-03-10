package api

import (
	"github.com/DevKayoS/fintech-kodify/internal/controllers"
	"github.com/DevKayoS/fintech-kodify/internal/pgstore/database"
	"github.com/DevKayoS/fintech-kodify/internal/services"
	"github.com/gin-gonic/gin"
)

func SetupAPI() *gin.Engine {
	r := gin.Default()

	healthService := services.NewHealthService(database.Pool)
	healthController := controllers.NewHealthController(healthService)

	a := NewAPI(healthController)
	a.BindRoutes(r)

	return r
}
