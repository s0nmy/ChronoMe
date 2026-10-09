package gormrepo

import (
	"context"
	"slices"
	"time"

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
	// すべての検索は user_id で絞り、アプリ層からの取り違えでも他ユーザーのデータを返さない。
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
		// タグ絞り込みは many-to-many の中間テーブル entry_tags を JOIN する。
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

func (r *EntryRepository) Update(ctx context.Context, entry *entity.Entry, columns []string) error {
	now := time.Now().UTC()
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&entity.Entry{}).Where("id = ? AND user_id = ? AND version = ?", entry.ID, entry.UserID, entry.Version).
			Select(append(append([]string{}, columns...), "version", "updated_at")).Omit("Tags").Updates(updateValues(map[string]any{"title": entry.Title, "notes": entry.Notes, "project_id": entry.ProjectID, "started_at": entry.StartedAt, "ended_at": entry.EndedAt, "duration_sec": entry.DurationSec, "is_break": entry.IsBreak, "ratio": entry.Ratio}, columns, now))
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return repository.ErrConflict
		}

		if slices.Contains(columns, "Tags") {
			// Only replace join rows; never save tag entities from a stale snapshot.
			if err := tx.Where("entry_id = ?", entry.ID).Delete(&entity.EntryTag{}).Error; err != nil {
				return err
			}
			links := make([]entity.EntryTag, 0, len(entry.Tags))
			for _, tag := range entry.Tags {
				links = append(links, entity.EntryTag{EntryID: entry.ID, TagID: tag.ID})
			}
			if len(links) > 0 {
				if err := tx.Create(&links).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err == nil {
		entry.Version++
		entry.UpdatedAt = now
	}
	return err
}

func (r *EntryRepository) Delete(ctx context.Context, userID uuid.UUID, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("user_id = ? AND id = ?", userID, id).Delete(&entity.Entry{}).Error
}

func (r *EntryRepository) ReplaceTags(ctx context.Context, entry *entity.Entry, tagIDs []uuid.UUID) error {
	db := r.db.WithContext(ctx)
	assoc := db.Model(entry).Association("Tags")
	if len(tagIDs) == 0 {
		// 空配列は「タグをすべて外す」という明示的な更新として扱う。
		return assoc.Clear()
	}
	tags := make([]entity.Tag, len(tagIDs))
	for i, id := range tagIDs {
		tags[i] = entity.Tag{ID: id}
	}
	return assoc.Replace(tags)
}
