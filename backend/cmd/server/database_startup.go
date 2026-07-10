package main

import (
	"log"
	"time"

	"chronome/internal/adapter/infra/config"
	"chronome/internal/adapter/infra/database"

	"gorm.io/gorm"
)

const (
	databaseStartupAttempts = 12
	databaseStartupDelay    = 5 * time.Second
)

func openDatabaseWithRetry(cfg config.Config) (*gorm.DB, error) {
	var lastErr error
	for attempt := 1; attempt <= databaseStartupAttempts; attempt++ {
		db, err := database.Open(cfg)
		if err != nil {
			lastErr = err
		} else if err := database.Automigrate(db); err != nil {
			lastErr = err
		} else {
			return db, nil
		}

		if attempt == databaseStartupAttempts {
			break
		}
		log.Printf(
			"database startup attempt %d/%d failed: %v; retrying in %s",
			attempt,
			databaseStartupAttempts,
			lastErr,
			databaseStartupDelay,
		)
		time.Sleep(databaseStartupDelay)
	}
	return nil, lastErr
}
