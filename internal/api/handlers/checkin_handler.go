package handlers

import (
	"errors"
	"net/http"

	"github.com/erkinov-wtf/vital-sync/internal/api/services"
	"github.com/erkinov-wtf/vital-sync/internal/enums"
	"github.com/erkinov-wtf/vital-sync/internal/models"
	"github.com/erkinov-wtf/vital-sync/internal/pkg/errs"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CheckinHandler struct {
	checkinService *services.CheckinService
}

func NewCheckinHandler(checkinService *services.CheckinService) *CheckinHandler {
	return &CheckinHandler{checkinService: checkinService}
}

func (h *CheckinHandler) StartCheckin(c *gin.Context) {
	var body struct {
		PatientID  uuid.UUID  `json:"patient_id" binding:"required"`
		ScheduleID *uuid.UUID `json:"schedule_id"`
	}

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	checkin, err := h.checkinService.StartCheckin(body.PatientID, body.ScheduleID)
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrActiveCheckinExists):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case errors.Is(err, gorm.ErrRecordNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "patient not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start checkin: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, checkin)
}

func (h *CheckinHandler) EndCheckin(c *gin.Context) {
	patientUserID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid patient id"})
		return
	}

	checkin, err := h.checkinService.EndCheckin(patientUserID)
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrCheckinNotActive):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, gorm.ErrRecordNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "checkin not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to end checkin: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, checkin)
}

func (h *CheckinHandler) GetActiveCheckin(c *gin.Context) {
	patientID, err := uuid.Parse(c.Param("patientId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid patient id"})
		return
	}

	checkin, err := h.checkinService.GetActiveCheckin(patientID)
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrNoActiveCheckin):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, gorm.ErrRecordNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "patient not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch active checkin: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, checkin)
}

func (h *CheckinHandler) ListCompletedCheckins(c *gin.Context) {
	patientID, err := uuid.Parse(c.Param("patientId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid patient id"})
		return
	}

	checkins, err := h.checkinService.ListCompletedByPatient(patientID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "patient not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list checkins: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": checkins})
}

func (h *CheckinHandler) UpdateCheckinAI(c *gin.Context) {
	checkinID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid checkin id"})
		return
	}

	var body struct {
		AIAnalysis    *models.JSONB        `json:"ai_analysis"`
		MedicalStatus *enums.MedicalStatus `json:"medical_status"`
		RiskScore     *int                 `json:"risk_score"`
		Alert         *struct {
			Severity  enums.AlertSeverity `json:"severity"`
			AlertType enums.AlertType     `json:"alert_type"`
			Title     string              `json:"title"`
			Message   string              `json:"message"`
			Details   *models.JSONB       `json:"details"`
		} `json:"alert"`
	}

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var alertInput *services.CheckinAIAlertInput
	if body.Alert != nil {
		alertInput = &services.CheckinAIAlertInput{
			Severity:  body.Alert.Severity,
			AlertType: body.Alert.AlertType,
			Title:     body.Alert.Title,
			Message:   body.Alert.Message,
			Details:   body.Alert.Details,
		}
	}

	updated, err := h.checkinService.UpdateAIFields(checkinID, services.CheckinAIUpdate{
		AIAnalysis:    body.AIAnalysis,
		MedicalStatus: body.MedicalStatus,
		RiskScore:     body.RiskScore,
		Alert:         alertInput,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "checkin not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update checkin analysis: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, updated)
}

func (h *CheckinHandler) ReviewCheckin(c *gin.Context) {
	checkinID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid checkin id"})
		return
	}

	var body struct {
		DoctorID    uuid.UUID `json:"doctor_id" binding:"required"`
		DoctorNotes *string   `json:"doctor_notes"`
	}

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := h.checkinService.ReviewCheckin(checkinID, body.DoctorID, body.DoctorNotes)
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "doctor or checkin not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to review checkin: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, updated)
}

func (h *CheckinHandler) AddQuestions(c *gin.Context) {
	checkinID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid checkin id"})
		return
	}

	var body struct {
		Items []interface{} `json:"items" binding:"required"`
	}

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(body.Items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "items cannot be empty"})
		return
	}

	checkin, err := h.checkinService.AddQuestions(checkinID, body.Items)
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrCheckinNotActive):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, gorm.ErrRecordNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "checkin not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update questions: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, checkin)
}

func (h *CheckinHandler) AddAnswers(c *gin.Context) {
	checkinID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid checkin id"})
		return
	}

	var body struct {
		Items []interface{} `json:"items" binding:"required"`
	}

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(body.Items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "items cannot be empty"})
		return
	}

	checkin, err := h.checkinService.AddAnswers(checkinID, body.Items)
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrCheckinNotActive):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, gorm.ErrRecordNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "checkin not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update answers: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, checkin)
}

func (h *CheckinHandler) GetCheckin(c *gin.Context) {
	checkinID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid checkin id"})
		return
	}

	checkin, err := h.checkinService.GetByID(checkinID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "checkin not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch checkin: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, checkin)
}

func (h *CheckinHandler) ManualCheckin(c *gin.Context) {
	patientID, err := uuid.Parse(c.Param("patientId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid patient id"})
		return
	}

	checkingType := c.Query("type")
	checkin, err := h.checkinService.StartManualCheckin(patientID, checkingType)
	if errors.Is(err, errs.ErrActiveCheckinExists) {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "patient not found"})
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start manual checkin: " + err.Error()})
	}

	c.JSON(http.StatusCreated, checkin)
}
