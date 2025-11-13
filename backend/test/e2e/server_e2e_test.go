package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"chronome/internal/adapter/db/gormrepo"
	"chronome/internal/adapter/http/handler"
	"chronome/internal/adapter/http/middleware"
	"chronome/internal/adapter/infra/config"
	"chronome/internal/adapter/infra/database"
	sess "chronome/internal/adapter/infra/session"
	infTime "chronome/internal/adapter/infra/time"
	"chronome/internal/domain/entity"
	"chronome/internal/usecase"
)

func TestChronoMeEndToEnd(t *testing.T) {
	fx := newFixture(t)

	email := "e2e-user@example.com"
	password := "ChronoMePassw0rd!"

	logStep(t, "starting signup scenario")
	signedUp := fx.signup(email, password)
	require.Equal(t, email, signedUp.Email)

	fx.expectLoginFailure(email, "bad-password")

	logStep(t, "performing login")
	loggedIn := fx.login(email, password)
	require.Equal(t, signedUp.ID, loggedIn.ID)

	profile := fx.me()
	require.Equal(t, signedUp.ID, profile.ID)

	logStep(t, "creating project and tag")
	project := fx.createProject("E2E Project", "#3B82F6")
	projects := fx.listProjects()
	require.Len(t, projects, 1)
	require.Equal(t, project.ID, projects[0].ID)

	tag := fx.createTag("Deep Work")
	tag = fx.updateTag(tag.ID, "Deep Focus", "#F97316")
	tags := fx.listTags()
	require.Len(t, tags, 1)
	require.Equal(t, tag.ID, tags[0].ID)
	require.Equal(t, "Deep Focus", tags[0].Name)
	require.Equal(t, "#F97316", tags[0].Color)

	start := time.Now().UTC().Add(-2 * time.Hour).Truncate(time.Second)
	end := start.Add(90 * time.Minute)

	logStep(t, "creating primary entry at %s", start.Format(time.RFC3339))
	entry := fx.createEntry(project.ID, "Implement reporting flow", start, nil)
	require.Nil(t, entry.EndedAt)
	require.EqualValues(t, 0, entry.DurationSec)

	entry = fx.stopEntry(entry.ID, end)
	require.NotNil(t, entry.EndedAt)
	require.EqualValues(t, 90*60, entry.DurationSec)

	secondStart := start.Add(24 * time.Hour).Add(30 * time.Minute)
	secondEnd := secondStart.Add(30 * time.Minute)

	logStep(t, "creating tagged entry at %s", secondStart.Format(time.RFC3339))
	secondEntry := fx.createEntry(project.ID, "Tag exploration", secondStart, []uuid.UUID{tag.ID})
	secondEntry = fx.stopEntry(secondEntry.ID, secondEnd)

	entries := fx.listEntries()
	require.Len(t, entries, 2)
	loadedFirst, ok := entryByID(entries, entry.ID)
	require.True(t, ok)
	require.EqualValues(t, 90*60, loadedFirst.DurationSec)
	loadedSecond, ok := entryByID(entries, secondEntry.ID)
	require.True(t, ok)
	require.EqualValues(t, 30*60, loadedSecond.DurationSec)

	reportDate := start.Format("2006-01-02")
	logStep(t, "fetching daily report for %s", reportDate)
	report := fx.dailyReport(reportDate)
	require.Equal(t, reportDate, report.Date)
	require.EqualValues(t, 90*60, report.TotalSeconds)
	require.Len(t, report.Entries, 1)
	dailyEntry, ok := entryByID(report.Entries, entry.ID)
	require.True(t, ok)
	require.Equal(t, entry.ID, dailyEntry.ID)

	totalSeconds := (90 + 30) * 60

	weekStart := mondayOf(start)
	logStep(t, "fetching weekly report for week starting %s", weekStart.Format("2006-01-02"))
	weekly := fx.weeklyReport(weekStart.Format("2006-01-02"))
	require.Equal(t, weekStart.Format("2006-01-02"), weekly.WeekStart)
	require.EqualValues(t, totalSeconds, weekly.TotalSeconds)
	projectWeekly, ok := projectStat(weekly.Projects, project.ID)
	require.True(t, ok)
	require.EqualValues(t, totalSeconds, projectWeekly.TotalSeconds)
	tagWeekly, ok := tagStat(weekly.Tags, tag.ID)
	require.True(t, ok)
	require.EqualValues(t, 30*60, tagWeekly.TotalSeconds)

	monthKey := start.Format("2006-01")
	logStep(t, "fetching monthly report for %s", monthKey)
	monthly := fx.monthlyReport(monthKey)
	require.Equal(t, monthKey, monthly.Month)
	require.EqualValues(t, totalSeconds, monthly.TotalSeconds)
	require.NotEmpty(t, monthly.Weeks)
	projectMonthly, ok := projectStat(monthly.Projects, project.ID)
	require.True(t, ok)
	require.EqualValues(t, totalSeconds, projectMonthly.TotalSeconds)
	tagMonthly, ok := tagStat(monthly.Tags, tag.ID)
	require.True(t, ok)
	require.EqualValues(t, 30*60, tagMonthly.TotalSeconds)

	logStep(t, "archiving project and deleting tag")
	archived := fx.archiveProject(project.ID)
	require.True(t, archived.IsArchived)
	require.Equal(t, project.ID, archived.ID)

	fx.deleteTag(tag.ID)
	require.Len(t, fx.listTags(), 0)
}

