package routes

import (
	"github.com/erkinov-wtf/vital-sync/internal/api/handlers"
	"github.com/gin-gonic/gin"
)

func registerCheckinScheduleRoutes(r *gin.RouterGroup, handler *handlers.CheckinScheduleHandler) {
	schedules := r.Group("/checkin-schedules")
	{
		schedules.POST("", handler.CreateSchedule)
		schedules.GET("", handler.ListSchedules)
		schedules.GET("/:id", handler.GetSchedule)
		schedules.PUT("/:id", handler.UpdateSchedule)
		schedules.DELETE("/:id", handler.DeleteSchedule)
	}
}
