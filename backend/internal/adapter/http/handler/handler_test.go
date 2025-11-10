package handler

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"chronome/internal/adapter/http/middleware"
	"chronome/internal/adapter/infra/config"
	sess "chronome/internal/adapter/infra/session"
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
	h, store, cfg := newAPIHandlerForTests(t, projectRepo, entryRepo, nil)

	userID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/api/projects/", nil)
	addSessionCookie(t, store, cfg, req, userID)
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
	h, store, cfg := newAPIHandlerForTests(t, nil, nil, tagRepo)
	userID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/api/tags/", nil)
	addSessionCookie(t, store, cfg, req, userID)
	rec := httptest.NewRecorder()

	h.Router().ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), `"tags"`)
	require.Equal(t, userID, captured)
}

func TestAPIHandler_CreateEntryValidationError(t *testing.T) {
	h, store, cfg := newAPIHandlerForTests(t, &fakes.FakeProjectRepository{}, &fakes.FakeEntryRepository{}, nil)
	userID := uuid.New()
	body := bytes.NewBufferString(`{"notes":"missing title"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/entries/", body)
	addSessionCookie(t, store, cfg, req, userID)
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
	h, store, cfg := newAPIHandlerForTests(t, &fakes.FakeProjectRepository{}, entryRepo, nil)
	userID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/api/entries/", nil)
	addSessionCookie(t, store, cfg, req, userID)
	rec := httptest.NewRecorder()

	h.Router().ServeHTTP(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	require.Contains(t, rec.Body.String(), "database offline")
}

func TestAPIHandler_MonthlyReportRequiresMonth(t *testing.T) {
	h, store, cfg := newAPIHandlerForTests(t, nil, nil, nil)
	userID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/api/reports/monthly", nil)
	addSessionCookie(t, store, cfg, req, userID)
	rec := httptest.NewRecorder()

	h.Router().ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAPIHandler_WeeklyReportSuccess(t *testing.T) {
	h, store, cfg := newAPIHandlerForTests(t, nil, &fakes.FakeEntryRepository{
		ListFn: func(context.Context, uuid.UUID, repository.EntryFilter) ([]entity.Entry, error) {
			return nil, nil
		},
	}, nil)
	userID := uuid.New()
	weekStart := time.Now().UTC().Format("2006-01-02")
	req := httptest.NewRequest(http.MethodGet, "/api/reports/weekly?week_start="+weekStart, nil)
	addSessionCookie(t, store, cfg, req, userID)
	rec := httptest.NewRecorder()

	h.Router().ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestAPIHandler_LoginSetsSecureCookie(t *testing.T) {
	entryRepo := &fakes.FakeEntryRepository{}
	projectRepo := &fakes.FakeProjectRepository{}
	store, err := sess.NewSignedCookieStore("another-secret")
	require.NoError(t, err)

	hash, err := bcrypt.GenerateFromPassword([]byte("s3cret"), bcrypt.MinCost)
	require.NoError(t, err)
	user := &entity.User{
		ID:           uuid.New(),
		Email:        "user@example.com",
		PasswordHash: string(hash),
	}
	userRepo := &fakes.FakeUserRepository{
		GetByEmailFn: func(context.Context, string) (*entity.User, error) {
			return user, nil
		},
	}
	auth := usecase.NewAuthUsecase(userRepo)
	cfg := config.Config{
		AllowedOrigin:          "http://localhost:5173",
		SessionTTLValue:        time.Hour,
		SessionSecret:          "another-secret",
		SessionCookieSecure:    true,
		DefaultProjectColorHex: "#3B82F6",
	}
	tagUC := usecase.NewTagUsecase(&fakes.FakeTagRepository{}, cfg)
	entryUC := usecase.NewEntryUsecase(entryRepo, &fakes.FakeTagRepository{}, fakes.FixedTimeProvider{})
	handler := NewAPIHandler(cfg, store, auth, usecase.NewProjectUsecase(projectRepo, cfg), tagUC, entryUC, usecase.NewReportUsecase(entryRepo, projectRepo))

	body := bytes.NewBufferString(`{"email":"user@example.com","password":"s3cret"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", body)
	rec := httptest.NewRecorder()

	handler.Router().ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var sessionCookie *http.Cookie
	var csrfCookie *http.Cookie
	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == middleware.SessionCookieName {
			sessionCookie = cookie
		}
		if cookie.Name == middleware.CSRFCookieName {
			csrfCookie = cookie
		}
	}
	require.NotNil(t, sessionCookie)
	require.True(t, sessionCookie.Secure)
	require.True(t, sessionCookie.HttpOnly)
	require.Equal(t, http.SameSiteLaxMode, sessionCookie.SameSite)
	require.Equal(t, "/", sessionCookie.Path)
	require.NotZero(t, sessionCookie.Expires)
	require.True(t, sessionCookie.Expires.After(time.Now().Add(30*time.Minute)))
	require.Equal(t, int(cfg.SessionTTL().Seconds()), sessionCookie.MaxAge)
	require.NotNil(t, csrfCookie)
	require.False(t, csrfCookie.HttpOnly)
	require.True(t, csrfCookie.Secure)
	require.Equal(t, http.SameSiteLaxMode, csrfCookie.SameSite)
	require.Equal(t, middleware.CSRFCookieName, csrfCookie.Name)
}

// helpers -------------------------------------------------------------------

func newAPIHandlerForTests(t *testing.T, projectRepo *fakes.FakeProjectRepository, entryRepo *fakes.FakeEntryRepository, tagRepo *fakes.FakeTagRepository) (*APIHandler, sess.Store, config.Config) {
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
	userRepo := &fakes.FakeUserRepository{
		GetByEmailFn: func(context.Context, string) (*entity.User, error) {
			return nil, errors.New("not found")
		},
		GetByIDFn: func(_ context.Context, id uuid.UUID) (*entity.User, error) {
			return &entity.User{ID: id, TimeZone: "UTC"}, nil
		},
	}
	cfg := config.Config{
		AllowedOrigin:          "http://localhost:5173",
		SessionTTLValue:        time.Hour,
		SessionSecret:          "test-secret",
		SessionCookieSecure:    false,
		DefaultProjectColorHex: "#3B82F6",
	}
	store, err := sess.NewSignedCookieStore(cfg.SessionSecret)
	require.NoError(t, err)
	auth := usecase.NewAuthUsecase(userRepo)
	projects := usecase.NewProjectUsecase(projectRepo, cfg)
	tags := usecase.NewTagUsecase(tagRepo, cfg)
	entries := usecase.NewEntryUsecase(entryRepo, tagRepo, fakes.FixedTimeProvider{})
	reports := usecase.NewReportUsecase(entryRepo, projectRepo)
	return NewAPIHandler(cfg, store, auth, projects, tags, entries, reports), store, cfg
}

func addSessionCookie(t *testing.T, store sess.Store, cfg config.Config, req *http.Request, userID uuid.UUID) {
	t.Helper()
	sessionID, err := store.Create(userID, cfg.SessionTTL())
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{
		Name:  middleware.SessionCookieName,
		Value: sessionID,
	})
	csrfToken := uuid.NewString()
	req.AddCookie(&http.Cookie{
		Name:  middleware.CSRFCookieName,
		Value: csrfToken,
	})
	req.Header.Set(middleware.CSRFHeaderName, csrfToken)
}
