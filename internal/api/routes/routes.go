package routes

import (
	"github.com/erkinov-wtf/vital-sync/internal/api/handlers"
	"github.com/erkinov-wtf/vital-sync/internal/http"
)

func RegisterRoutes(router *http.Router,
	orgHnr *handlers.OrganizationHandler,
	userHnr *handlers.UserHandler,
) {
	api := router.Engine().Group("/api/v1")
	{
		registerOrgRoutes(api, orgHnr)
		registerUserRoutes(api, userHnr)
	}
}
