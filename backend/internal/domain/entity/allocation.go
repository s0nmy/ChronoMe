package entity

import (
	"time"

	"github.com/google/uuid"
)

// AllocationRequest は分配リクエストの履歴を保持する。
type AllocationRequest struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	TotalMinutes int       `gorm:"not null" json:"total_minutes"`
	CreatedAt    time.Time `json:"created_at"`
}

func (AllocationRequest) TableName() string {
	return "allocation_requests"
}

// TaskAllocation はタスクごとの分配結果を保持する。
type TaskAllocation struct {
	ID               uint              `gorm:"primaryKey;autoIncrement" json:"id"`
	RequestID        uuid.UUID         `gorm:"type:uuid;index;not null" json:"request_id"`
	Request          AllocationRequest `gorm:"foreignKey:RequestID;references:ID;constraint:OnDelete:CASCADE;" json:"-"`
	TaskID           string            `gorm:"size:120;not null" json:"task_id"`
	Ratio            float64           `gorm:"not null" json:"ratio"`
	AllocatedMinutes int               `gorm:"not null" json:"allocated_minutes"`
	MinMinutes       *int              `json:"min_minutes,omitempty"`
	MaxMinutes       *int              `json:"max_minutes,omitempty"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
}

func (TaskAllocation) TableName() string {
	return "task_allocations"
}
