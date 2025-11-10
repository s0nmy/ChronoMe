package entity

import (
	"time"

	"github.com/google/uuid"
)

// EntryTag models the association table between entries and tags.
type EntryTag struct {
	EntryID   uuid.UUID `gorm:"type:uuid;primaryKey"`
	TagID     uuid.UUID `gorm:"type:uuid;primaryKey"`
	CreatedAt time.Time
}

func (EntryTag) TableName() string {
	return "entry_tags"
}
