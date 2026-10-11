package gormrepo

import (
	"chronome/internal/domain/entity"
	"chronome/internal/domain/repository"
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestEntryUpdateConflictAndPartialColumns(t *testing.T) {
	db := newTestDB(t)
	repo := NewEntryRepository(db)
	ctx := context.Background()
	entry := &entity.Entry{ID: uuid.New(), UserID: uuid.New(), Title: "original", Notes: "keep", StartedAt: time.Now().Add(-time.Hour), Ratio: 1}
	require.NoError(t, repo.Create(ctx, entry))
	first, err := repo.GetByID(ctx, entry.UserID, entry.ID)
	require.NoError(t, err)
	second, err := repo.GetByID(ctx, entry.UserID, entry.ID)
	require.NoError(t, err)
	first.Title = "renamed"
	first.Notes = "must not persist"
	require.NoError(t, repo.Update(ctx, first, []string{"title"}))
	end := time.Now()
	second.EndedAt = &end
	second.UpdateDuration(end)
	require.ErrorIs(t, repo.Update(ctx, second, []string{"ended_at", "duration_sec"}), repository.ErrConflict)
	current, err := repo.GetByID(ctx, entry.UserID, entry.ID)
	require.NoError(t, err)
	require.Equal(t, "renamed", current.Title)
	require.Equal(t, "keep", current.Notes)
	require.Nil(t, current.EndedAt)
	require.Equal(t, entry.Version+1, current.Version)
	require.Equal(t, first.UpdatedAt.UnixNano(), current.UpdatedAt.UnixNano())
	current.EndedAt = &end
	current.UpdateDuration(end)
	require.NoError(t, repo.Update(ctx, current, []string{"ended_at", "duration_sec"}))
	current.Notes = ""
	current.IsBreak = false
	current.EndedAt = nil
	require.NoError(t, repo.Update(ctx, current, []string{"notes", "is_break", "ended_at"}))
	require.NoError(t, repo.Delete(ctx, current.UserID, current.ID))
	require.ErrorIs(t, repo.Update(ctx, current, []string{"title"}), repository.ErrConflict)
	var count int64
	require.NoError(t, db.Model(&entity.Entry{}).Where("id = ?", current.ID).Count(&count).Error)
	require.Zero(t, count)
}

func TestProjectAndTagUpdateConflict(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	user := uuid.New()
	pr := NewProjectRepository(db)
	p := &entity.Project{ID: uuid.New(), UserID: user, Name: "name", Color: "#111111", Description: "keep", IsArchived: true}
	require.NoError(t, pr.Create(ctx, p))
	stale := *p
	p.Name = "new"
	p.Description = "ignored"
	require.NoError(t, pr.Update(ctx, p, []string{"name"}))
	stale.Color = "#222222"
	require.ErrorIs(t, pr.Update(ctx, &stale, []string{"color"}), repository.ErrConflict)
	got, err := pr.GetByID(ctx, user, p.ID)
	require.NoError(t, err)
	require.Equal(t, "new", got.Name)
	require.Equal(t, "keep", got.Description)
	got.IsArchived = false
	got.Description = ""
	require.NoError(t, pr.Update(ctx, got, []string{"is_archived", "description"}))
	got, err = pr.GetByID(ctx, user, p.ID)
	require.NoError(t, err)
	require.False(t, got.IsArchived)
	require.Empty(t, got.Description)
	tr := NewTagRepository(db)
	tag := &entity.Tag{ID: uuid.New(), UserID: user, Name: "name", Color: "#111111"}
	require.NoError(t, tr.Create(ctx, tag))
	old := *tag
	tag.Name = "new"
	tag.Color = "#333333"
	require.NoError(t, tr.Update(ctx, tag, []string{"name"}))
	require.ErrorIs(t, tr.Update(ctx, &old, []string{"color"}), repository.ErrConflict)
	loaded, err := tr.GetByID(ctx, user, tag.ID)
	require.NoError(t, err)
	require.Equal(t, "new", loaded.Name)
	require.Equal(t, "#111111", loaded.Color)
	loaded.UserID = uuid.New()
	require.ErrorIs(t, tr.Update(ctx, loaded, []string{"name"}), repository.ErrConflict)
}

func TestEntryTagsAtomicUpdate(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	repo := NewEntryRepository(db)
	tag := entity.Tag{ID: uuid.New(), UserID: uuid.New(), Name: "tag", Color: "#111111"}
	require.NoError(t, db.Create(&tag).Error)
	entry := &entity.Entry{ID: uuid.New(), UserID: tag.UserID, Title: "original", StartedAt: time.Now(), Ratio: 1}
	require.NoError(t, repo.Create(ctx, entry))
	stale := *entry
	entry.Tags = []entity.Tag{tag}
	require.NoError(t, repo.Update(ctx, entry, []string{"Tags"}))
	stale.Tags = nil
	require.ErrorIs(t, repo.Update(ctx, &stale, []string{"Tags"}), repository.ErrConflict)
	loaded, err := repo.GetByID(ctx, entry.UserID, entry.ID)
	require.NoError(t, err)
	require.Len(t, loaded.Tags, 1)
	loaded.Title = "must rollback"
	loaded.Tags = []entity.Tag{tag}
	require.NoError(t, db.Exec("CREATE TRIGGER fail_tag_insert BEFORE INSERT ON entry_tags BEGIN SELECT RAISE(ABORT, 'injected tag failure'); END").Error)
	require.Error(t, repo.Update(ctx, loaded, []string{"title", "Tags"}))
	current, err := repo.GetByID(ctx, entry.UserID, entry.ID)
	require.NoError(t, err)
	require.Equal(t, "original", current.Title)
	require.Equal(t, entry.Version, current.Version)
	require.Len(t, current.Tags, 1)
	require.NoError(t, db.Exec("DROP TRIGGER fail_tag_insert").Error)
	current.Tags = nil
	require.NoError(t, repo.Update(ctx, current, []string{"Tags"}))
	current, err = repo.GetByID(ctx, entry.UserID, entry.ID)
	require.NoError(t, err)
	require.Empty(t, current.Tags)
}
