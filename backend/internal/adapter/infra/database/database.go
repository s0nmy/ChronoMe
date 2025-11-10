package database

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"chronome/internal/adapter/infra/config"
	"chronome/internal/domain/entity"
)

// Open establishes a gorm.DB based on config.
func Open(cfg config.Config) (*gorm.DB, error) {
	var (
		db  *gorm.DB
		err error
	)
	gormConfig := &gorm.Config{Logger: logger.Default.LogMode(logger.Warn)}
	switch cfg.DBDriver {
	case "postgres":
		db, err = gorm.Open(postgres.Open(cfg.DBDsn), gormConfig)
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(cfg.DBDsn), gormConfig)
	default:
		err = fmt.Errorf("unsupported db driver: %s", cfg.DBDriver)
	}
	if err != nil {
		return nil, err
	}
	return db, nil
}

// Automigrate ensures database schema exists for the core entities.
func Automigrate(db *gorm.DB) error {
	return db.AutoMigrate(&entity.User{}, &entity.Project{}, &entity.Entry{}, &entity.Tag{}, &entity.EntryTag{})
}
