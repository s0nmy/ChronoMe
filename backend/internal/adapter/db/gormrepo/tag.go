package gormrepo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"chronome/internal/domain/entity"
	"chronome/internal/domain/repository"
)

// TagRepository は repository.TagRepository を実装する。
type TagRepository struct {
	db *gorm.DB
}

func NewTagRepository(db *gorm.DB) *TagRepository {
	return &TagRepository{db: db}
}

func (r *TagRepository) Create(ctx context.Context, tag *entity.Tag) error {
	return r.db.WithContext(ctx).Create(tag).Error
}

func (r *TagRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]entity.Tag, error) {
	var tags []entity.Tag
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at desc").Find(&tags).Error; err != nil {
		return nil, err
	}
	return tags, nil
}

func (r *TagRepository) GetByID(ctx context.Context, userID uuid.UUID, id uuid.UUID) (*entity.Tag, error) {
	var tag entity.Tag
	if err := r.db.WithContext(ctx).Where("user_id = ? AND id = ?", userID, id).First(&tag).Error; err != nil {
		return nil, err
	}
	return &tag, nil
}

func (r *TagRepository) Update(ctx context.Context, tag *entity.Tag, columns []string) error {
	now := time.Now().UTC()
	fields := map[string]any{
		"name":  tag.Name,
		"color": tag.Color,
	}
	values := updateValues(fields, columns, now)
	selectedColumns := append([]string{}, columns...)
	selectedColumns = append(selectedColumns, "version", "updated_at")

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&entity.Tag{}).
			Where("id = ? AND user_id = ? AND version = ?", tag.ID, tag.UserID, tag.Version).
			Select(selectedColumns).
			Omit("Tags").
			Updates(values)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return repository.ErrConflict
		}
		return nil
	})
	if err == nil {
		tag.Version++
		tag.UpdatedAt = now
	}
	return err
}

func (r *TagRepository) Delete(ctx context.Context, userID uuid.UUID, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("user_id = ? AND id = ?", userID, id).Delete(&entity.Tag{}).Error
}
