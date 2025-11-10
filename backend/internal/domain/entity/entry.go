package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Entry represents a block of time that may still be running when EndedAt is zero.
type Entry struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserID      uuid.UUID  `gorm:"type:uuid;index;not null"`
	ProjectID   *uuid.UUID `gorm:"type:uuid"`
	Title       string     `gorm:"size:120;not null"`
	Notes       string     `gorm:"type:text"`
	StartedAt   time.Time  `gorm:"not null"`
	EndedAt     *time.Time `gorm:""`
	DurationSec int64      `gorm:"not null;default:0"`
	IsBreak     bool       `gorm:"not null;default:false"`
	Ratio       float64    `gorm:"not null;default:1"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Tags        []Tag `gorm:"many2many:entry_tags;constraint:OnDelete:CASCADE;" json:"tags,omitempty"`
}

func (e *Entry) Validate() error {
	if e.Title == "" {
		return errors.New("title is required")
	}
	if e.StartedAt.IsZero() {
		return errors.New("started_at is required")
	}
	if e.Ratio <= 0 {
		return errors.New("ratio must be positive")
	}
	return nil
}

// UpdateDuration recalculates duration using StartedAt/EndedAt.
func (e *Entry) UpdateDuration(now time.Time) {
	end := now
	if e.EndedAt != nil {
		end = *e.EndedAt
	}
	if end.After(e.StartedAt) {
		e.DurationSec = int64(end.Sub(e.StartedAt).Seconds())
	}
}
