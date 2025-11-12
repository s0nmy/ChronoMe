package entity

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// User represents an account that owns all other resources.
type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Email        string    `gorm:"uniqueIndex;size:254;not null" json:"email"`
	PasswordHash string    `gorm:"not null" json:"-"`
	DisplayName  string    `gorm:"size:50" json:"display_name"`
	TimeZone     string    `gorm:"size:40;default:UTC" json:"time_zone"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
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
