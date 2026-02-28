package gormrepo

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"chronome/internal/domain/entity"
	"chronome/internal/domain/repository"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared&_foreign_keys=1"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&entity.User{},
		&entity.Project{},
		&entity.Entry{},
		&entity.Tag{},
		&entity.EntryTag{},
		&entity.AllocationRequest{},
		&entity.TaskAllocation{},
	))
	require.NoError(t, db.Exec("PRAGMA foreign_keys = ON").Error)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})
	return db
}

func TestProjectRepository_CRUD(t *testing.T) {
	db := newTestDB(t)
	repo := NewProjectRepository(db)
	ctx := context.Background()
	userID := uuid.New()

	now := time.Now().UTC()
	first := &entity.Project{ID: uuid.New(), UserID: userID, Name: "Backend", Color: "#111111", CreatedAt: now.Add(-time.Minute)}
	second := &entity.Project{ID: uuid.New(), UserID: userID, Name: "Frontend", Color: "#222222", CreatedAt: now}
	require.NoError(t, repo.Create(ctx, first))
	require.NoError(t, repo.Create(ctx, second))

	projects, err := repo.ListByUser(ctx, userID)
	require.NoError(t, err)
	require.Len(t, projects, 2)
	require.Equal(t, second.ID, projects[0].ID, "expected most recent project first")

	loaded, err := repo.GetByID(ctx, userID, first.ID)
	require.NoError(t, err)
	require.Equal(t, "Backend", loaded.Name)

	loaded.Name = "Backend v2"
	require.NoError(t, repo.Update(ctx, loaded))

	updated, err := repo.GetByID(ctx, userID, first.ID)
	require.NoError(t, err)
	require.Equal(t, "Backend v2", updated.Name)

	require.NoError(t, repo.Delete(ctx, userID, second.ID))
	_, err = repo.GetByID(ctx, userID, second.ID)
	require.Error(t, err)
}

func TestEntryRepository_ListFilters(t *testing.T) {
	db := newTestDB(t)
	repo := NewEntryRepository(db)
	ctx := context.Background()
	userID := uuid.New()
	projectA := uuid.New()
	projectB := uuid.New()

	entries := []entity.Entry{
		{ID: uuid.New(), UserID: userID, ProjectID: &projectA, Title: "A", StartedAt: time.Now().Add(-3 * time.Hour), DurationSec: 600, Ratio: 1},
		{ID: uuid.New(), UserID: userID, ProjectID: &projectB, Title: "B", StartedAt: time.Now().Add(-2 * time.Hour), DurationSec: 1200, Ratio: 1},
		{ID: uuid.New(), UserID: userID, Title: "C", StartedAt: time.Now().Add(-30 * time.Minute), DurationSec: 300, Ratio: 1},
	}
	for i := range entries {
		require.NoError(t, repo.Create(ctx, &entries[i]))
	}

	from := time.Now().Add(-90 * time.Minute)
	to := time.Now()
	result, err := repo.ListByUser(ctx, userID, repository.EntryFilter{From: &from, To: &to})
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Equal(t, "C", result[0].Title)

	result, err = repo.ListByUser(ctx, userID, repository.EntryFilter{ProjectID: &projectA})
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Equal(t, "A", result[0].Title)
}

func TestEntryRepository_UserScoping(t *testing.T) {
	db := newTestDB(t)
	repo := NewEntryRepository(db)
	ctx := context.Background()
	userID := uuid.New()
	otherUser := uuid.New()

	entry := &entity.Entry{ID: uuid.New(), UserID: userID, Title: "Scoped", StartedAt: time.Now().Add(-time.Hour), Ratio: 1}
	require.NoError(t, repo.Create(ctx, entry))

	_, err := repo.GetByID(ctx, otherUser, entry.ID)
	require.Error(t, err, "other users should not see the entry")

	require.NoError(t, repo.Delete(ctx, userID, entry.ID))
	_, err = repo.GetByID(ctx, userID, entry.ID)
	require.Error(t, err)
}

