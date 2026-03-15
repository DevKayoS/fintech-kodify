package rest

import (
	"github.com/DevKayoS/fintech-kodify/internal/adapters/rest/controllers"
	"github.com/DevKayoS/fintech-kodify/internal/adapters/rest/routes"
	"github.com/DevKayoS/fintech-kodify/internal/middleware"
	"github.com/gin-gonic/gin"
)

type API struct {
	HealthController     *controllers.HealthController
	AuthController       *controllers.AuthController
	UserController       *controllers.UserController
	ExpenseController    *controllers.ExpenseController
	InvestmentController *controllers.InvestmentController
	RevenueController    *controllers.RevenueController
	DashboardController  *controllers.DashboardController
}

func NewAPI(
	hc *controllers.HealthController,
	ac *controllers.AuthController,
	uc *controllers.UserController,
	ec *controllers.ExpenseController,
	ic *controllers.InvestmentController,
	rc *controllers.RevenueController,
	dc *controllers.DashboardController,
) *API {
	return &API{
		HealthController:     hc,
		AuthController:       ac,
		UserController:       uc,
		ExpenseController:    ec,
		InvestmentController: ic,
		RevenueController:    rc,
		DashboardController:  dc,
	}
}

func (a *API) BindRoutes(r *gin.Engine) {
	r.Use(middleware.ErrorHandler())

	v1 := r.Group("/api/v1")

	// Rotas públicas — sem autenticação
	public := v1.Group("/")
	{
		routes.SetupHealthRoutes(public, a.HealthController)
		routes.SetupAuthRoutes(public, a.AuthController)
	}

	// Rotas protegidas — exigem JWT válido
	protected := v1.Group("/")
	protected.Use(middleware.AuthMiddleware())
	{
		routes.SetupUserRoutes(protected, a.UserController)
		routes.SetupExpenseRoutes(protected, a.ExpenseController)
		routes.SetupInvestmentRoutes(protected, a.InvestmentController)
		routes.SetupRevenueRoutes(protected, a.RevenueController)
		routes.SetupDashboardRoutes(protected, a.DashboardController)
	}
}
