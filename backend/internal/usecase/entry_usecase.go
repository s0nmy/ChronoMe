package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"chronome/internal/domain/entity"
	"chronome/internal/domain/repository"
	"chronome/internal/usecase/dto"
	"chronome/internal/usecase/provider"
)

// EntryUsecase は時間エントリ周りの業務処理を制御する。
type EntryUsecase struct {
	entries repository.EntryRepository
	tags    repository.TagRepository
	clock   provider.Clock
}

func NewEntryUsecase(entries repository.EntryRepository, tags repository.TagRepository, clock provider.Clock) *EntryUsecase {
	return &EntryUsecase{entries: entries, tags: tags, clock: clock}
}

func (u *EntryUsecase) Create(ctx context.Context, userID uuid.UUID, input dto.EntryCreateRequest) (*entity.Entry, error) {
	data, err := input.Normalize()
	if err != nil {
		return nil, err
	}
	started := u.clock.Now()
	if data.StartedAt != nil {
		started = data.StartedAt.UTC()
	}
	entry := &entity.Entry{
		ID:        uuid.New(),
		UserID:    userID,
		ProjectID: data.ProjectID,
		Title:     data.Title,
		Notes:     data.Notes,
		StartedAt: started,
		IsBreak:   data.IsBreak,
		Ratio:     data.Ratio,
	}
	if data.EndedAt != nil {
		end := data.EndedAt.UTC()
		entry.EndedAt = &end
	}
	tags, err := u.loadTags(ctx, userID, data.TagIDs)
	if err != nil {
		return nil, err
	}
	entry.Tags = tags
	if entry.EndedAt != nil {
		entry.UpdateDuration(entry.EndedAt.UTC())
	}
	if err := entry.Validate(); err != nil {
		return nil, err
	}
	if err := u.entries.Create(ctx, entry); err != nil {
		return nil, err
	}
	if len(tags) > 0 {
		if err := u.entries.ReplaceTags(ctx, entry, tagIDsFrom(tags)); err != nil {
			return nil, err
		}
	}
	return entry, nil
}

func (u *EntryUsecase) List(ctx context.Context, userID uuid.UUID, filter repository.EntryFilter) ([]entity.Entry, error) {
	return u.entries.ListByUser(ctx, userID, filter)
}

func (u *EntryUsecase) Update(ctx context.Context, userID uuid.UUID, id uuid.UUID, input dto.EntryUpdateRequest) (*entity.Entry, error) {
	updates, err := input.Normalize()
	if err != nil {
		return nil, err
	}
	entry, err := u.entries.GetByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if updates.Title != nil {
		entry.Title = *updates.Title
	}
	if updates.Notes != nil {
		entry.Notes = *updates.Notes
	}
	if updates.ProjectID != nil {
		entry.ProjectID = updates.ProjectID
	}
	if updates.StartedAt != nil {
		entry.StartedAt = updates.StartedAt.UTC()
	}
	if updates.EndedAtSet {
		if updates.EndedAt == nil {
			entry.EndedAt = nil
		} else {
			end := updates.EndedAt.UTC()
			entry.EndedAt = &end
		}
	}
	if updates.IsBreak != nil {
		entry.IsBreak = *updates.IsBreak
	}
	if updates.Ratio != nil {
		entry.Ratio = *updates.Ratio
	}
	var tags []entity.Tag
	if updates.TagIDsSet {
		tags, err = u.loadTags(ctx, userID, updates.TagIDs)
		if err != nil {
			return nil, err
		}
		entry.Tags = tags
	}
	entry.UpdateDuration(u.clock.Now())
	if err := entry.Validate(); err != nil {
		return nil, err
	}
	if err := u.entries.Update(ctx, entry); err != nil {
		return nil, err
	}
	if updates.TagIDsSet {
		if err := u.entries.ReplaceTags(ctx, entry, tagIDsFrom(tags)); err != nil {
			return nil, err
		}
	}
	return entry, nil
}

func (u *EntryUsecase) Delete(ctx context.Context, userID uuid.UUID, id uuid.UUID) error {
	if id == uuid.Nil {
		return errors.New("id is required")
	}
	return u.entries.Delete(ctx, userID, id)
}

func (u *EntryUsecase) loadTags(ctx context.Context, userID uuid.UUID, tagIDs []uuid.UUID) ([]entity.Tag, error) {
	if len(tagIDs) == 0 {
		return nil, nil
	}
	if u.tags == nil {
		return nil, errors.New("tag repository is not configured")
	}
	seen := make(map[uuid.UUID]struct{}, len(tagIDs))
	validated := make([]entity.Tag, 0, len(tagIDs))
	for _, id := range tagIDs {
		if id == uuid.Nil {
			return nil, dto.ValidationError{Field: "tag_ids", Message: "contains invalid UUID"}
		}
		if _, ok := seen[id]; ok {
			continue
		}
		tag, err := u.tags.GetByID(ctx, userID, id)
		if err != nil {
			return nil, dto.ValidationError{Field: "tag_ids", Message: "contains unknown tag"}
		}
		seen[id] = struct{}{}
		validated = append(validated, *tag)
	}
	return validated, nil
}

func tagIDsFrom(tags []entity.Tag) []uuid.UUID {
	if len(tags) == 0 {
		return nil
	}
	result := make([]uuid.UUID, len(tags))
	for i := range tags {
		result[i] = tags[i].ID
	}
	return result
}