type fixture struct {
	t        *testing.T
	client   *http.Client
	server   *httptest.Server
	baseURL  string
	baseHost *url.URL
	origin   string
}

func newFixture(t *testing.T) *fixture {
	t.Helper()

	cfg := config.Config{
		Address:                ":0",
		DBDriver:               "sqlite",
		DBDsn:                  "file:chronome_e2e?mode=memory&cache=shared",
		SessionTTLValue:        2 * time.Hour,
		SessionSecret:          "0123456789abcdefghijklmnopqrstuvwxyz-secret",
		SessionCookieSecure:    false,
		AllowedOrigin:          "http://localhost",
		Environment:            "test",
		DefaultProjectColorHex: "#3B82F6",
	}

	db, err := database.Open(cfg)
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)

	require.NoError(t, database.Automigrate(db))

	sessionStore := sess.NewMemoryStore()
	userRepo := gormrepo.NewUserRepository(db)
	projectRepo := gormrepo.NewProjectRepository(db)
	entryRepo := gormrepo.NewEntryRepository(db)
	tagRepo := gormrepo.NewTagRepository(db)

	authUC := usecase.NewAuthUsecase(userRepo)
	projectUC := usecase.NewProjectUsecase(projectRepo, cfg)
	tagUC := usecase.NewTagUsecase(tagRepo, cfg)
	entryUC := usecase.NewEntryUsecase(entryRepo, tagRepo, infTime.SystemClock{})
	reportUC := usecase.NewReportUsecase(entryRepo, projectRepo)

	apiHandler := handler.NewAPIHandler(cfg, sessionStore, authUC, projectUC, tagUC, entryUC, reportUC)
	server := httptest.NewServer(apiHandler.Router())

	jar, err := cookiejar.New(nil)
	require.NoError(t, err)

	baseURL, err := url.Parse(server.URL)
	require.NoError(t, err)

	t.Cleanup(func() {
		server.Close()
		_ = sqlDB.Close()
	})

	return &fixture{
		t:        t,
		client:   &http.Client{Jar: jar},
		server:   server,
		baseURL:  server.URL,
		baseHost: baseURL,
		origin:   cfg.AllowedOrigin,
	}
}

func (f *fixture) signup(email, password string) entity.User {
	f.t.Helper()
	var resp struct {
		User entity.User `json:"user"`
	}
	status := f.doJSON(http.MethodPost, "/api/auth/signup", map[string]string{
		"email":        email,
		"password":     password,
		"display_name": "E2E User",
		"time_zone":    "UTC",
	}, &resp)
	require.Equal(f.t, http.StatusCreated, status)
	return resp.User
}

