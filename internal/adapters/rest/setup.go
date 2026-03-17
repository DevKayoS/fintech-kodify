package rest

import (
	"time"

	"github.com/DevKayoS/fintech-kodify/internal/adapters/rest/controllers"
	"github.com/DevKayoS/fintech-kodify/internal/infrastructure/pgstore/database"
	"github.com/DevKayoS/fintech-kodify/internal/middleware"
	authUC "github.com/DevKayoS/fintech-kodify/internal/usecases/auth"
	expenseUC "github.com/DevKayoS/fintech-kodify/internal/usecases/expense"
	"github.com/DevKayoS/fintech-kodify/internal/usecases/health"
	investmentUC "github.com/DevKayoS/fintech-kodify/internal/usecases/investment"
	revenueUC "github.com/DevKayoS/fintech-kodify/internal/usecases/revenue"
	statementUC "github.com/DevKayoS/fintech-kodify/internal/usecases/statement"
	summaryUC "github.com/DevKayoS/fintech-kodify/internal/usecases/summary"
	userUC "github.com/DevKayoS/fintech-kodify/internal/usecases/user"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupAPI() *gin.Engine {
	r := gin.Default()

	r.Use(middleware.RateLimiter(60, time.Minute))

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://fin.kodify.com.br"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

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
