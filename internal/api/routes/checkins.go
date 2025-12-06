package routes

import (
	"github.com/erkinov-wtf/vital-sync/internal/api/handlers"
	"github.com/gin-gonic/gin"
)

func registerCheckinRoutes(r *gin.RouterGroup, handler *handlers.CheckinHandler) {
	checkins := r.Group("/checkins")
	{
		checkins.POST("/start", handler.StartCheckin)
		checkins.POST("/:id/end", handler.EndCheckin)
		checkins.GET("/active/:patientId", handler.GetActiveCheckin)
		checkins.POST("/:id/questions", handler.AddQuestions)
		checkins.POST("/:id/answers", handler.AddAnswers)
		checkins.GET("/:id", handler.GetCheckin)
	}
}
