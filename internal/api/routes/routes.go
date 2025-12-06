package routes

import (
	"github.com/erkinov-wtf/vital-sync/internal/api/handlers"
	"github.com/erkinov-wtf/vital-sync/internal/http"
)

func RegisterRoutes(
	router *http.Router,
	orgHnr *handlers.OrganizationHandler,
	userHnr *handlers.UserHandler,
	checkinHnr *handlers.CheckinHandler,
	checkinScheduleHnr *handlers.CheckinScheduleHandler,
	vitalReadingHnr *handlers.VitalReadingHandler,
) {
	api := router.Engine().Group("/api/v1")
	{
		registerOrgRoutes(api, orgHnr)
		registerUserRoutes(api, userHnr)
		registerCheckinRoutes(api, checkinHnr)
		registerCheckinScheduleRoutes(api, checkinScheduleHnr)
		registerVitalReadingRoutes(api, vitalReadingHnr)
	}
}
