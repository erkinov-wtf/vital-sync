package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/erkinov-wtf/vital-sync/internal/api/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

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

	c.JSON(http.StatusOK, gin.H{"data": alerts})
}
