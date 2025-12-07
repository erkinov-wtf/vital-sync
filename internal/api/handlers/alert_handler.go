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

type alertResponse struct {
	ID             uuid.UUID           `json:"ID"`
	CheckinID      *uuid.UUID          `json:"CheckinID"`
	PatientUserID  uuid.UUID           `json:"PatientUserID"`
	Severity       enums.AlertSeverity `json:"Severity"`
	AlertType      enums.AlertType     `json:"AlertType"`
	Title          string              `json:"Title"`
	Message        string              `json:"Message"`
	Details        models.JSONB        `json:"Details"`
	IsAcknowledged bool                `json:"IsAcknowledged"`
	AcknowledgedBy *uuid.UUID          `json:"AcknowledgedBy"`
	AcknowledgedAt *time.Time          `json:"AcknowledgedAt"`
	ActionTaken    *string             `json:"ActionTaken"`
	CreatedAt      time.Time           `json:"CreatedAt"`
}

type AlertHandler struct {
	alertService *services.AlertService
}

func NewAlertHandler(service *services.AlertService) *AlertHandler {
	return &AlertHandler{alertService: service}
}

func (h *AlertHandler) ListDoctorAlerts(c *gin.Context) {
	doctorID, err := uuid.Parse(c.Param("doctorId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid doctor id"})
		return
	}

	includeAcknowledged := false
	if raw := c.Query("show_all"); raw != "" {
		val, err := strconv.ParseBool(raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid show_all value"})
			return
		}
		includeAcknowledged = val
	}

	alerts, err := h.alertService.ListByDoctor(doctorID, includeAcknowledged)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "doctor not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list alerts: " + err.Error()})
		return
	}

	resp := make([]alertResponse, 0, len(alerts))
	for _, a := range alerts {
		patientUserID := uuid.Nil
		if a.Patient != nil {
			patientUserID = a.Patient.UserID
		}

		resp = append(resp, alertResponse{
			ID:             a.ID,
			CheckinID:      a.CheckinID,
			PatientUserID:  patientUserID,
			Severity:       a.Severity,
			AlertType:      a.AlertType,
			Title:          a.Title,
			Message:        a.Message,
			Details:        a.Details,
			IsAcknowledged: a.IsAcknowledged,
			AcknowledgedBy: a.AcknowledgedBy,
			AcknowledgedAt: a.AcknowledgedAt,
			ActionTaken:    a.ActionTaken,
			CreatedAt:      a.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": resp})
}
