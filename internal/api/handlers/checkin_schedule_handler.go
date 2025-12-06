package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/erkinov-wtf/vital-sync/internal/api/services"
	"github.com/erkinov-wtf/vital-sync/internal/enums"
	"github.com/erkinov-wtf/vital-sync/internal/pkg/errs"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CheckinScheduleHandler struct {
	checkinScheduleService *services.CheckinScheduleService
}

func NewCheckinScheduleHandler(service *services.CheckinScheduleService) *CheckinScheduleHandler {
	return &CheckinScheduleHandler{
		checkinScheduleService: service,
	}
}

func (h *CheckinScheduleHandler) CreateSchedule(c *gin.Context) {
	var body struct {
		PatientID     uuid.UUID               `json:"patient_id" binding:"required"`
		Frequency     enums.ScheduleFrequency `json:"frequency" binding:"required"`
		TimeSlots     []string                `json:"time_slots" binding:"required"`
		Timezone      *string                 `json:"timezone"`
		IsActive      *bool                   `json:"is_active"`
		NextCheckinAt *time.Time              `json:"next_checkin_at"`
	}

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	parsedSlots, err := parseTimeSlots(body.TimeSlots)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid time slot: " + err.Error()})
		return
	}

	schedule, err := h.checkinScheduleService.Create(services.CreateScheduleInput{
		PatientID:     body.PatientID,
		Frequency:     body.Frequency,
		TimeSlots:     parsedSlots,
		Timezone:      body.Timezone,
		IsActive:      body.IsActive,
		NextCheckinAt: body.NextCheckinAt,
	})
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrScheduleExists):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case errors.Is(err, gorm.ErrRecordNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "patient not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create schedule: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, schedule)
}

func (h *CheckinScheduleHandler) GetSchedule(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid schedule id"})
		return
	}

	schedule, err := h.checkinScheduleService.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "schedule not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch schedule: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, schedule)
}

func (h *CheckinScheduleHandler) ListSchedules(c *gin.Context) {
	includeInactive := false
	if v := c.Query("include_inactive"); v != "" {
		parsed, err := parseBool(v)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid include_inactive value"})
			return
		}
		includeInactive = parsed
	}

	var patientID *uuid.UUID
	if raw := c.Query("patient_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid patient_id"})
			return
		}
		patientID = &id
	}

	schedules, err := h.checkinScheduleService.List(includeInactive, patientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list schedules: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": schedules})
}

func (h *CheckinScheduleHandler) UpdateSchedule(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid schedule id"})
		return
	}

	var body struct {
		Frequency     *enums.ScheduleFrequency `json:"frequency"`
		TimeSlots     *[]string                `json:"time_slots"`
		Timezone      *string                  `json:"timezone"`
		IsActive      *bool                    `json:"is_active"`
		NextCheckinAt *time.Time               `json:"next_checkin_at"`
	}

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var parsedSlots *[]time.Time
	if body.TimeSlots != nil {
		slots, err := parseTimeSlots(*body.TimeSlots)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid time slot: " + err.Error()})
			return
		}
		parsedSlots = &slots
	}

	schedule, err := h.checkinScheduleService.Update(id, services.UpdateScheduleInput{
		Frequency:     body.Frequency,
		TimeSlots:     parsedSlots,
		Timezone:      body.Timezone,
		IsActive:      body.IsActive,
		NextCheckinAt: body.NextCheckinAt,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "schedule not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update schedule: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, schedule)
}

func (h *CheckinScheduleHandler) DeleteSchedule(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid schedule id"})
		return
	}

	if err := h.checkinScheduleService.Delete(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "schedule not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete schedule: " + err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func parseTimeSlots(slots []string) ([]time.Time, error) {
	out := make([]time.Time, 0, len(slots))
	for _, s := range slots {
		t, err := time.Parse("15:04", s)
		if err != nil {
			// also allow HH:MM:SS
			t, err = time.Parse("15:04:05", s)
			if err != nil {
				return nil, err
			}
		}
		out = append(out, t)
	}
	return out, nil
}

func parseBool(val string) (bool, error) {
	switch val {
	case "true", "1", "TRUE", "True":
		return true, nil
	case "false", "0", "FALSE", "False":
		return false, nil
	default:
		return false, errors.New("invalid boolean")
	}
}
