package main

import (
	"github.com/erkinov-wtf/vital-sync/internal/api/handlers"
	"github.com/erkinov-wtf/vital-sync/internal/api/routes"
	"github.com/erkinov-wtf/vital-sync/internal/api/services"
	"github.com/erkinov-wtf/vital-sync/internal/config"
	"github.com/erkinov-wtf/vital-sync/internal/http"
	"github.com/erkinov-wtf/vital-sync/internal/pkg/logger"
	"github.com/erkinov-wtf/vital-sync/internal/storages/database"
)

func main() {
	cfg := config.MustLoad()
	lgr := logger.SetupLogger(cfg.Env)
	db, err := database.LoadDB(cfg, lgr)
	if err != nil {
		lgr.Error("couldn't load DB")
		return
	}

	// svc init
	authSvc := services.NewAuthService(cfg, db.DB)
	orgSvc := services.NewOrganizationService(db.DB)
	checkinSvc := services.NewCheckinService(db.DB)
	checkinScheduleSvc := services.NewCheckinScheduleService(db.DB)
	vitalReadingSvc := services.NewVitalReadingService(db.DB)
	alertSvc := services.NewAlertService(db.DB)
	userSvc := services.NewUserService(db.DB, lgr)

	// hnr init
	orgHnr := handlers.NewOrganizationHandler(orgSvc)
	checkinHnr := handlers.NewCheckinHandler(checkinSvc)
	checkinScheduleHnr := handlers.NewCheckinScheduleHandler(checkinScheduleSvc)
	vitalReadingHnr := handlers.NewVitalReadingHandler(vitalReadingSvc)
	alertHnr := handlers.NewAlertHandler(alertSvc)
	userHnr := handlers.NewUserHandler(userSvc)

	// engine and routes
	router := http.NewRouter(cfg, authSvc)
	routes.RegisterRoutes(router, orgHnr, userHnr, checkinHnr, checkinScheduleHnr, vitalReadingHnr, alertHnr)

	err = router.Run()
	if err != nil {
		lgr.Error("cant run the http engine", err.Error())
		return
	}
}