func TestEntryRepository_TagAssociationsAndFiltering(t *testing.T) {
	db := newTestDB(t)
	repo := NewEntryRepository(db)
	tagRepo := NewTagRepository(db)
	ctx := context.Background()
	userID := uuid.New()

	tagA := &entity.Tag{ID: uuid.New(), UserID: userID, Name: "Focus", Color: "#111111"}
	tagB := &entity.Tag{ID: uuid.New(), UserID: userID, Name: "Meeting", Color: "#222222"}
	require.NoError(t, tagRepo.Create(ctx, tagA))
	require.NoError(t, tagRepo.Create(ctx, tagB))

	entry := &entity.Entry{ID: uuid.New(), UserID: userID, Title: "Scoped", StartedAt: time.Now().Add(-time.Hour), Ratio: 1}
	require.NoError(t, repo.Create(ctx, entry))
	require.NoError(t, repo.ReplaceTags(ctx, entry, []uuid.UUID{tagA.ID}))

	loaded, err := repo.GetByID(ctx, userID, entry.ID)
	require.NoError(t, err)
	require.Len(t, loaded.Tags, 1)
	require.Equal(t, tagA.ID, loaded.Tags[0].ID)

	require.NoError(t, repo.ReplaceTags(ctx, entry, []uuid.UUID{tagB.ID}))

	result, err := repo.ListByUser(ctx, userID, repository.EntryFilter{TagID: &tagB.ID})
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Equal(t, entry.ID, result[0].ID)
	require.Len(t, result[0].Tags, 1)
	require.Equal(t, tagB.ID, result[0].Tags[0].ID)
}

func TestTagRepository_CreateAndList(t *testing.T) {
	db := newTestDB(t)
	repo := NewTagRepository(db)
	ctx := context.Background()
	userID := uuid.New()

	tag := &entity.Tag{ID: uuid.New(), UserID: userID, Name: "Deep Work", Color: "#F97316"}
	require.NoError(t, repo.Create(ctx, tag))

	list, err := repo.ListByUser(ctx, userID)
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.Equal(t, "Deep Work", list[0].Name)

	found, err := repo.GetByID(ctx, userID, tag.ID)
	require.NoError(t, err)
	require.Equal(t, tag.ID, found.ID)

	found.Name = "Shallow"
	require.NoError(t, repo.Update(ctx, found))

	updated, err := repo.GetByID(ctx, userID, tag.ID)
	require.NoError(t, err)
	require.Equal(t, "Shallow", updated.Name)

	require.NoError(t, repo.Delete(ctx, userID, tag.ID))
	_, err = repo.GetByID(ctx, userID, tag.ID)
	require.Error(t, err)
}

func TestUserRepository_NormalizesOnCreate(t *testing.T) {
	db := newTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	rawEmail := "USER@Example.Com"
	user := &entity.User{ID: uuid.New(), Email: rawEmail, PasswordHash: "secret"}
	require.NoError(t, repo.Create(ctx, user))
	require.Equal(t, "user@example.com", user.Email)

	found, err := repo.GetByEmail(ctx, rawEmail)
	require.NoError(t, err)
	require.Equal(t, user.ID, found.ID)

	byID, err := repo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, "user@example.com", byID.Email)
}

func TestAllocationRepository_Create(t *testing.T) {
	db := newTestDB(t)
	repo := NewAllocationRepository(db)
	ctx := context.Background()

	request := &entity.AllocationRequest{ID: uuid.New(), TotalMinutes: 120, CreatedAt: time.Now().UTC()}
	allocations := []entity.TaskAllocation{
		{RequestID: request.ID, TaskID: "task-a", Ratio: 0.6, AllocatedMinutes: 72, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()},
		{RequestID: request.ID, TaskID: "task-b", Ratio: 0.4, AllocatedMinutes: 48, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()},
	}

	require.NoError(t, repo.Create(ctx, request, allocations))

	var requestCount int64
	require.NoError(t, db.Model(&entity.AllocationRequest{}).Where("id = ?", request.ID).Count(&requestCount).Error)
	require.Equal(t, int64(1), requestCount)

	var allocationCount int64
	require.NoError(t, db.Model(&entity.TaskAllocation{}).Where("request_id = ?", request.ID).Count(&allocationCount).Error)
	require.Equal(t, int64(2), allocationCount)
}

func TestAllocationRepository_Create_RollsBackOnFailure(t *testing.T) {
	db := newTestDB(t)
	repo := NewAllocationRepository(db)
	ctx := context.Background()

	request := &entity.AllocationRequest{ID: uuid.New(), TotalMinutes: 60, CreatedAt: time.Now().UTC()}
	allocations := []entity.TaskAllocation{
		{RequestID: uuid.New(), TaskID: "task-a", Ratio: 1, AllocatedMinutes: 60, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()},
	}

	require.Error(t, repo.Create(ctx, request, allocations))

	var requestCount int64
	require.NoError(t, db.Model(&entity.AllocationRequest{}).Where("id = ?", request.ID).Count(&requestCount).Error)
	require.Equal(t, int64(0), requestCount)

	var allocationCount int64
	require.NoError(t, db.Model(&entity.TaskAllocation{}).Count(&allocationCount).Error)
	require.Equal(t, int64(0), allocationCount)
}
