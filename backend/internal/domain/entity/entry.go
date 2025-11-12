package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Entry represents a block of time that may still be running when EndedAt is zero.
type Entry struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	UserID      uuid.UUID  `gorm:"type:uuid;index;not null" json:"user_id"`
	ProjectID   *uuid.UUID `gorm:"type:uuid" json:"project_id,omitempty"`
	Title       string     `gorm:"size:120;not null" json:"title"`
	Notes       string     `gorm:"type:text" json:"notes"`
	StartedAt   time.Time  `gorm:"not null" json:"started_at"`
	EndedAt     *time.Time `json:"ended_at,omitempty"`
	DurationSec int64      `gorm:"not null;default:0" json:"duration_sec"`
	IsBreak     bool       `gorm:"not null;default:false" json:"is_break"`
	Ratio       float64    `gorm:"not null;default:1" json:"ratio"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Tags        []Tag      `gorm:"many2many:entry_tags;constraint:OnDelete:CASCADE;" json:"tags,omitempty"`
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
