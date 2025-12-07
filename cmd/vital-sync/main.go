package main

import (
	"context"

	"github.com/erkinov-wtf/vital-sync/internal/api/handlers"
	"github.com/erkinov-wtf/vital-sync/internal/api/routes"
	"github.com/erkinov-wtf/vital-sync/internal/api/services"
	"github.com/erkinov-wtf/vital-sync/internal/config"
	"github.com/erkinov-wtf/vital-sync/internal/http"
	"github.com/erkinov-wtf/vital-sync/internal/pkg/logger"
	"github.com/erkinov-wtf/vital-sync/internal/storages/database"
	"github.com/erkinov-wtf/vital-sync/internal/workers"
)

func main() {
	cfg := config.MustLoad()
	lgr := logger.SetupLogger(cfg.Env)
	db, err := database.LoadDB(cfg, lgr)
	if err != nil {
		lgr.Error("couldn't load DB")
		return
	}

	ctx := context.Background()

	// svc init
	authSvc := services.NewAuthService(cfg, db.DB)
	orgSvc := services.NewOrganizationService(db.DB)
	checkinSvc := services.NewCheckinService(db.DB, cfg)
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

	// workers
	checkinScheduler := workers.NewCheckinScheduler(db.DB, lgr, checkinSvc, cfg.Timezone)
	checkinScheduler.Start(ctx)

	// engine and routes
	router := http.NewRouter(cfg, authSvc)
	routes.RegisterRoutes(router, orgHnr, userHnr, checkinHnr, checkinScheduleHnr, vitalReadingHnr, alertHnr)

	err = router.Run()
	if err != nil {
		lgr.Error("cant run the http engine", "error", err)
		return
	}
}
