package handler

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"chronome/internal/adapter/infra/config"
	"chronome/internal/domain/entity"
	"chronome/internal/domain/repository"
	"chronome/internal/usecase"
	"chronome/test/fakes"
)

func TestAPIHandler_ListProjectsSuccess(t *testing.T) {
	var receivedUser uuid.UUID
	projectRepo := &fakes.FakeProjectRepository{
		ListFn: func(ctx context.Context, userID uuid.UUID) ([]entity.Project, error) {
			receivedUser = userID
			return []entity.Project{
				{ID: uuid.New(), UserID: userID, Name: "Chrono", Color: "#111111"},
			}, nil
		},
	}
	entryRepo := &fakes.FakeEntryRepository{}
	h, cfg := newAPIHandlerForTests(t, projectRepo, entryRepo, nil, nil)

	userID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/api/projects/", nil)
	addAuthHeader(t, req, cfg, userID)
	rec := httptest.NewRecorder()

	h.Router().ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), `"projects"`)
	require.Equal(t, userID, receivedUser)
}

func TestAPIHandler_ListTagsSuccess(t *testing.T) {
	var captured uuid.UUID
	tagRepo := &fakes.FakeTagRepository{
		ListFn: func(ctx context.Context, userID uuid.UUID) ([]entity.Tag, error) {
			captured = userID
			return []entity.Tag{{ID: uuid.New(), UserID: userID, Name: "Focus", Color: "#000000"}}, nil
		},
	}
	h, cfg := newAPIHandlerForTests(t, nil, nil, tagRepo, nil)
	userID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/api/tags/", nil)
	addAuthHeader(t, req, cfg, userID)
	rec := httptest.NewRecorder()

	h.Router().ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), `"tags"`)
	require.Equal(t, userID, captured)
}

