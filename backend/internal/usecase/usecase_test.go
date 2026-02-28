package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"chronome/internal/domain/entity"
	"chronome/internal/domain/repository"
	"chronome/internal/usecase/dto"
	"chronome/internal/usecase/provider"
	"chronome/test/fakes"
)

func TestEntryUsecase_CreateSetsDefaults(t *testing.T) {
	ctx := context.Background()
	var captured *entity.Entry
	repo := &fakes.FakeEntryRepository{
		CreateFn: func(_ context.Context, entry *entity.Entry) error {
			captured = entry
			return nil
		},
	}
	now := time.Date(2024, 1, 2, 15, 4, 5, 0, time.UTC)
	clock := fakes.FixedTimeProvider{NowFunc: func() time.Time { return now }}
	uc := NewEntryUsecase(repo, &fakes.FakeTagRepository{}, clock)

	entry, err := uc.Create(ctx, uuid.New(), dto.EntryCreateRequest{Title: "Focus"})
	require.NoError(t, err)
	require.Equal(t, now, entry.StartedAt)
	require.Equal(t, 1.0, entry.Ratio)
	require.Equal(t, entry.ID, captured.ID)
}

func TestEntryUsecase_CreateValidatesTagOwnership(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	tagID := uuid.New()
	tagRepo := &fakes.FakeTagRepository{
		GetByIDFn: func(context.Context, uuid.UUID, uuid.UUID) (*entity.Tag, error) {
			return &entity.Tag{ID: tagID, UserID: userID, Name: "Valid", Color: "#111111"}, nil
		},
	}
	var replaced []uuid.UUID
	repo := &fakes.FakeEntryRepository{
		CreateFn: func(context.Context, *entity.Entry) error { return nil },
		ReplaceTagsFn: func(_ context.Context, _ *entity.Entry, tagIDs []uuid.UUID) error {
			replaced = tagIDs
			return nil
		},
	}
	uc := NewEntryUsecase(repo, tagRepo, fakes.FixedTimeProvider{})

	req := dto.EntryCreateRequest{Title: "Tagged", TagIDs: []string{tagID.String(), tagID.String()}}
	_, err := uc.Create(ctx, userID, req)
	require.NoError(t, err)
	require.Equal(t, []uuid.UUID{tagID}, replaced)
}

func TestEntryUsecase_CreateValidatesTitle(t *testing.T) {
	uc := NewEntryUsecase(&fakes.FakeEntryRepository{}, &fakes.FakeTagRepository{}, fakes.FixedTimeProvider{})
	_, err := uc.Create(context.Background(), uuid.New(), dto.EntryCreateRequest{})
	var valErr dto.ValidationError
	require.True(t, errors.As(err, &valErr))
}

func TestEntryUsecase_UpdateValidatesChangedFields(t *testing.T) {
	now := time.Now().Add(-time.Hour).UTC()
	existing := &entity.Entry{ID: uuid.New(), UserID: uuid.New(), Title: "Focus", StartedAt: now, Ratio: 1}
	repo := &fakes.FakeEntryRepository{
		GetByIDFn: func(context.Context, uuid.UUID, uuid.UUID) (*entity.Entry, error) {
			cloned := *existing
			return &cloned, nil
		},
	}
	uc := NewEntryUsecase(repo, &fakes.FakeTagRepository{}, fakes.FixedTimeProvider{NowFunc: func() time.Time { return now.Add(time.Hour) }})

	invalid := -2.0
	_, err := uc.Update(context.Background(), existing.UserID, existing.ID, dto.EntryUpdateRequest{Ratio: &invalid})
	var valErr dto.ValidationError
	require.True(t, errors.As(err, &valErr))
}

