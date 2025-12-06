package routes

import (
	"github.com/erkinov-wtf/vital-sync/internal/api/handlers"
	"github.com/gin-gonic/gin"
)

func registerVitalReadingRoutes(r *gin.RouterGroup, handler *handlers.VitalReadingHandler) {
	vitals := r.Group("/vital-readings")
	{
		vitals.POST("", handler.Create)
		vitals.GET("", handler.List)
		vitals.GET("/:id", handler.Get)
		vitals.PUT("/:id", handler.Update)
		vitals.DELETE("/:id", handler.Delete)
	}
}