func (f *fixture) expectLoginFailure(email, password string) {
	f.t.Helper()
	status, body := f.do(http.MethodPost, "/api/auth/login", map[string]string{
		"email":    email,
		"password": password,
	})
	require.Equal(f.t, http.StatusUnauthorized, status, string(body))
}

func (f *fixture) login(email, password string) entity.User {
	f.t.Helper()
	var resp struct {
		User entity.User `json:"user"`
	}
	status := f.doJSON(http.MethodPost, "/api/auth/login", map[string]string{
		"email":    email,
		"password": password,
	}, &resp)
	require.Equal(f.t, http.StatusOK, status)
	require.NotEmpty(f.t, f.csrfToken(), "login should issue CSRF token")
	return resp.User
}

func (f *fixture) me() entity.User {
	f.t.Helper()
	var resp struct {
		User entity.User `json:"user"`
	}
	status := f.doJSON(http.MethodGet, "/api/auth/me", nil, &resp)
	require.Equal(f.t, http.StatusOK, status)
	return resp.User
}

func (f *fixture) createProject(name, color string) entity.Project {
	f.t.Helper()
	var project entity.Project
	status := f.doJSON(http.MethodPost, "/api/projects/", map[string]string{
		"name":        name,
		"color":       color,
		"description": "E2E project",
	}, &project)
	require.Equal(f.t, http.StatusCreated, status)
	return project
}

func (f *fixture) listProjects() []entity.Project {
	f.t.Helper()
	var resp struct {
		Projects []entity.Project `json:"projects"`
	}
	status := f.doJSON(http.MethodGet, "/api/projects/", nil, &resp)
	require.Equal(f.t, http.StatusOK, status)
	return resp.Projects
}

func (f *fixture) createEntry(projectID uuid.UUID, title string, started time.Time, tagIDs []uuid.UUID) entity.Entry {
	f.t.Helper()
	var entry entity.Entry
	payload := map[string]any{
		"title":      title,
		"notes":      "Focus work",
		"project_id": projectID.String(),
		"started_at": started.Format(time.RFC3339),
	}
	if len(tagIDs) > 0 {
		payload["tag_ids"] = uuidList(tagIDs)
	}
	status := f.doJSON(http.MethodPost, "/api/entries/", payload, &entry)
	require.Equal(f.t, http.StatusCreated, status)
	return entry
}

func (f *fixture) stopEntry(id uuid.UUID, ended time.Time) entity.Entry {
	f.t.Helper()
	var entry entity.Entry
	status := f.doJSON(http.MethodPatch, fmt.Sprintf("/api/entries/%s", id.String()), map[string]string{
		"ended_at": ended.Format(time.RFC3339),
	}, &entry)
	require.Equal(f.t, http.StatusOK, status)
	return entry
}

func (f *fixture) listEntries() []entity.Entry {
	f.t.Helper()
	var resp struct {
		Entries []entity.Entry `json:"entries"`
	}
	status := f.doJSON(http.MethodGet, "/api/entries/", nil, &resp)
	require.Equal(f.t, http.StatusOK, status)
	return resp.Entries
}

func (f *fixture) dailyReport(date string) usecase.DailyReport {
	f.t.Helper()
	var report usecase.DailyReport
	status := f.doJSON(http.MethodGet, "/api/reports/daily?date="+url.QueryEscape(date), nil, &report)
	require.Equal(f.t, http.StatusOK, status)
	return report
}

func (f *fixture) weeklyReport(weekStart string) usecase.WeeklyReport {
	f.t.Helper()
	var report usecase.WeeklyReport
	status := f.doJSON(http.MethodGet, "/api/reports/weekly?week_start="+url.QueryEscape(weekStart), nil, &report)
	require.Equal(f.t, http.StatusOK, status)
	return report
}

func (f *fixture) monthlyReport(month string) usecase.MonthlyReport {
	f.t.Helper()
	var report usecase.MonthlyReport
	status := f.doJSON(http.MethodGet, "/api/reports/monthly?month="+url.QueryEscape(month), nil, &report)
	require.Equal(f.t, http.StatusOK, status)
	return report
}

