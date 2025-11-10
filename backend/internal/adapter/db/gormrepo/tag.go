package gormrepo

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"chronome/internal/domain/entity"
)

// TagRepository implements repository.TagRepository.
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

func (r *TagRepository) Update(ctx context.Context, tag *entity.Tag) error {
	return r.db.WithContext(ctx).Save(tag).Error
}

func (r *TagRepository) Delete(ctx context.Context, userID uuid.UUID, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("user_id = ? AND id = ?", userID, id).Delete(&entity.Tag{}).Error
}
