package handler

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/google/uuid"

	"chronome/internal/adapter/http/middleware"
	"chronome/internal/adapter/infra/config"
	sess "chronome/internal/adapter/infra/session"
	"chronome/internal/domain/entity"
	"chronome/internal/domain/repository"
	"chronome/internal/usecase"
	"chronome/internal/usecase/dto"
)

const (
	maxReportLookbackDays = 370
	maxReportFutureDays   = 31
)

// APIHandler は HTTP エンドポイントをユースケースに接続する。
type APIHandler struct {
	auth     *usecase.AuthUsecase
	projects *usecase.ProjectUsecase
	tags     *usecase.TagUsecase
	entries  *usecase.EntryUsecase
	reports  *usecase.ReportUsecase
	allocs   *usecase.AllocationUsecase
	sessions sess.Store
	cfg      config.Config
}

func NewAPIHandler(cfg config.Config, sessions sess.Store, auth *usecase.AuthUsecase, projects *usecase.ProjectUsecase, tags *usecase.TagUsecase, entries *usecase.EntryUsecase, reports *usecase.ReportUsecase, allocs *usecase.AllocationUsecase) *APIHandler {
	return &APIHandler{
		auth:     auth,
		projects: projects,
		tags:     tags,
		entries:  entries,
		reports:  reports,
		allocs:   allocs,
		sessions: sessions,
		cfg:      cfg,
	}
}

// Router はミドルウェアとルートを登録した chi ルーターを構築する。
func (h *APIHandler) Router() *chi.Mux {
	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{h.cfg.AllowedOrigin},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", middleware.CSRFHeaderName},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Use(middleware.WithSession(h.sessions))

	r.Get("/healthz", h.healthz)

	r.Route("/api", func(api chi.Router) {
		api.Route("/auth", func(auth chi.Router) {
			auth.Post("/signup", h.signup)
			auth.Post("/login", h.login)
			auth.With(middleware.RequireAuth).Get("/me", h.me)
			auth.With(middleware.RequireAuth, middleware.RequireCSRF(h.cfg.AllowedOrigin)).Post("/logout", h.logout)
		})

		api.With(middleware.RequireAuth).Route("/projects", func(pr chi.Router) {
			pr.Get("/", h.listProjects)
			pr.With(middleware.RequireCSRF(h.cfg.AllowedOrigin)).Post("/", h.createProject)
			pr.With(middleware.RequireCSRF(h.cfg.AllowedOrigin)).Patch("/{id}", h.updateProject)
			pr.With(middleware.RequireCSRF(h.cfg.AllowedOrigin)).Delete("/{id}", h.deleteProject)
		})
		api.With(middleware.RequireAuth).Route("/tags", func(tr chi.Router) {
			tr.Get("/", h.listTags)
			tr.With(middleware.RequireCSRF(h.cfg.AllowedOrigin)).Post("/", h.createTag)
			tr.With(middleware.RequireCSRF(h.cfg.AllowedOrigin)).Patch("/{id}", h.updateTag)
			tr.With(middleware.RequireCSRF(h.cfg.AllowedOrigin)).Delete("/{id}", h.deleteTag)
		})

		api.With(middleware.RequireAuth).Route("/entries", func(er chi.Router) {
			er.Get("/", h.listEntries)
			er.With(middleware.RequireCSRF(h.cfg.AllowedOrigin)).Post("/", h.createEntry)
			er.With(middleware.RequireCSRF(h.cfg.AllowedOrigin)).Patch("/{id}", h.updateEntry)
			er.With(middleware.RequireCSRF(h.cfg.AllowedOrigin)).Delete("/{id}", h.deleteEntry)
		})

		api.With(middleware.RequireAuth).Route("/allocations", func(ar chi.Router) {
			ar.With(middleware.RequireCSRF(h.cfg.AllowedOrigin)).Post("/", h.createAllocation)
		})

		api.With(middleware.RequireAuth).Route("/reports", func(rr chi.Router) {
			rr.Get("/daily", h.dailyReport)
			rr.Get("/weekly", h.weeklyReport)
			rr.Get("/monthly", h.monthlyReport)
		})
	})

	return r
}

