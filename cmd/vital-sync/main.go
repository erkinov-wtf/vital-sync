package vital_sync

import (
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

	// engine and routes
	router := http.NewRouter(cfg, authSvc)

	router.Run()
}