func TestEntryUsecase_UpdateRejectsUnknownTags(t *testing.T) {
	userID := uuid.New()
	entryID := uuid.New()
	repo := &fakes.FakeEntryRepository{
		GetByIDFn: func(context.Context, uuid.UUID, uuid.UUID) (*entity.Entry, error) {
			return &entity.Entry{ID: entryID, UserID: userID, Title: "Focus", StartedAt: time.Now().Add(-time.Hour), Ratio: 1}, nil
		},
	}
	tagRepo := &fakes.FakeTagRepository{
		GetByIDFn: func(context.Context, uuid.UUID, uuid.UUID) (*entity.Tag, error) {
			return nil, errors.New("not found")
		},
	}
	uc := NewEntryUsecase(repo, tagRepo, fakes.FixedTimeProvider{})
	ids := []string{uuid.NewString()}
	_, err := uc.Update(context.Background(), userID, entryID, dto.EntryUpdateRequest{TagIDs: &ids})
	var valErr dto.ValidationError
	require.True(t, errors.As(err, &valErr))
	require.Contains(t, valErr.Error(), "tag_ids")
}

func TestEntryUsecase_DeleteRequiresID(t *testing.T) {
	uc := NewEntryUsecase(&fakes.FakeEntryRepository{}, &fakes.FakeTagRepository{}, fakes.FixedTimeProvider{})
	err := uc.Delete(context.Background(), uuid.New(), uuid.Nil)
	require.EqualError(t, err, "id is required")
}

func TestProjectUsecase_CreateDefaultsColor(t *testing.T) {
	ctx := context.Background()
	var stored *entity.Project
	repo := &fakes.FakeProjectRepository{
		CreateFn: func(_ context.Context, project *entity.Project) error {
			stored = project
			return nil
		},
	}
	uc := NewProjectUsecase(repo, stubConfig{})

	project, err := uc.Create(ctx, uuid.New(), dto.ProjectCreateRequest{Name: "Chrono"})
	require.NoError(t, err)
	require.Equal(t, "#3B82F6", project.Color)
	require.Equal(t, project.ID, stored.ID)
}

func TestProjectUsecase_DeleteRequiresID(t *testing.T) {
	uc := NewProjectUsecase(&fakes.FakeProjectRepository{}, stubConfig{})
	err := uc.Delete(context.Background(), uuid.New(), uuid.Nil)
	require.EqualError(t, err, "id is required")
}

func TestAuthUsecase_SignupRejectsExistingUser(t *testing.T) {
	repo := &fakes.FakeUserRepository{
		GetByEmailFn: func(context.Context, string) (*entity.User, error) {
			return &entity.User{ID: uuid.New()}, nil
		},
	}
	uc := NewAuthUsecase(repo)

	_, err := uc.Signup(context.Background(), SignupParams{Email: "taken@example.com", Password: "secret"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "user already exists")
}

func TestAuthUsecase_LoginFailsForInvalidPassword(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.DefaultCost)
	require.NoError(t, err)
	repo := &fakes.FakeUserRepository{
		GetByEmailFn: func(context.Context, string) (*entity.User, error) {
			return &entity.User{ID: uuid.New(), Email: "user@example.com", PasswordHash: string(hash)}, nil
		},
	}
	uc := NewAuthUsecase(repo)

	_, err = uc.Login(context.Background(), "user@example.com", "wrong")
	require.EqualError(t, err, "invalid credentials")
}

func TestReportUsecase_DailyAggregatesEntries(t *testing.T) {
	var capturedFilter repository.EntryFilter
	repo := &fakes.FakeEntryRepository{
		ListFn: func(_ context.Context, _ uuid.UUID, filter repository.EntryFilter) ([]entity.Entry, error) {
			capturedFilter = filter
			return []entity.Entry{
				{DurationSec: 600},
				{DurationSec: 120},
			}, nil
		},
	}
	projectRepo := &fakes.FakeProjectRepository{}
	uc := NewReportUsecase(repo, projectRepo)

	loc := time.FixedZone("JST", 9*3600)
	start := time.Date(2024, 1, 5, 0, 0, 0, 0, loc)
	report, err := uc.Daily(context.Background(), uuid.New(), ReportRange{
		Start:    start,
		End:      start.Add(24 * time.Hour),
		Location: loc,
	})
	require.NoError(t, err)
	require.Equal(t, int64(720), report.TotalSeconds)
	require.Equal(t, "2024-01-05", report.Date)
	require.NotNil(t, capturedFilter.From)
	require.NotNil(t, capturedFilter.To)
	require.True(t, capturedFilter.From.Before(*capturedFilter.To))
}

