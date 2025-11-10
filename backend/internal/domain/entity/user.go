package entity

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// User represents an account that owns all other resources.
type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	Email        string    `gorm:"uniqueIndex;size:254;not null"`
	PasswordHash string    `gorm:"not null"`
	DisplayName  string    `gorm:"size:50"`
	TimeZone     string    `gorm:"size:40;default:UTC"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Normalize prepares the entity before persistence.
func (u *User) Normalize() {
	u.Email = strings.ToLower(strings.TrimSpace(u.Email))
	if u.TimeZone == "" {
		u.TimeZone = "UTC"
	}
}

// Validate performs minimal server side checks.
func (u *User) Validate() error {
	if u.Email == "" {
		return errors.New("email is required")
	}
	if len(u.PasswordHash) == 0 {
		return errors.New("password hash is required")
	}
	if len(u.DisplayName) > 50 {
		return errors.New("display name is too long")
	}
	if u.TimeZone == "" {
		return errors.New("time zone is required")
	}
	return nil
}
