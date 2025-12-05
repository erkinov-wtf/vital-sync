package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/erkinov-wtf/vital-sync/internal/api/services"
	"github.com/erkinov-wtf/vital-sync/internal/enums"
	"github.com/erkinov-wtf/vital-sync/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DoctorHandler struct {
	doctorService *services.DoctorService
}

func NewDoctorHandler(doctorService *services.DoctorService) *DoctorHandler {
	return &DoctorHandler{
		doctorService: doctorService,
	}
}

func (h *DoctorHandler) CreateDoctor(c *gin.Context) {
	var body struct {
		Email          string        `json:"email" binding:"required,email"`
		PhoneNumber    string        `json:"phone_number" binding:"required"`
		Password       string        `json:"password" binding:"required"`
		FirstName      string        `json:"first_name" binding:"required"`
		LastName       string        `json:"last_name" binding:"required"`
		Gender         *enums.Gender `json:"gender"`
		IsActive       *bool         `json:"is_active"`
		OrganizationID uuid.UUID     `json:"organization_id" binding:"required"`
	}

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	doctor := models.User{
		Email:        body.Email,
		PhoneNumber:  body.PhoneNumber,
		PasswordHash: body.Password,
		FirstName:    body.FirstName,
		LastName:     body.LastName,
		Gender:       body.Gender,
	}
	if body.IsActive != nil {
		doctor.IsActive = *body.IsActive
	}

	if _, _, err := h.doctorService.CreateDoctor(&doctor, body.OrganizationID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create doctor: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, doctor)
}

func (h *DoctorHandler) ListDoctors(c *gin.Context) {
	includeInactive := false
	if includeInactiveStr := c.Query("include_inactive"); includeInactiveStr != "" {
		val, err := strconv.ParseBool(includeInactiveStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid include_inactive value"})
			return
		}
		includeInactive = val
	}

	doctors, err := h.doctorService.ListDoctors(includeInactive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list doctors: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": doctors})
}

func (h *DoctorHandler) GetDoctor(c *gin.Context) {
	doctorID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid doctor id"})
		return
	}

	doctor, err := h.doctorService.GetDoctorByID(doctorID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "doctor not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch doctor: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, doctor)
}

func (h *DoctorHandler) UpdateDoctor(c *gin.Context) {
	doctorID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid doctor id"})
		return
	}

	var body struct {
		Email       *string       `json:"email"`
		PhoneNumber *string       `json:"phone_number"`
		FirstName   *string       `json:"first_name"`
		LastName    *string       `json:"last_name"`
		Gender      *enums.Gender `json:"gender"`
		IsActive    *bool         `json:"is_active"`
		Password    *string       `json:"password"`
	}

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := h.doctorService.UpdateDoctor(doctorID, services.DoctorUpdate{
		Email:       body.Email,
		PhoneNumber: body.PhoneNumber,
		FirstName:   body.FirstName,
		LastName:    body.LastName,
		Gender:      body.Gender,
		IsActive:    body.IsActive,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "doctor not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update doctor: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, updated)
}

func (h *DoctorHandler) DeleteDoctor(c *gin.Context) {
	doctorID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid doctor id"})
		return
	}

	if err := h.doctorService.DeleteDoctor(doctorID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "doctor not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete doctor: " + err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *DoctorHandler) UnassignFromOrganization(c *gin.Context) {
	doctorID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid doctor id"})
		return
	}
	orgIDStr := c.Param("organizationId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}

	if err := h.doctorService.UnassignFromOrganization(doctorID, orgID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "assignment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unassign doctor: " + err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *DoctorHandler) ListDoctorOrganizations(c *gin.Context) {
	doctorID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid doctor id"})
		return
	}

	includeInactive := false
	if includeInactiveStr := c.Query("include_inactive"); includeInactiveStr != "" {
		val, err := strconv.ParseBool(includeInactiveStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid include_inactive value"})
			return
		}
		includeInactive = val
	}

	relations, err := h.doctorService.ListDoctorOrganizations(doctorID, includeInactive)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "doctor not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list doctor organizations: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": relations})
}

func (h *DoctorHandler) ListDoctorsByOrganization(c *gin.Context) {
	orgID, err := uuid.Parse(c.Param("organizationId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}

	includeInactive := false
	if includeInactiveStr := c.Query("include_inactive"); includeInactiveStr != "" {
		val, err := strconv.ParseBool(includeInactiveStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid include_inactive value"})
			return
		}
		includeInactive = val
	}

	doctors, err := h.doctorService.ListDoctorsByOrganization(orgID, includeInactive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list doctors: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": doctors})
}
