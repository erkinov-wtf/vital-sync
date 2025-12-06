package handlers

import (
	"errors"
	"net/http"

	"github.com/erkinov-wtf/vital-sync/internal/api/services"
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