func TestAuthUsecase_SignupValidatesRequiredFields(t *testing.T) {
	uc := NewAuthUsecase(&fakes.FakeUserRepository{
		GetByEmailFn: func(context.Context, string) (*entity.User, error) {
			return nil, errors.New("not found")
		},
	})
	_, err := uc.Signup(context.Background(), SignupParams{Email: "", Password: ""})
	require.EqualError(t, err, "email and password are required")
}

func TestTagUsecase_CreateUsesDefaultColor(t *testing.T) {
	var saved *entity.Tag
	repo := &fakes.FakeTagRepository{
		CreateFn: func(_ context.Context, tag *entity.Tag) error {
			saved = tag
			return nil
		},
	}
	uc := NewTagUsecase(repo, stubConfig{})

	tag, err := uc.Create(context.Background(), uuid.New(), dto.TagCreateRequest{Name: "Focus"})
	require.NoError(t, err)
	require.Equal(t, "#3B82F6", tag.Color)
	require.Equal(t, saved.ID, tag.ID)
}

func TestTagUsecase_UpdateValidatesColor(t *testing.T) {
	tagID := uuid.New()
	repo := &fakes.FakeTagRepository{
		GetByIDFn: func(context.Context, uuid.UUID, uuid.UUID) (*entity.Tag, error) {
			return &entity.Tag{ID: tagID, UserID: uuid.New(), Name: "Focus", Color: "#000000"}, nil
		},
	}
	uc := NewTagUsecase(repo, stubConfig{})
	invalid := "blue"
	_, err := uc.Update(context.Background(), uuid.New(), tagID, dto.TagUpdateRequest{Color: &invalid})
	var valErr dto.ValidationError
	require.True(t, errors.As(err, &valErr))
}

func TestReportUsecase_WeeklyAggregates(t *testing.T) {
	userID := uuid.New()
	projectID := uuid.New()
	tagID := uuid.New()
	repo := &fakes.FakeEntryRepository{
		ListFn: func(_ context.Context, _ uuid.UUID, filter repository.EntryFilter) ([]entity.Entry, error) {
			return []entity.Entry{
				{DurationSec: 600, StartedAt: filter.From.Add(2 * time.Hour), ProjectID: &projectID, Tags: []entity.Tag{{ID: tagID, Name: "Deep Work", Color: "#ff0000"}}},
				{DurationSec: 300, StartedAt: filter.From.AddDate(0, 0, 3)},
			}, nil
		},
	}
	projectRepo := &fakes.FakeProjectRepository{
		ListFn: func(context.Context, uuid.UUID) ([]entity.Project, error) {
			return []entity.Project{{ID: projectID, UserID: userID, Name: "Backend", Color: "#111111"}}, nil
		},
	}
	uc := NewReportUsecase(repo, projectRepo)
	loc := time.FixedZone("UTC+1", 3600)
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, loc)
	report, err := uc.Weekly(context.Background(), userID, ReportRange{
		Start:    start,
		End:      start.AddDate(0, 0, 7),
		Location: loc,
	})
	require.NoError(t, err)
	require.Len(t, report.Days, 7)
	require.Equal(t, int64(900), report.TotalSeconds)
	require.Equal(t, "2024-01-01", report.WeekStart)
	require.Len(t, report.Projects, 2)
	require.Equal(t, "Backend", report.Projects[0].Name)
	require.Equal(t, int64(600), report.Projects[0].TotalSeconds)
	require.Equal(t, "Unassigned", report.Projects[1].Name)
	require.Equal(t, int64(300), report.Projects[1].TotalSeconds)
	require.Len(t, report.Tags, 1)
	require.Equal(t, tagID, report.Tags[0].TagID)
}

