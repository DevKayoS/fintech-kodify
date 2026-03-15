package rest

import (
	"github.com/DevKayoS/fintech-kodify/internal/adapters/rest/controllers"
	"github.com/DevKayoS/fintech-kodify/internal/adapters/rest/routes"
	"github.com/DevKayoS/fintech-kodify/internal/middleware"
	"github.com/gin-gonic/gin"
)

type API struct {
	HealthController *controllers.HealthController
	UserController   *controllers.UserController
}

func NewAPI(hc *controllers.HealthController, uc *controllers.UserController) *API {
	return &API{
		HealthController: hc,
		UserController:   uc,
	}
}

func (a *API) BindRoutes(r *gin.Engine) {
	r.Use(middleware.ErrorHandler())

	v1 := r.Group("/api/v1")

	public := v1.Group("/")
	{
		routes.SetupHealthRoutes(public, a.HealthController)
	}

	protected := v1.Group("/")
	protected.Use(middleware.AuthMiddleware())
	{
		routes.SetupUserRoutes(protected, a.UserController)
	}
}
