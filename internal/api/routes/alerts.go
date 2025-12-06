package routes

import (
	"github.com/erkinov-wtf/vital-sync/internal/api/handlers"
	"github.com/gin-gonic/gin"
)

func registerAlertRoutes(r *gin.RouterGroup, handler *handlers.AlertHandler) {
	alerts := r.Group("/alerts")
	{
		alerts.GET("/:doctorId", handler.ListDoctorAlerts)
	}
}
