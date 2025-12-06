package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/erkinov-wtf/vital-sync/internal/api/services"
	"github.com/erkinov-wtf/vital-sync/internal/enums"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type VitalReadingHandler struct {
	vitalReadingService *services.VitalReadingService
}

func NewVitalReadingHandler(service *services.VitalReadingService) *VitalReadingHandler {
	return &VitalReadingHandler{vitalReadingService: service}
}

func (h *VitalReadingHandler) Create(c *gin.Context) {
	var body struct {
		CheckinID             uuid.UUID        `json:"checkin_id" binding:"required"`
		PatientID             uuid.UUID        `json:"patient_id" binding:"required"`
		VitalType             enums.VitalType  `json:"vital_type" binding:"required"`
		Unit                  *enums.VitalUnit `json:"unit"`
		ValueNumeric          *float64         `json:"value_numeric"`
		ValueText             *string          `json:"value_text"`
		IsAbnormal            *bool            `json:"is_abnormal"`
		DeviationFromBaseline *float64         `json:"deviation_from_baseline"`
	}

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	reading, err := h.vitalReadingService.Create(services.CreateVitalReadingInput{
		CheckinID:             body.CheckinID,
		PatientID:             body.PatientID,
		VitalType:             body.VitalType,
		Unit:                  body.Unit,
		ValueNumeric:          body.ValueNumeric,
		ValueText:             body.ValueText,
		IsAbnormal:            body.IsAbnormal,
		DeviationFromBaseline: body.DeviationFromBaseline,
	})
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "patient or checkin not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create vital reading: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, reading)
}

func (h *VitalReadingHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	reading, err := h.vitalReadingService.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "vital reading not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch vital reading: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, reading)
}

func (h *VitalReadingHandler) List(c *gin.Context) {
	var patientID *uuid.UUID
	if raw := c.Query("patient_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid patient_id"})
			return
		}
		patientID = &id
	}

	var checkinID *uuid.UUID
	if raw := c.Query("checkin_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid checkin_id"})
			return
		}
		checkinID = &id
	}

	var vitalType *enums.VitalType
	if raw := c.Query("vital_type"); raw != "" {
		vt := enums.VitalType(raw)
		vitalType = &vt
	}

	onlyAbnormal := false
	if raw := c.Query("only_abnormal"); raw != "" {
		val, err := strconv.ParseBool(raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid only_abnormal value"})
			return
		}
		onlyAbnormal = val
	}

	readings, err := h.vitalReadingService.List(patientID, checkinID, vitalType, onlyAbnormal)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list vital readings: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": readings})
}

func (h *VitalReadingHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var body struct {
		VitalType             *enums.VitalType `json:"vital_type"`
		Unit                  *enums.VitalUnit `json:"unit"`
		ValueNumeric          *float64         `json:"value_numeric"`
		ValueText             *string          `json:"value_text"`
		IsAbnormal            *bool            `json:"is_abnormal"`
		DeviationFromBaseline *float64         `json:"deviation_from_baseline"`
	}

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	reading, err := h.vitalReadingService.Update(id, services.UpdateVitalReadingInput{
		VitalType:             body.VitalType,
		Unit:                  body.Unit,
		ValueNumeric:          body.ValueNumeric,
		ValueText:             body.ValueText,
		IsAbnormal:            body.IsAbnormal,
		DeviationFromBaseline: body.DeviationFromBaseline,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "vital reading not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update vital reading: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, reading)
}

func (h *VitalReadingHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.vitalReadingService.Delete(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "vital reading not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete vital reading: " + err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