func TestAPIHandler_CreateEntryValidationError(t *testing.T) {
	h, cfg := newAPIHandlerForTests(t, &fakes.FakeProjectRepository{}, &fakes.FakeEntryRepository{}, nil, nil)
	userID := uuid.New()
	body := bytes.NewBufferString(`{"notes":"missing title"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/entries/", body)
	addAuthHeader(t, req, cfg, userID)
	rec := httptest.NewRecorder()

	h.Router().ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Contains(t, rec.Body.String(), "title")
}

func TestAPIHandler_ListEntriesServerError(t *testing.T) {
	entryRepo := &fakes.FakeEntryRepository{
		ListFn: func(context.Context, uuid.UUID, repository.EntryFilter) ([]entity.Entry, error) {
			return nil, errors.New("database offline")
		},
	}
	h, cfg := newAPIHandlerForTests(t, &fakes.FakeProjectRepository{}, entryRepo, nil, nil)
	userID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/api/entries/", nil)
	addAuthHeader(t, req, cfg, userID)
	rec := httptest.NewRecorder()

	h.Router().ServeHTTP(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	require.Contains(t, rec.Body.String(), "database offline")
}

func TestAPIHandler_MonthlyReportRequiresMonth(t *testing.T) {
	h, cfg := newAPIHandlerForTests(t, nil, nil, nil, nil)
	userID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/api/reports/monthly", nil)
	addAuthHeader(t, req, cfg, userID)
	rec := httptest.NewRecorder()

	h.Router().ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAPIHandler_WeeklyReportSuccess(t *testing.T) {
	h, cfg := newAPIHandlerForTests(t, nil, &fakes.FakeEntryRepository{
		ListFn: func(context.Context, uuid.UUID, repository.EntryFilter) ([]entity.Entry, error) {
			return nil, nil
		},
	}, nil, nil)
	userID := uuid.New()
	weekStart := time.Now().UTC().Format("2006-01-02")
	req := httptest.NewRequest(http.MethodGet, "/api/reports/weekly?week_start="+weekStart, nil)
	addAuthHeader(t, req, cfg, userID)
	rec := httptest.NewRecorder()

	h.Router().ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestAPIHandler_CreateAllocationSuccess(t *testing.T) {
	var receivedRequest *entity.AllocationRequest
	var receivedAllocations []entity.TaskAllocation
	allocationRepo := &fakes.FakeAllocationRepository{
		CreateFn: func(ctx context.Context, request *entity.AllocationRequest, allocations []entity.TaskAllocation) error {
			receivedRequest = request
			receivedAllocations = allocations
			return nil
		},
	}
	h, cfg := newAPIHandlerForTests(t, nil, nil, nil, allocationRepo)
	userID := uuid.New()
	body := bytes.NewBufferString(`{"total_minutes":60,"tasks":[{"task_id":"task-a","ratio":1},{"task_id":"task-b","ratio":1}]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/allocations/", body)
	addAuthHeader(t, req, cfg, userID)
	rec := httptest.NewRecorder()

	h.Router().ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)
	require.NotNil(t, receivedRequest)
	require.Equal(t, 60, receivedRequest.TotalMinutes)
	require.Len(t, receivedAllocations, 2)
	allocations := map[string]int{}
	for _, allocation := range receivedAllocations {
		allocations[allocation.TaskID] = allocation.AllocatedMinutes
	}
	require.Equal(t, 30, allocations["task-a"])
	require.Equal(t, 30, allocations["task-b"])
}

func TestAPIHandler_CreateAllocationValidationError(t *testing.T) {
	called := false
	allocationRepo := &fakes.FakeAllocationRepository{
		CreateFn: func(context.Context, *entity.AllocationRequest, []entity.TaskAllocation) error {
			called = true
			return nil
		},
	}
	h, cfg := newAPIHandlerForTests(t, nil, nil, nil, allocationRepo)
	userID := uuid.New()
	body := bytes.NewBufferString(`{"total_minutes":0,"tasks":[{"task_id":"task-a","ratio":1}]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/allocations/", body)
	addAuthHeader(t, req, cfg, userID)
	rec := httptest.NewRecorder()

	h.Router().ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	require.False(t, called)
}

func TestAPIHandler_CreateAllocationConstraintError(t *testing.T) {
	called := false
	allocationRepo := &fakes.FakeAllocationRepository{
		CreateFn: func(context.Context, *entity.AllocationRequest, []entity.TaskAllocation) error {
			called = true
			return nil
		},
	}
	h, cfg := newAPIHandlerForTests(t, nil, nil, nil, allocationRepo)
	userID := uuid.New()
	body := bytes.NewBufferString(`{"total_minutes":10,"tasks":[{"task_id":"task-a","ratio":1,"min_minutes":20}]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/allocations/", body)
	addAuthHeader(t, req, cfg, userID)
	rec := httptest.NewRecorder()

	h.Router().ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	require.False(t, called)
}

func TestAPIHandler_CreateAllocationRepositoryError(t *testing.T) {
	called := false
	allocationRepo := &fakes.FakeAllocationRepository{
		CreateFn: func(context.Context, *entity.AllocationRequest, []entity.TaskAllocation) error {
			called = true
			return errors.New("db error")
		},
	}
	h, cfg := newAPIHandlerForTests(t, nil, nil, nil, allocationRepo)
	userID := uuid.New()
	body := bytes.NewBufferString(`{"total_minutes":60,"tasks":[{"task_id":"task-a","ratio":1},{"task_id":"task-b","ratio":1}]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/allocations/", body)
	addAuthHeader(t, req, cfg, userID)
	rec := httptest.NewRecorder()

	h.Router().ServeHTTP(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	require.True(t, called)
}

// ヘルパー ------------------------------------------------------------------

func newAPIHandlerForTests(t *testing.T, projectRepo *fakes.FakeProjectRepository, entryRepo *fakes.FakeEntryRepository, tagRepo *fakes.FakeTagRepository, allocationRepo *fakes.FakeAllocationRepository) (*APIHandler, config.Config) {
	t.Helper()
	if projectRepo == nil {
		projectRepo = &fakes.FakeProjectRepository{}
	}
	if entryRepo == nil {
		entryRepo = &fakes.FakeEntryRepository{}
	}
	if tagRepo == nil {
		tagRepo = &fakes.FakeTagRepository{}
	}
	if allocationRepo == nil {
		allocationRepo = &fakes.FakeAllocationRepository{}
	}
	userRepo := &fakes.FakeUserRepository{
		GetByEmailFn: func(context.Context, string) (*entity.User, error) {
			return nil, errors.New("not found")
		},
		GetByIDFn: func(_ context.Context, id uuid.UUID) (*entity.User, error) {
			return &entity.User{ID: id, TimeZone: "UTC"}, nil
		},
		GetBySupabaseIDFn: func(_ context.Context, supabaseID uuid.UUID) (*entity.User, error) {
			return &entity.User{ID: supabaseID, Email: "user@example.com", SupabaseUserID: &supabaseID, TimeZone: "UTC"}, nil
		},
	}
	cfg := config.Config{
		AllowedOrigin:          "http://localhost:5173",
		DefaultProjectColorHex: "#3B82F6",
		SupabaseJWTSecret:      "test-secret",
	}
	projects := usecase.NewProjectUsecase(projectRepo, cfg)
	tags := usecase.NewTagUsecase(tagRepo, cfg)
	entries := usecase.NewEntryUsecase(entryRepo, tagRepo, fakes.FixedTimeProvider{})
	reports := usecase.NewReportUsecase(entryRepo, projectRepo)
	allocationUC := usecase.NewAllocationUsecase(allocationRepo, fakes.FixedTimeProvider{})
	return NewAPIHandler(cfg, userRepo, projects, tags, entries, reports, allocationUC), cfg
}

func addAuthHeader(t *testing.T, req *http.Request, cfg config.Config, userID uuid.UUID) {
	t.Helper()
	token, err := testJWT(userID, cfg.SupabaseJWTSecret)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
}

func testJWT(subject uuid.UUID, secret string) (string, error) {
	header, err := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	if err != nil {
		return "", err
	}
	payload, err := json.Marshal(map[string]any{
		"sub":   subject.String(),
		"email": "user@example.com",
		"exp":   time.Now().Add(time.Hour).Unix(),
	})
	if err != nil {
		return "", err
	}
	unsigned := base64.RawURLEncoding.EncodeToString(header) + "." + base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(unsigned))
	return unsigned + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}
