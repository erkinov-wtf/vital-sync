package http

import (
	"fmt"

	"github.com/erkinov-wtf/vital-sync/internal/config"
	"github.com/gin-gonic/gin"
)

// Router wraps the gin engine
type Router struct {
	engine *gin.Engine
	config *config.Config
}

// NewRouter creates a new router instance
func NewRouter(cfg *config.Config) *Router {
	if cfg.Env == config.ReleaseEnv {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// Middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	//TODO use custom AUTH() middleware

	return &Router{
		engine: r,
		config: cfg,
	}
}

func (r *Router) Engine() *gin.Engine {
	return r.engine
}

func (r *Router) Run() error {
	return r.engine.Run(fmt.Sprintf("%s:%d", r.config.Internal.Server.Host, r.config.Internal.Server.Port))
}
