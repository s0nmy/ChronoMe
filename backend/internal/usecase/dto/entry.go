package dto

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// EntryCreateRequest はエントリ作成の JSON ペイロードを受け取る。
type EntryCreateRequest struct {
	Title     string   `json:"title"`
	Notes     string   `json:"notes"`
	ProjectID *string  `json:"project_id"`
	StartedAt *string  `json:"started_at"`
	EndedAt   *string  `json:"ended_at"`
	IsBreak   *bool    `json:"is_break"`
	Ratio     *float64 `json:"ratio"`
	TagIDs    []string `json:"tag_ids"`
}

// EntryCreateData はユースケースで使う正規化データ。
type EntryCreateData struct {
	Title     string
	Notes     string
	ProjectID *uuid.UUID
	StartedAt *time.Time
	EndedAt   *time.Time
	IsBreak   bool
	Ratio     float64
	TagIDs    []uuid.UUID
}

// Normalize はリクエストを検証し型付けデータへ変換する。
func (r EntryCreateRequest) Normalize() (EntryCreateData, error) {
	title := strings.TrimSpace(r.Title)
	if title == "" {
		return EntryCreateData{}, ValidationError{Field: "title", Message: "is required"}
	}
	projectID, err := parseUUIDPtr(r.ProjectID, "project_id")
	if err != nil {
		return EntryCreateData{}, err
	}
	startedAt, err := parseTimePtr(r.StartedAt, "started_at")
	if err != nil {
		return EntryCreateData{}, err
	}
	endedAt, err := parseTimePtr(r.EndedAt, "ended_at")
	if err != nil {
		return EntryCreateData{}, err
	}
	isBreak := false
	if r.IsBreak != nil {
		isBreak = *r.IsBreak
	}
	ratio := 1.0
	if r.Ratio != nil {
		if *r.Ratio <= 0 {
			return EntryCreateData{}, ValidationError{Field: "ratio", Message: "must be positive"}
		}
		ratio = *r.Ratio
	}
	tagIDs, err := parseUUIDList(r.TagIDs, "tag_ids")
	if err != nil {
		return EntryCreateData{}, err
	}
	return EntryCreateData{
		Title:     title,
		Notes:     r.Notes,
		ProjectID: projectID,
		StartedAt: startedAt,
		EndedAt:   endedAt,
		IsBreak:   isBreak,
		Ratio:     ratio,
		TagIDs:    tagIDs,
	}, nil
}

// EntryUpdateRequest は更新用のパッチペイロードを受け取る。
type EntryUpdateRequest struct {
	Title     *string   `json:"title"`
	Notes     *string   `json:"notes"`
	ProjectID *string   `json:"project_id"`
	StartedAt *string   `json:"started_at"`
	EndedAt   *string   `json:"ended_at"`
	IsBreak   *bool     `json:"is_break"`
	Ratio     *float64  `json:"ratio"`
	TagIDs    *[]string `json:"tag_ids"`
}

// EntryUpdateData は型付けされた正規化表現。
type EntryUpdateData struct {
	Title      *string
	Notes      *string
	ProjectID  *uuid.UUID
	StartedAt  *time.Time
	EndedAt    *time.Time
	EndedAtSet bool
	IsBreak    *bool
	Ratio      *float64
	TagIDs     []uuid.UUID
	TagIDsSet  bool
}

// Normalize はパッチデータを検証する。
func (r EntryUpdateRequest) Normalize() (EntryUpdateData, error) {
	if r.Title != nil {
		trimmed := strings.TrimSpace(*r.Title)
		if trimmed == "" {
			return EntryUpdateData{}, ValidationError{Field: "title", Message: "is required"}
		}
		r.Title = &trimmed
	}
	if r.Ratio != nil && *r.Ratio <= 0 {
		return EntryUpdateData{}, ValidationError{Field: "ratio", Message: "must be positive"}
	}
	projectID, err := parseUUIDPtr(r.ProjectID, "project_id")
	if err != nil {
		return EntryUpdateData{}, err
	}
	startedAt, err := parseTimePtr(r.StartedAt, "started_at")
	if err != nil {
		return EntryUpdateData{}, err
	}
	endedAt, err := parseTimePtr(r.EndedAt, "ended_at")
	if err != nil {
		return EntryUpdateData{}, err
	}
	var tagIDs []uuid.UUID
	tagIDsSet := false
	if r.TagIDs != nil {
		tagIDsSet = true
		tagIDs, err = parseUUIDList(*r.TagIDs, "tag_ids")
		if err != nil {
			return EntryUpdateData{}, err
		}
	}
	return EntryUpdateData{
		Title:      r.Title,
		Notes:      r.Notes,
		ProjectID:  projectID,
		StartedAt:  startedAt,
		EndedAt:    endedAt,
		EndedAtSet: r.EndedAt != nil,
		IsBreak:    r.IsBreak,
		Ratio:      r.Ratio,
		TagIDs:     tagIDs,
		TagIDsSet:  tagIDsSet,
	}, nil
}

func parseUUIDPtr(raw *string, field string) (*uuid.UUID, error) {
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return nil, nil
	}
	id, err := uuid.Parse(strings.TrimSpace(*raw))
	if err != nil {
		return nil, ValidationError{Field: field, Message: "is invalid UUID"}
	}
	return &id, nil
}

func parseTimePtr(raw *string, field string) (*time.Time, error) {
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, strings.TrimSpace(*raw))
	if err != nil {
		return nil, ValidationError{Field: field, Message: "must be RFC3339"}
	}
	tt := t.UTC()
	return &tt, nil
}

// EntryFilter は一覧取得のクエリパラメータをまとめる（ISO 8601 文字列）。
type EntryFilter struct {
	From *time.Time
	To   *time.Time
}

// BuildFilter は文字列フィルタを検証して変換する。
func BuildFilter(fromRaw, toRaw string) (EntryFilter, error) {
	from, err := parseQueryTime(fromRaw, "from")
	if err != nil {
		return EntryFilter{}, err
	}
	to, err := parseQueryTime(toRaw, "to")
	if err != nil {
		return EntryFilter{}, err
	}
	if from != nil && to != nil && from.After(*to) {
		return EntryFilter{}, ValidationError{Field: "from", Message: "must be before to"}
	}
	return EntryFilter{From: from, To: to}, nil
}

func parseQueryTime(raw, field string) (*time.Time, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, strings.TrimSpace(raw))
	if err != nil {
		return nil, ValidationError{Field: field, Message: "must be RFC3339"}
	}
	tt := t.UTC()
	return &tt, nil
}

func parseUUIDList(raw []string, field string) ([]uuid.UUID, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	ids := make([]uuid.UUID, 0, len(raw))
	for _, item := range raw {
		if strings.TrimSpace(item) == "" {
			continue
		}
		parsed, err := uuid.Parse(strings.TrimSpace(item))
		if err != nil {
			return nil, ValidationError{Field: field, Message: "contains invalid UUID"}
		}
		ids = append(ids, parsed)
	}
	return ids, nil
}
