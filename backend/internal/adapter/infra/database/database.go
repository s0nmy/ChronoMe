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

// Open は設定に基づいて gorm.DB を作成する。
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

// Automigrate は主要エンティティのスキーマが存在することを保証する。
func Automigrate(db *gorm.DB) error {
	return db.AutoMigrate(&entity.User{}, &entity.Project{}, &entity.Entry{}, &entity.Tag{}, &entity.EntryTag{})
}
