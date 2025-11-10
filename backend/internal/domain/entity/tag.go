package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Tag labels entries for fine-grained filtering.
type Tag struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;index;not null" json:"-"`
	Name      string    `gorm:"size:40;not null" json:"name"`
	Color     string    `gorm:"size:7;not null" json:"color"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (t *Tag) Validate() error {
	if t.Name == "" {
		return errors.New("tag name is required")
	}
	if len(t.Color) != 7 || t.Color[0] != '#' {
		return errors.New("color must be #RRGGBB")
	}
	return nil
}
