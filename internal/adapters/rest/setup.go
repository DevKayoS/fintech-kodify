package rest

import (
	"github.com/DevKayoS/fintech-kodify/internal/adapters/rest/controllers"
	"github.com/DevKayoS/fintech-kodify/internal/infrastructure/pgstore/database"
	"github.com/DevKayoS/fintech-kodify/internal/usecases/health"
	"github.com/DevKayoS/fintech-kodify/internal/usecases/user"
	"github.com/gin-gonic/gin"
)

func SetupAPI() *gin.Engine {
	r := gin.Default()

	healthUseCase := health.NewHealthUseCase(database.Pool)
	healthController := controllers.NewHealthController(healthUseCase)

	userUseCase := user.NewUserUseCase(database.Pool)
	userController := controllers.NewUserController(userUseCase)

	a := NewAPI(healthController, userController)
	a.BindRoutes(r)

	return r
}
