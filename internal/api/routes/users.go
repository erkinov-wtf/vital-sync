package routes

import (
	"github.com/erkinov-wtf/vital-sync/internal/api/handlers"
	"github.com/gin-gonic/gin"
)

func registerUserRoutes(r *gin.RouterGroup, handler *handlers.UserHandler) {
	users := r.Group("/users")
	{
		// doctors
		users.POST("/doctors", handler.CreateDoctor)
		users.GET("/doctors", handler.ListDoctors)
		users.GET("/doctors/:id", handler.GetDoctor)
		users.GET("/doctors/:id/organizations", handler.ListDoctorOrganizations)
		users.PUT("/doctors/:id", handler.UpdateDoctor)
		users.DELETE("/doctors/:id", handler.DeleteDoctor)

		// patients
		users.POST("/patients", handler.CreatePatient)
		users.GET("/patients", handler.ListPatients)
		users.GET("/patients/:id", handler.GetPatient)
		users.POST("/patients/:id/medical", handler.CreatePatientMedicalInfo)
		users.PUT("/patients/:id", handler.UpdatePatient)
		users.PUT("/patients/:id/medical", handler.UpdatePatientMedicalInfo)
		users.GET("/patients/telegram/:username", handler.GetUserByTgUsername)
	}
}