func (h *APIHandler) healthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (h *APIHandler) signup(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Email       string `json:"email"`
		Password    string `json:"password"`
		DisplayName string `json:"display_name"`
		TimeZone    string `json:"time_zone"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		respondError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	user, err := h.auth.Signup(r.Context(), usecase.SignupParams(payload))
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]any{"user": mapUser(user)})
}

func (h *APIHandler) login(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		respondError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	user, err := h.auth.Login(r.Context(), payload.Email, payload.Password)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	sessionID, err := h.sessions.Create(user.ID, h.cfg.SessionTTL())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "session error")
		return
	}
	expiresAt := time.Now().UTC().Add(h.cfg.SessionTTL())
	http.SetCookie(w, &http.Cookie{
		Name:     middleware.SessionCookieName,
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.cfg.SessionCookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(h.cfg.SessionTTL().Seconds()),
		Expires:  expiresAt,
	})
	csrfToken, err := generateCSRFToken()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "session error")
		return
	}
	h.setCSRFCookie(w, csrfToken, expiresAt)
	respondJSON(w, http.StatusOK, map[string]any{"user": mapUser(user)})
}

func (h *APIHandler) logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(middleware.SessionCookieName); err == nil {
		h.sessions.Delete(cookie.Value)
		cookie.Value = ""
		cookie.Path = "/"
		cookie.HttpOnly = true
		cookie.Secure = h.cfg.SessionCookieSecure
		cookie.SameSite = http.SameSiteLaxMode
		cookie.MaxAge = -1
		cookie.Expires = time.Unix(0, 0)
		http.SetCookie(w, cookie)
	}
	h.clearCSRFCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h *APIHandler) me(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	user, err := h.auth.GetProfile(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusNotFound, "user not found")
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"user": mapUser(user)})
}

func (h *APIHandler) listProjects(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	projects, err := h.projects.List(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"projects": projects})
}

func (h *APIHandler) createProject(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	var payload dto.ProjectCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		respondError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	project, err := h.projects.Create(r.Context(), userID, payload)
	if err != nil {
		respondUsecaseError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, project)
}

func (h *APIHandler) updateProject(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	pid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var payload dto.ProjectUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		respondError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	project, err := h.projects.Update(r.Context(), userID, pid, payload)
	if err != nil {
		respondUsecaseError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, project)
}

func (h *APIHandler) deleteProject(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	pid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.projects.Delete(r.Context(), userID, pid); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *APIHandler) listTags(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	tags, err := h.tags.List(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"tags": tags})
}

func (h *APIHandler) createTag(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	var payload dto.TagCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		respondError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	tag, err := h.tags.Create(r.Context(), userID, payload)
	if err != nil {
		respondUsecaseError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, tag)
}

func (h *APIHandler) updateTag(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	tid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var payload dto.TagUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		respondError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	tag, err := h.tags.Update(r.Context(), userID, tid, payload)
	if err != nil {
		respondUsecaseError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, tag)
}

func (h *APIHandler) deleteTag(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	tid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.tags.Delete(r.Context(), userID, tid); err != nil {
		respondUsecaseError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *APIHandler) listEntries(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	filter, err := buildEntryFilter(r)
	if err != nil {
		respondUsecaseError(w, err)
		return
	}
	entries, err := h.entries.List(r.Context(), userID, filter)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"entries": entries})
}

func (h *APIHandler) createEntry(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	var payload dto.EntryCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		respondError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	entry, err := h.entries.Create(r.Context(), userID, payload)
	if err != nil {
		respondUsecaseError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, entry)
}

func (h *APIHandler) updateEntry(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	eid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var payload dto.EntryUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		respondError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	entry, err := h.entries.Update(r.Context(), userID, eid, payload)
	if err != nil {
		respondUsecaseError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, entry)
}

func (h *APIHandler) deleteEntry(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	eid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.entries.Delete(r.Context(), userID, eid); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *APIHandler) createAllocation(w http.ResponseWriter, r *http.Request) {
	var payload dto.AllocationRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		respondError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	result, err := h.allocs.Allocate(r.Context(), payload)
	if err != nil {
		var valErr dto.ValidationError
		var constraintErr usecase.AllocationConstraintError
		switch {
		case errors.As(err, &valErr):
			respondError(w, http.StatusUnprocessableEntity, valErr.Error())
			return
		case errors.As(err, &constraintErr):
			respondError(w, http.StatusUnprocessableEntity, constraintErr.Error())
			return
		default:
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	respondJSON(w, http.StatusCreated, map[string]any{
		"request_id":    result.RequestID,
		"total_minutes": result.TotalMinutes,
		"allocations":   result.Allocations,
	})
}

func (h *APIHandler) dailyReport(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	user, err := h.auth.GetProfile(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusNotFound, "user not found")
		return
	}
	loc, err := h.resolveLocation(r, user)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid time_zone")
		return
	}
	day := time.Now().In(loc)
	if v := r.URL.Query().Get("date"); v != "" {
		parsed, parseErr := time.ParseInLocation("2006-01-02", v, loc)
		if parseErr != nil {
			respondError(w, http.StatusBadRequest, "invalid date")
			return
		}
		day = parsed
	}
	start := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, loc)
	if err := enforceReportWindow(start, loc); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	report, err := h.reports.Daily(r.Context(), userID, usecase.ReportRange{
		Start:    start,
		End:      start.AddDate(0, 0, 1),
		Location: loc,
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, report)
}

func (h *APIHandler) weeklyReport(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	user, err := h.auth.GetProfile(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusNotFound, "user not found")
		return
	}
	loc, err := h.resolveLocation(r, user)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid time_zone")
		return
	}
	start := normalizeWeekStart(time.Now().In(loc), loc)
	if v := r.URL.Query().Get("week_start"); v != "" {
		parsed, parseErr := time.ParseInLocation("2006-01-02", v, loc)
		if parseErr != nil {
			respondError(w, http.StatusBadRequest, "invalid week_start")
			return
		}
		start = normalizeWeekStart(parsed, loc)
	}
	if err := enforceReportWindow(start, loc); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	report, err := h.reports.Weekly(r.Context(), userID, usecase.ReportRange{
		Start:    start,
		End:      start.AddDate(0, 0, 7),
		Location: loc,
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, report)
}

func (h *APIHandler) monthlyReport(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	user, err := h.auth.GetProfile(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusNotFound, "user not found")
		return
	}
	loc, err := h.resolveLocation(r, user)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid time_zone")
		return
	}
	monthParam := r.URL.Query().Get("month")
	if monthParam == "" {
		respondError(w, http.StatusBadRequest, "month is required")
		return
	}
	parsed, err := time.ParseInLocation("2006-01", monthParam, loc)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid month format")
		return
	}
	start := time.Date(parsed.Year(), parsed.Month(), 1, 0, 0, 0, 0, loc)
	if err := enforceReportWindow(start, loc); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	report, err := h.reports.Monthly(r.Context(), userID, usecase.ReportRange{
		Start:    start,
		End:      start.AddDate(0, 1, 0),
		Location: loc,
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, report)
}

// 補助構造体 ----------------------------------------------------------------

// ヘルパー ------------------------------------------------------------------

func buildEntryFilter(r *http.Request) (repository.EntryFilter, error) {
	filter, err := dto.BuildFilter(r.URL.Query().Get("from"), r.URL.Query().Get("to"))
	if err != nil {
		return repository.EntryFilter{}, err
	}
	var projectID *uuid.UUID
	if v := r.URL.Query().Get("project_id"); v != "" {
		id, parseErr := uuid.Parse(v)
		if parseErr != nil {
			return repository.EntryFilter{}, dto.ValidationError{Field: "project_id", Message: "is invalid UUID"}
		}
		projectID = &id
	}
	var tagID *uuid.UUID
	if v := r.URL.Query().Get("tag_id"); v != "" {
		id, parseErr := uuid.Parse(v)
		if parseErr != nil {
			return repository.EntryFilter{}, dto.ValidationError{Field: "tag_id", Message: "is invalid UUID"}
		}
		tagID = &id
	}
	return repository.EntryFilter{From: filter.From, To: filter.To, ProjectID: projectID, TagID: tagID}, nil
}

func respondUsecaseError(w http.ResponseWriter, err error) {
	var valErr dto.ValidationError
	switch {
	case errors.As(err, &valErr):
		respondError(w, http.StatusBadRequest, valErr.Error())
	default:
		respondError(w, http.StatusBadRequest, err.Error())
	}
}

func normalizeWeekStart(t time.Time, loc *time.Location) time.Time {
	if loc == nil {
		loc = time.UTC
	}
	t = t.In(loc)
	t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	return t.AddDate(0, 0, -(weekday - 1))
}

func mapUser(user *entity.User) map[string]any {
	return map[string]any{
		"id":           user.ID,
		"email":        user.Email,
		"display_name": user.DisplayName,
		"time_zone":    user.TimeZone,
		"created_at":   user.CreatedAt,
	}
}

func respondJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]any{"error": message})
}

func generateCSRFToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func (h *APIHandler) setCSRFCookie(w http.ResponseWriter, token string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     middleware.CSRFCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: false,
		Secure:   h.cfg.SessionCookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(h.cfg.SessionTTL().Seconds()),
		Expires:  expiresAt,
	})
}

func (h *APIHandler) clearCSRFCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     middleware.CSRFCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: false,
		Secure:   h.cfg.SessionCookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})
}

func (h *APIHandler) resolveLocation(r *http.Request, user *entity.User) (*time.Location, error) {
	tz := strings.TrimSpace(r.URL.Query().Get("time_zone"))
	if tz == "" && user != nil {
		tz = strings.TrimSpace(user.TimeZone)
	}
	if tz == "" {
		tz = "UTC"
	}
	return time.LoadLocation(tz)
}

func enforceReportWindow(start time.Time, loc *time.Location) error {
	now := time.Now().In(loc)
	if start.Before(now.AddDate(0, 0, -maxReportLookbackDays)) {
		return errors.New("requested range is too far in the past")
	}
	if start.After(now.AddDate(0, 0, maxReportFutureDays)) {
		return errors.New("requested range is too far in the future")
	}
	return nil
}
