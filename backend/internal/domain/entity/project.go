package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Project groups entries for reporting purposes.
type Project struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID     uuid.UUID `gorm:"type:uuid;index;not null"`
	Name       string    `gorm:"size:80;not null"`
	Color      string    `gorm:"size:7;not null"`
	IsArchived bool      `gorm:"not null;default:false"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (p *Project) Validate() error {
	if p.Name == "" {
		return errors.New("project name is required")
	}
	if len(p.Name) > 80 {
		return errors.New("project name too long")
	}
	if len(p.Color) != 7 || p.Color[0] != '#' {
		return errors.New("color must be #RRGGBB")
	}
	return nil
}