func (f *fixture) createTag(name string) entity.Tag {
	f.t.Helper()
	var tag entity.Tag
	status := f.doJSON(http.MethodPost, "/api/tags/", map[string]string{
		"name": name,
	}, &tag)
	require.Equal(f.t, http.StatusCreated, status)
	return tag
}

func (f *fixture) listTags() []entity.Tag {
	f.t.Helper()
	var resp struct {
		Tags []entity.Tag `json:"tags"`
	}
	status := f.doJSON(http.MethodGet, "/api/tags/", nil, &resp)
	require.Equal(f.t, http.StatusOK, status)
	return resp.Tags
}

func (f *fixture) updateTag(id uuid.UUID, name, color string) entity.Tag {
	f.t.Helper()
	var tag entity.Tag
	payload := map[string]string{}
	if name != "" {
		payload["name"] = name
	}
	if color != "" {
		payload["color"] = color
	}
	status := f.doJSON(http.MethodPatch, fmt.Sprintf("/api/tags/%s", id.String()), payload, &tag)
	require.Equal(f.t, http.StatusOK, status)
	return tag
}

func (f *fixture) deleteTag(id uuid.UUID) {
	f.t.Helper()
	status, _ := f.do(http.MethodDelete, fmt.Sprintf("/api/tags/%s", id.String()), nil)
	require.Equal(f.t, http.StatusNoContent, status)
}

func (f *fixture) archiveProject(id uuid.UUID) entity.Project {
	f.t.Helper()
	var project entity.Project
	status := f.doJSON(http.MethodPatch, fmt.Sprintf("/api/projects/%s", id.String()), map[string]any{
		"is_archived": true,
	}, &project)
	require.Equal(f.t, http.StatusOK, status)
	return project
}

func (f *fixture) doJSON(method, path string, body any, target any) int {
	status, raw := f.do(method, path, body)
	if target != nil && len(raw) > 0 {
		require.NoError(f.t, json.Unmarshal(raw, target))
	}
	return status
}

func (f *fixture) do(method, path string, body any) (int, []byte) {
	f.t.Helper()
	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		require.NoError(f.t, err)
		reader = bytes.NewReader(payload)
	}
	req, err := http.NewRequest(method, f.baseURL+path, reader)
	require.NoError(f.t, err)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if method != http.MethodGet && method != http.MethodHead && method != http.MethodOptions {
		req.Header.Set("Origin", f.origin)
		if token := f.csrfToken(); token != "" {
			req.Header.Set(middleware.CSRFHeaderName, token)
		}
	}
	resp, err := f.client.Do(req)
	require.NoError(f.t, err)
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	require.NoError(f.t, err)
	return resp.StatusCode, data
}

func (f *fixture) csrfToken() string {
	f.t.Helper()
	for _, cookie := range f.client.Jar.Cookies(f.baseHost) {
		if cookie.Name == middleware.CSRFCookieName {
			return cookie.Value
		}
	}
	return ""
}

func uuidList(values []uuid.UUID) []string {
	result := make([]string, len(values))
	for i := range values {
		result[i] = values[i].String()
	}
	return result
}

func mondayOf(t time.Time) time.Time {
	for t.Weekday() != time.Monday {
		t = t.AddDate(0, 0, -1)
	}
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func projectStat(list []usecase.ProjectBreakdown, id uuid.UUID) (usecase.ProjectBreakdown, bool) {
	for _, item := range list {
		if item.ProjectID != nil && *item.ProjectID == id {
			return item, true
		}
	}
	return usecase.ProjectBreakdown{}, false
}

func tagStat(list []usecase.TagBreakdown, id uuid.UUID) (usecase.TagBreakdown, bool) {
	for _, item := range list {
		if item.TagID == id {
			return item, true
		}
	}
	return usecase.TagBreakdown{}, false
}

func entryByID(entries []entity.Entry, id uuid.UUID) (entity.Entry, bool) {
	for _, entry := range entries {
		if entry.ID == id {
			return entry, true
		}
	}
	return entity.Entry{}, false
}

func logStep(t *testing.T, format string, args ...any) {
	t.Helper()
	t.Logf("[E2E] "+format, args...)
}
