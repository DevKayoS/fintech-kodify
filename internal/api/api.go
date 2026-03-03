package api

import (
	"github.com/DevKayoS/fintech-kodify/internal/controllers"
	"github.com/DevKayoS/fintech-kodify/internal/middleware"
	"github.com/DevKayoS/fintech-kodify/internal/routes"
	"github.com/gin-gonic/gin"
)

type API struct {
	HealthController *controllers.HealthController
	// TODO: adicionar controllers conforme implementados
	// TokenController      *controllers.TokenController
	// UserController       *controllers.UserController
	// ExpenseController    *controllers.ExpenseController
	// InvestmentController *controllers.InvestmentController
}

func NewAPI(hc *controllers.HealthController) *API {
	return &API{
		HealthController: hc,
	}
}

func (a *API) BindRoutes(r *gin.Engine) {
	r.Use(middleware.ErrorHandler())

	v1 := r.Group("/api/v1")

	public := v1.Group("/")
	{
		routes.SetupHealthRoutes(public, a.HealthController)
		// TODO: routes.SetupTokenRoutes(public, a.TokenController)
		// TODO: routes.SetupUserRoutes(public, a.UserController) // POST /users (criação pública)
	}

	// protected := v1.Group("/")
	// protected.Use(middleware.AuthMiddleware())
	// {
	// 	TODO: routes.SetupUserRoutes(protected, a.UserController)
	// 	TODO: routes.SetupExpenseRoutes(protected, a.ExpenseController)
	// 	TODO: routes.SetupInvestmentRoutes(protected, a.InvestmentController)
	// }
}
