package config

import (
	"log"
	"sync"
	"time"
)

var (
	Timezone     *time.Location
	timezoneOnce sync.Once
)

// setTimezone sets timezone from config and set db timezone from config
func setTimezone(cfg *Config) *time.Location {
	timezoneOnce.Do(func() {
		var err error
		Timezone, err = time.LoadLocation(cfg.Timezone)
		if err != nil {
			log.Printf("Failed to initialize timezone: %v", err)
		} else {
			log.Print("Timezone set successfully")
		}
	})

	time.Local = Timezone
	cfg.Internal.Database.Timezone = cfg.Timezone

	return Timezone
}
