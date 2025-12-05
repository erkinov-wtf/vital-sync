package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/erkinov-wtf/vital-sync/internal/api/services"
	"github.com/erkinov-wtf/vital-sync/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrganizationHandler struct {
	organizationService *services.OrganizationService
}

func NewOrganizationHandler(organizationService *services.OrganizationService) *OrganizationHandler {
	return &OrganizationHandler{
		organizationService: organizationService,
	}
}

func (h *OrganizationHandler) CreateOrganization(c *gin.Context) {
	var body struct {
		Name          string  `json:"name" binding:"required"`
		Address       *string `json:"address" binding:"required"`
		LicenseNumber string  `json:"license_number" binding:"required"`
		ContactEmail  *string `json:"contact_email" binding:"required"`
		ContactPhone  *string `json:"contact_phone" binding:"required"`
		IsActive      *bool   `json:"is_active" binding:"required"`
	}

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	org := models.Organization{
		Name:          body.Name,
		Address:       body.Address,
		LicenseNumber: body.LicenseNumber,
		ContactEmail:  body.ContactEmail,
		ContactPhone:  body.ContactPhone,
	}

	if body.IsActive != nil {
		org.IsActive = *body.IsActive
	}

	if err := h.organizationService.Create(&org); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create organization: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, org)
}

func (h *OrganizationHandler) ListOrganizations(c *gin.Context) {
	includeInactive := false
	if includeInactiveStr := c.Query("include_inactive"); includeInactiveStr != "" {
		parsed, err := strconv.ParseBool(includeInactiveStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid include_inactive value"})
			return
		}
		includeInactive = parsed
	}

	organizations, err := h.organizationService.List(includeInactive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list organizations: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": organizations})
}

func (h *OrganizationHandler) GetOrganization(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}

	organization, err := h.organizationService.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "organization not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch organization: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, organization)
}

func (h *OrganizationHandler) UpdateOrganization(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}

	var body struct {
		Name          *string `json:"name"`
		Address       *string `json:"address"`
		LicenseNumber *string `json:"license_number"`
		ContactEmail  *string `json:"contact_email"`
		ContactPhone  *string `json:"contact_phone"`
		IsActive      *bool   `json:"is_active"`
	}

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := h.organizationService.Update(id, services.OrganizationUpdate{
		Name:          body.Name,
		Address:       body.Address,
		LicenseNumber: body.LicenseNumber,
		ContactEmail:  body.ContactEmail,
		ContactPhone:  body.ContactPhone,
		IsActive:      body.IsActive,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "organization not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update organization: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, updated)
}

func (h *OrganizationHandler) DeleteOrganization(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}

	if err := h.organizationService.Delete(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "organization not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete organization: " + err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
