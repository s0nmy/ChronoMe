package gormrepo

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"chronome/internal/domain/entity"
	"chronome/internal/domain/repository"
)

// EntryRepository は GORM で repository.EntryRepository を実装する。
type EntryRepository struct {
	db *gorm.DB
}

func NewEntryRepository(db *gorm.DB) *EntryRepository {
	return &EntryRepository{db: db}
}

func (r *EntryRepository) Create(ctx context.Context, entry *entity.Entry) error {
	return r.db.WithContext(ctx).Create(entry).Error
}

func (r *EntryRepository) ListByUser(ctx context.Context, userID uuid.UUID, filter repository.EntryFilter) ([]entity.Entry, error) {
	query := r.db.WithContext(ctx).Model(&entity.Entry{}).Preload("Tags").Where("user_id = ?", userID)
	if filter.From != nil {
		query = query.Where("started_at >= ?", filter.From)
	}
	if filter.To != nil {
		query = query.Where("started_at < ?", filter.To)
	}
	if filter.ProjectID != nil {
		query = query.Where("project_id = ?", filter.ProjectID)
	}
	if filter.TagID != nil {
		query = query.Joins("JOIN entry_tags ON entry_tags.entry_id = entries.id").
			Where("entry_tags.tag_id = ?", *filter.TagID)
	}
	var entries []entity.Entry
	if err := query.Order("started_at desc").Find(&entries).Error; err != nil {
		return nil, err
	}
	return entries, nil
}

func (r *EntryRepository) GetByID(ctx context.Context, userID uuid.UUID, id uuid.UUID) (*entity.Entry, error) {
	var entry entity.Entry
	err := r.db.WithContext(ctx).Preload("Tags").Where("user_id = ? AND id = ?", userID, id).First(&entry).Error
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

func (r *EntryRepository) Update(ctx context.Context, entry *entity.Entry) error {
	return r.db.WithContext(ctx).Save(entry).Error
}

func (r *EntryRepository) Delete(ctx context.Context, userID uuid.UUID, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("user_id = ? AND id = ?", userID, id).Delete(&entity.Entry{}).Error
}

func (r *EntryRepository) ReplaceTags(ctx context.Context, entry *entity.Entry, tagIDs []uuid.UUID) error {
	db := r.db.WithContext(ctx)
	assoc := db.Model(entry).Association("Tags")
	if len(tagIDs) == 0 {
		return assoc.Clear()
	}
	tags := make([]entity.Tag, len(tagIDs))
	for i, id := range tagIDs {
		tags[i] = entity.Tag{ID: id}
	}
	return assoc.Replace(tags)
}
