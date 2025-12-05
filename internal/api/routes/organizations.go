package routes

import (
	"github.com/erkinov-wtf/vital-sync/internal/api/handlers"
	"github.com/gin-gonic/gin"
)

func registerOrgRoutes(r *gin.RouterGroup, handler *handlers.OrganizationHandler) {
	org := r.Group("/organizations")
	{
		org.POST("", handler.CreateOrganization)
		org.GET("", handler.ListOrganizations)
		org.GET("/:id", handler.GetOrganization)
		org.PUT("/:id", handler.UpdateOrganization)
		org.DELETE("/:id", handler.DeleteOrganization)
	}
}
