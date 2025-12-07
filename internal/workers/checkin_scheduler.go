package workers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"time"

	"github.com/erkinov-wtf/vital-sync/internal/api/services"
	"github.com/erkinov-wtf/vital-sync/internal/enums"
	"github.com/erkinov-wtf/vital-sync/internal/models"
	"github.com/erkinov-wtf/vital-sync/internal/pkg/errs"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CheckinScheduler struct {
	db           *gorm.DB
	logger       *slog.Logger
	checkinSvc   *services.CheckinService
	defaultTZ    string
	pollInterval time.Duration
}

func NewCheckinScheduler(db *gorm.DB, logger *slog.Logger, checkinSvc *services.CheckinService, defaultTZ string) *CheckinScheduler {
	return &CheckinScheduler{
		db:           db,
		logger:       logger,
		checkinSvc:   checkinSvc,
		defaultTZ:    defaultTZ,
		pollInterval: time.Minute,
	}
}

func (s *CheckinScheduler) Start(ctx context.Context) {
	s.logger.Info("starting checkin scheduler", "interval", s.pollInterval.String())
	go s.run(ctx)
}

func (s *CheckinScheduler) run(ctx context.Context) {
	s.processTick(time.Now())

	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			s.processTick(now)
		}
	}
}

func (s *CheckinScheduler) processTick(now time.Time) {
	var schedules []models.CheckinSchedule
	if err := s.db.Where("is_active = ?", true).Find(&schedules).Error; err != nil {
		s.logger.Error("failed to load checkin schedules", "error", err)
		return
	}

	for _, schedule := range schedules {
		if err := s.handleSchedule(schedule, now); err != nil {
			s.logger.Error("failed to process checkin schedule", "schedule_id", schedule.ID, "error", err)
		}
	}
}

func (s *CheckinScheduler) handleSchedule(schedule models.CheckinSchedule, tickTime time.Time) error {
	loc, err := s.loadLocation(schedule.Timezone)
	if err != nil {
		return fmt.Errorf("load timezone %q: %w", schedule.Timezone, err)
	}

	now := tickTime.In(loc)
	nextAt := schedule.NextCheckinAt
	if nextAt == nil {
		calculated, err := s.computeNextCheckinAt(schedule, now)
		if err != nil {
			return fmt.Errorf("compute next checkin time: %w", err)
		}
		nextAt = calculated
		if err := s.updateNextCheckinAt(schedule.ID, calculated); err != nil {
			return fmt.Errorf("set initial next_checkin_at: %w", err)
		}
		s.logger.Info("initialized next_checkin_at", "schedule_id", schedule.ID, "patient_id", schedule.PatientID, "next_at", calculated)
	}

	if nextAt == nil {
		return nil
	}

	if nextAt.In(loc).After(now) {
		s.logger.Debug("schedule not due yet", "schedule_id", schedule.ID, "patient_id", schedule.PatientID, "next_at", nextAt.In(loc))
		return nil
	}

	patientUserID, err := s.getPatientUserID(schedule.PatientID)
	if err != nil {
		return fmt.Errorf("get patient user id: %w", err)
	}

	if _, err := s.checkinSvc.GetActiveCheckin(patientUserID); err == nil {
		s.logger.Info("active checkin already in progress, skipping scheduled start", "patient_id", patientUserID, "schedule_id", schedule.ID)
	} else if errors.Is(err, errs.ErrNoActiveCheckin) {
		checkin, err := s.checkinSvc.StartManualCheckin(patientUserID)
		if err != nil {
			return fmt.Errorf("start manual checkin: %w", err)
		}
		s.logger.Info("scheduled checkin started", "checkin_id", checkin.ID, "patient_id", patientUserID, "schedule_id", schedule.ID)
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("patient not found for schedule: %w", err)
	} else {
		return fmt.Errorf("check active checkin: %w", err)
	}

	nextAfter, err := s.computeNextCheckinAt(schedule, now.Add(time.Second))
	if err != nil {
		return fmt.Errorf("compute next checkin after run: %w", err)
	}

	if err := s.updateNextCheckinAt(schedule.ID, nextAfter); err != nil {
		return fmt.Errorf("update next_checkin_at: %w", err)
	}
	s.logger.Info("scheduled next checkin", "schedule_id", schedule.ID, "patient_id", schedule.PatientID, "next_at", nextAfter)

	return nil
}

func (s *CheckinScheduler) getPatientUserID(patientID uuid.UUID) (uuid.UUID, error) {
	var patient models.Patient
	if err := s.db.First(&patient, "id = ?", patientID).Error; err != nil {
		return uuid.Nil, err
	}
	return patient.UserID, nil
}

func (s *CheckinScheduler) updateNextCheckinAt(scheduleID uuid.UUID, nextAt *time.Time) error {
	return s.db.Model(&models.CheckinSchedule{}).
		Where("id = ?", scheduleID).
		Update("next_checkin_at", nextAt).
		Error
}

func (s *CheckinScheduler) computeNextCheckinAt(schedule models.CheckinSchedule, from time.Time) (*time.Time, error) {
	if len(schedule.TimeSlots) == 0 {
		return nil, errors.New("no time slots configured")
	}

	loc, err := s.loadLocation(schedule.Timezone)
	if err != nil {
		return nil, err
	}

	slots := make([]time.Time, len(schedule.TimeSlots))
	copy(slots, schedule.TimeSlots)
	sort.Slice(slots, func(i, j int) bool {
		hi, mi, si := slots[i].Clock()
		hj, mj, sj := slots[j].Clock()
		if hi != hj {
			return hi < hj
		}
		if mi != mj {
			return mi < mj
		}
		return si < sj
	})

	dayStep := 1
	switch schedule.Frequency {
	case enums.ScheduleFrequencyEveryOtherDay:
		dayStep = 2
	case enums.ScheduleFrequencyWeekly:
		dayStep = 7
	}

	base := from.In(loc)
	maxDaysToCheck := 28

	for dayOffset := 0; dayOffset <= maxDaysToCheck; dayOffset += dayStep {
		day := base.AddDate(0, 0, dayOffset)

		for _, slot := range slots {
			candidate := time.Date(day.Year(), day.Month(), day.Day(), slot.Hour(), slot.Minute(), slot.Second(), 0, loc)
			if candidate.Before(base) {
				continue
			}
			return &candidate, nil
		}
	}

	return nil, errors.New("could not determine next checkin time")
}

func (s *CheckinScheduler) loadLocation(name string) (*time.Location, error) {
	if name == "" {
		name = s.defaultTZ
	}

	loc, err := time.LoadLocation(name)
	if err != nil {
		if s.defaultTZ != "" && name != s.defaultTZ {
			return time.LoadLocation(s.defaultTZ)
		}
		return nil, err
	}

	return loc, nil
}
