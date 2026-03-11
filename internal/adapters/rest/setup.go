package rest

import (
	"github.com/DevKayoS/fintech-kodify/internal/adapters/rest/controllers"
	"github.com/DevKayoS/fintech-kodify/internal/infrastructure/pgstore/database"
	"github.com/DevKayoS/fintech-kodify/internal/usecases/health"
	"github.com/gin-gonic/gin"
)

func SetupAPI() *gin.Engine {
	r := gin.Default()

	healthUseCase := health.NewHealthUseCase(database.Pool)
	healthController := controllers.NewHealthController(healthUseCase)

	a := NewAPI(healthController)
	a.BindRoutes(r)

	return r
}
