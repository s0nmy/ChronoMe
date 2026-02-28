package gormrepo

import (
	"context"

	"gorm.io/gorm"

	"chronome/internal/domain/entity"
)

// AllocationRepository は分配履歴を保存する。
type AllocationRepository struct {
	db *gorm.DB
}

func NewAllocationRepository(db *gorm.DB) *AllocationRepository {
	return &AllocationRepository{db: db}
}

func (r *AllocationRepository) Create(ctx context.Context, request *entity.AllocationRequest, allocations []entity.TaskAllocation) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(request).Error; err != nil {
			return err
		}
		if len(allocations) == 0 {
			return nil
		}
		return tx.Create(&allocations).Error
	})
}
