package routes

import (
	"github.com/erkinov-wtf/vital-sync/internal/api/handlers"
	"github.com/gin-gonic/gin"
)

func registerUserRoutes(r *gin.RouterGroup, handler *handlers.UserHandler) {
	org := r.Group("/users")
	{
		// doctors
		org.POST("/doctors", handler.CreateDoctor)
		org.GET("/doctors", handler.ListDoctors)
		org.GET("/doctors/:id", handler.GetDoctor)
		org.GET("/doctors/:id/organizations", handler.ListDoctorOrganizations)
		org.PUT("/doctors/:id", handler.UpdateDoctor)
		org.DELETE("/doctors/:id", handler.DeleteDoctor)

		// patients
		org.POST("/patients", handler.CreatePatient)
		org.GET("/patients", handler.ListPatients)
		org.GET("/patients/:id", handler.GetPatient)
		org.POST("/patients/:id/medical", handler.CreatePatientMedicalInfo)
		org.PUT("/patients/:id", handler.UpdatePatient)
		org.PUT("/patients/:id/medical", handler.UpdatePatientMedicalInfo)
	}
}
