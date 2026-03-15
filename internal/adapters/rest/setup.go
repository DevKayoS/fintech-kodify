package rest

import (
	"github.com/DevKayoS/fintech-kodify/internal/adapters/rest/controllers"
	"github.com/DevKayoS/fintech-kodify/internal/infrastructure/pgstore/database"
	authUC "github.com/DevKayoS/fintech-kodify/internal/usecases/auth"
	expenseUC "github.com/DevKayoS/fintech-kodify/internal/usecases/expense"
	"github.com/DevKayoS/fintech-kodify/internal/usecases/health"
	investmentUC "github.com/DevKayoS/fintech-kodify/internal/usecases/investment"
	revenueUC "github.com/DevKayoS/fintech-kodify/internal/usecases/revenue"
	statementUC "github.com/DevKayoS/fintech-kodify/internal/usecases/statement"
	summaryUC "github.com/DevKayoS/fintech-kodify/internal/usecases/summary"
	userUC "github.com/DevKayoS/fintech-kodify/internal/usecases/user"
	"github.com/gin-gonic/gin"
)

func SetupAPI() *gin.Engine {
	r := gin.Default()

	pool := database.Pool

	healthUseCase := health.NewHealthUseCase(pool)
	healthController := controllers.NewHealthController(healthUseCase)

	authUseCase := authUC.NewAuthUseCase(pool)

	userUseCase := userUC.NewUserUseCase(pool)
	userController := controllers.NewUserController(userUseCase)

	authController := controllers.NewAuthController(authUseCase)

	expenseUseCase := expenseUC.NewExpenseUseCase(pool)
	expenseController := controllers.NewExpenseController(expenseUseCase)

	investmentUseCase := investmentUC.NewInvestmentUseCase(pool)
	investmentController := controllers.NewInvestmentController(investmentUseCase)

	revenueUseCase := revenueUC.NewRevenueUseCase(pool)
	revenueController := controllers.NewRevenueController(revenueUseCase)

	summaryUseCase := summaryUC.NewSummaryUseCase(pool)
	statementUseCase := statementUC.NewStatementUseCase(pool)
	dashboardController := controllers.NewDashboardController(summaryUseCase, statementUseCase)

	a := NewAPI(healthController, authController, userController, expenseController, investmentController, revenueController, dashboardController)
	a.BindRoutes(r)

	return r
}
