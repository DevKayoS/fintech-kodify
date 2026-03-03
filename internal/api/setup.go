package api

import (
	"github.com/DevKayoS/fintech-kodify/internal/controllers"
	"github.com/gin-gonic/gin"
)

func SetupAPI() *gin.Engine {
	r := gin.Default()

	healthController := controllers.NewHealthController()

	// TODO: inicializar services e controllers quando implementados
	// tokenService := token.NewTokenService(database.Pool)
	// tokenController := controllers.NewTokenController(tokenService)
	//
	// userService := user.NewUserService(database.Pool)
	// userController := controllers.NewUserController(userService)
	//
	// expenseService := expense.NewExpenseService(database.Pool)
	// expenseController := controllers.NewExpenseController(expenseService)
	//
	// investmentService := investment.NewInvestmentService(database.Pool)
	// investmentController := controllers.NewInvestmentController(investmentService)

	a := NewAPI(healthController)
	a.BindRoutes(r)

	return r
}