func TestReportUsecase_MonthlyBreakdown(t *testing.T) {
	projectID := uuid.New()
	repo := &fakes.FakeEntryRepository{
		ListFn: func(_ context.Context, _ uuid.UUID, filter repository.EntryFilter) ([]entity.Entry, error) {
			return []entity.Entry{
				{DurationSec: 600, StartedAt: filter.From.Add(24 * time.Hour), ProjectID: &projectID},
				{DurationSec: 300, StartedAt: filter.From.AddDate(0, 0, 10)},
			}, nil
		},
	}
	projectRepo := &fakes.FakeProjectRepository{
		ListFn: func(context.Context, uuid.UUID) ([]entity.Project, error) {
			return []entity.Project{{ID: projectID, Name: "Backend", Color: "#111"}}, nil
		},
	}
	uc := NewReportUsecase(repo, projectRepo)
	loc := time.UTC
	start := time.Date(2024, 2, 1, 0, 0, 0, 0, loc)
	report, err := uc.Monthly(context.Background(), uuid.New(), ReportRange{
		Start:    start,
		End:      start.AddDate(0, 1, 0),
		Location: loc,
	})
	require.NoError(t, err)
	require.Equal(t, "2024-02", report.Month)
	require.Equal(t, int64(900), report.TotalSeconds)
	require.NotEmpty(t, report.Projects)
	require.Equal(t, "Backend", report.Projects[0].Name)
}

func TestDistributeAllocations_MinSumExceedsTotal(t *testing.T) {
	_, err := distributeAllocations(dto.AllocationRequestData{
		TotalMinutes: 30,
		Tasks: []dto.AllocationTaskData{
			{TaskID: "a", Ratio: 1, MinMinutes: intPtr(20)},
			{TaskID: "b", Ratio: 1, MinMinutes: intPtr(15)},
		},
	})
	require.EqualError(t, err, "total_minutes is smaller than the sum of min_minutes")
}

func TestDistributeAllocations_AllMaxBoundedExceedsTotal(t *testing.T) {
	_, err := distributeAllocations(dto.AllocationRequestData{
		TotalMinutes: 50,
		Tasks: []dto.AllocationTaskData{
			{TaskID: "a", Ratio: 1, MaxMinutes: intPtr(10)},
			{TaskID: "b", Ratio: 2, MaxMinutes: intPtr(20)},
		},
	})
	require.EqualError(t, err, "total_minutes exceeds the sum of max_minutes")
}

func TestDistributeAllocations_RespectsMaxConstraints(t *testing.T) {
	allocations, err := distributeAllocations(dto.AllocationRequestData{
		TotalMinutes: 10,
		Tasks: []dto.AllocationTaskData{
			{TaskID: "a", Ratio: 9, MaxMinutes: intPtr(5)},
			{TaskID: "b", Ratio: 1, MaxMinutes: intPtr(10)},
		},
	})
	require.NoError(t, err)
	require.Equal(t, 5, allocations[0].AllocatedMinutes)
	require.Equal(t, 5, allocations[1].AllocatedMinutes)
}

func TestDistributeAllocations_StableRemainderOrdering(t *testing.T) {
	allocations, err := distributeAllocations(dto.AllocationRequestData{
		TotalMinutes: 10,
		Tasks: []dto.AllocationTaskData{
			{TaskID: "a", Ratio: 1},
			{TaskID: "b", Ratio: 1},
			{TaskID: "c", Ratio: 1},
		},
	})
	require.NoError(t, err)
	require.Equal(t, []int{4, 3, 3}, []int{
		allocations[0].AllocatedMinutes,
		allocations[1].AllocatedMinutes,
		allocations[2].AllocatedMinutes,
	})
}

type stubConfig struct{}

func (stubConfig) DefaultProjectColor() string {
	return "#3B82F6"
}

func (stubConfig) SessionTTL() time.Duration {
	return time.Hour
}

var _ provider.AppConfig = stubConfig{}

func intPtr(value int) *int {
	return &value
}
