package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/erkinov-wtf/vital-sync/internal/api/services"
	"github.com/erkinov-wtf/vital-sync/internal/enums"
	"github.com/erkinov-wtf/vital-sync/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) CreateDoctor(c *gin.Context) {
	var body struct {
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
		PhoneNumber:  body.PhoneNumber,
		PasswordHash: body.Password,
		FirstName:    body.FirstName,
		LastName:     body.LastName,
		Gender:       body.Gender,
	}
	if body.IsActive != nil {
		doctor.IsActive = *body.IsActive
	}

	if _, _, err := h.userService.CreateDoctor(&doctor, body.OrganizationID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create doctor: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, doctor)
}

func (h *UserHandler) ListDoctors(c *gin.Context) {
	includeInactive := false
	if includeInactiveStr := c.Query("include_inactive"); includeInactiveStr != "" {
		val, err := strconv.ParseBool(includeInactiveStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid include_inactive value"})
			return
		}
		includeInactive = val
	}

	doctors, err := h.userService.ListDoctors(includeInactive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list doctors: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": doctors})
}

func (h *UserHandler) GetDoctor(c *gin.Context) {
	doctorID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid doctor id"})
		return
	}

	doctor, err := h.userService.GetDoctorByID(doctorID)
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

func (h *UserHandler) UpdateDoctor(c *gin.Context) {
	doctorID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid doctor id"})
		return
	}

	var body struct {
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

	updated, err := h.userService.UpdateDoctor(doctorID, services.DoctorUpdate{
		PhoneNumber: body.PhoneNumber,
		FirstName:   body.FirstName,
		LastName:    body.LastName,
		Gender:      body.Gender,
		IsActive:    body.IsActive,
		Password:    body.Password,
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

func (h *UserHandler) DeleteDoctor(c *gin.Context) {
	doctorID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid doctor id"})
		return
	}

	if err := h.userService.DeleteDoctor(doctorID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "doctor not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete doctor: " + err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *UserHandler) UnassignFromOrganization(c *gin.Context) {
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

	if err := h.userService.UnassignFromOrganization(doctorID, orgID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "assignment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unassign doctor: " + err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *UserHandler) ListDoctorOrganizations(c *gin.Context) {
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

	relations, err := h.userService.ListDoctorOrganizations(doctorID, includeInactive)
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

func (h *UserHandler) CreatePatient(c *gin.Context) {
	var body struct {
		PhoneNumber string        `json:"phone_number" binding:"required"`
		Password    string        `json:"password" binding:"required"`
		FirstName   string        `json:"first_name" binding:"required"`
		LastName    string        `json:"last_name" binding:"required"`
		Gender      *enums.Gender `json:"gender"`
		IsActive    *bool         `json:"is_active"`
	}

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.CreatePatientUser(services.CreatePatientUserInput{
		PhoneNumber: body.PhoneNumber,
		Password:    body.Password,
		FirstName:   body.FirstName,
		LastName:    body.LastName,
		Gender:      body.Gender,
		IsActive:    body.IsActive,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create patient user: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (h *UserHandler) CreatePatientMedicalInfo(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid patient id"})
		return
	}

	var body struct {
		DoctorID                 uuid.UUID                  `json:"doctor_id" binding:"required"`
		ConditionSummary         string                     `json:"condition_summary" binding:"required"`
		Comorbidities            []string                   `json:"comorbidities"`
		CurrentMedications       models.JSONB               `json:"current_medications"`
		Allergies                []string                   `json:"allergies"`
		BaselineVitals           models.JSONB               `json:"baseline_vitals"`
		RiskLevel                *enums.RiskLevel           `json:"risk_level"`
		MonitoringFrequency      *enums.MonitoringFrequency `json:"monitoring_frequency"`
		Status                   *enums.PatientStatus       `json:"status"`
		DischargeDate            *time.Time                 `json:"discharge_date"`
		DischargeNotes           *string                    `json:"discharge_notes"`
		EmergencyContactName     *string                    `json:"emergency_contact_name"`
		EmergencyContactPhone    *string                    `json:"emergency_contact_phone"`
		EmergencyContactRelation *string                    `json:"emergency_contact_relation"`
	}

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	patient, err := h.userService.CreatePatientMedicalInfo(userID, services.PatientMedicalInput{
		DoctorID:                 body.DoctorID,
		ConditionSummary:         body.ConditionSummary,
		Comorbidities:            body.Comorbidities,
		CurrentMedications:       body.CurrentMedications,
		Allergies:                body.Allergies,
		BaselineVitals:           body.BaselineVitals,
		RiskLevel:                body.RiskLevel,
		MonitoringFrequency:      body.MonitoringFrequency,
		Status:                   body.Status,
		DischargeDate:            body.DischargeDate,
		DischargeNotes:           body.DischargeNotes,
		EmergencyContactName:     body.EmergencyContactName,
		EmergencyContactPhone:    body.EmergencyContactPhone,
		EmergencyContactRelation: body.EmergencyContactRelation,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "patient or doctor not found"})
			return
		}
		if err.Error() == "patient medical info already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create patient medical info: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, patient)
}

func (h *UserHandler) UpdatePatient(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid patient id"})
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

	user, err := h.userService.UpdatePatientUser(userID, services.UpdatePatientUserInput{
		Email:       body.Email,
		PhoneNumber: body.PhoneNumber,
		FirstName:   body.FirstName,
		LastName:    body.LastName,
		Gender:      body.Gender,
		IsActive:    body.IsActive,
		Password:    body.Password,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "patient not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update patient user: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) UpdatePatientMedicalInfo(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid patient id"})
		return
	}

	var body struct {
		DoctorID                 *uuid.UUID                 `json:"doctor_id"`
		ConditionSummary         *string                    `json:"condition_summary"`
		Comorbidities            *[]string                  `json:"comorbidities"`
		CurrentMedications       *models.JSONB              `json:"current_medications"`
		Allergies                *[]string                  `json:"allergies"`
		BaselineVitals           *models.JSONB              `json:"baseline_vitals"`
		RiskLevel                *enums.RiskLevel           `json:"risk_level"`
		MonitoringFrequency      *enums.MonitoringFrequency `json:"monitoring_frequency"`
		Status                   *enums.PatientStatus       `json:"status"`
		DischargeDate            *time.Time                 `json:"discharge_date"`
		DischargeNotes           *string                    `json:"discharge_notes"`
		EmergencyContactName     *string                    `json:"emergency_contact_name"`
		EmergencyContactPhone    *string                    `json:"emergency_contact_phone"`
		EmergencyContactRelation *string                    `json:"emergency_contact_relation"`
	}

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	patient, err := h.userService.UpdatePatientMedicalInfo(userID, services.PatientMedicalUpdate{
		DoctorID:                 body.DoctorID,
		ConditionSummary:         body.ConditionSummary,
		Comorbidities:            body.Comorbidities,
		CurrentMedications:       body.CurrentMedications,
		Allergies:                body.Allergies,
		BaselineVitals:           body.BaselineVitals,
		RiskLevel:                body.RiskLevel,
		MonitoringFrequency:      body.MonitoringFrequency,
		Status:                   body.Status,
		DischargeDate:            body.DischargeDate,
		DischargeNotes:           body.DischargeNotes,
		EmergencyContactName:     body.EmergencyContactName,
		EmergencyContactPhone:    body.EmergencyContactPhone,
		EmergencyContactRelation: body.EmergencyContactRelation,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "patient not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update patient medical info: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, patient)
}

func (h *UserHandler) ListPatients(c *gin.Context) {
	includeInactive := false
	if includeInactiveStr := c.Query("include_inactive"); includeInactiveStr != "" {
		val, err := strconv.ParseBool(includeInactiveStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid include_inactive value"})
			return
		}
		includeInactive = val
	}

	patients, err := h.userService.ListPatientUsers(includeInactive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list patients: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": patients})
}

func (h *UserHandler) GetPatient(c *gin.Context) {
	patientID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid patient id"})
		return
	}

	patient, err := h.userService.GetPatientDetailsByUserID(patientID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "patient not found or no medical data found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch patient: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, patient)
}
