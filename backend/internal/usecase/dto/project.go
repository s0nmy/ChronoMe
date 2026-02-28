package dto

import (
	"strings"
)

// ProjectCreateRequest は作成リクエストの入力を表す。
type ProjectCreateRequest struct {
	Name        string `json:"name"`
	Color       string `json:"color"`
	Description string `json:"description"`
}

// Normalize は検証して整形済みフィールドを返す。
func (r ProjectCreateRequest) Normalize(defaultColor string) (ProjectInput, error) {
	name := strings.TrimSpace(r.Name)
	if name == "" {
		return ProjectInput{}, ValidationError{Field: "name", Message: "is required"}
	}
	color := strings.TrimSpace(r.Color)
	if color == "" {
		color = defaultColor
	}
	return ProjectInput{
		Name:        name,
		Color:       color,
		Description: strings.TrimSpace(r.Description),
	}, nil
}

// ProjectUpdateRequest は部分更新を扱う。
type ProjectUpdateRequest struct {
	Name        *string `json:"name"`
	Color       *string `json:"color"`
	Description *string `json:"description"`
	IsArchived  *bool   `json:"is_archived"`
}

// Normalize はトリム済み値を保証する。
func (r ProjectUpdateRequest) Normalize() (ProjectUpdateInput, error) {
	if r.Name != nil {
		trimmed := strings.TrimSpace(*r.Name)
		if trimmed == "" {
			return ProjectUpdateInput{}, ValidationError{Field: "name", Message: "is required"}
		}
		r.Name = &trimmed
	}
	if r.Color != nil {
		trimmed := strings.TrimSpace(*r.Color)
		if trimmed == "" {
			return ProjectUpdateInput{}, ValidationError{Field: "color", Message: "is required"}
		}
		r.Color = &trimmed
	}
	if r.Description != nil {
		trimmed := strings.TrimSpace(*r.Description)
		r.Description = &trimmed
	}
	return ProjectUpdateInput{
		Name:        r.Name,
		Color:       r.Color,
		Description: r.Description,
		IsArchived:  r.IsArchived,
	}, nil
}

// ProjectInput は正規化済みの表現。
type ProjectInput struct {
	Name        string
	Color       string
	Description string
}

// ProjectUpdateInput は任意更新を表す。
type ProjectUpdateInput struct {
	Name        *string
	Color       *string
	Description *string
	IsArchived  *bool
}
