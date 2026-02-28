package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Project はレポート用にエントリをまとめる。
type Project struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID      uuid.UUID `gorm:"type:uuid;index;not null" json:"user_id"`
	Name        string    `gorm:"size:80;not null" json:"name"`
	Description string    `gorm:"size:255" json:"description"`
	Color       string    `gorm:"size:7;not null" json:"color"`
	IsArchived  bool      `gorm:"not null;default:false" json:"is_archived"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
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
	if len(p.Description) > 255 {
		return errors.New("description is too long")
	}
	return nil
}
