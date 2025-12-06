package errs

import "errors"

var (
	ErrInvalidToken        = errors.New("invalid token format")
	ErrExpiredToken        = errors.New("token has expired")
	ErrActiveCheckinExists = errors.New("active checkin already exists for this patient")
	ErrNoActiveCheckin     = errors.New("no active checkin found for this patient")
	ErrCheckinNotActive    = errors.New("checkin is not active")
	ErrScheduleExists      = errors.New("checkin schedule already exists for this patient")
	ErrCheckinNotCompleted = errors.New("checkin is not completed yet")
	ErrMissingAlertFields  = errors.New("alert input missing required fields")
	ErrCheckinNotAnalyzed  = errors.New("checkin has not been analyzed yet")
)
