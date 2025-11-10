package dto

import (
	"strings"
)

// ProjectCreateRequest describes the incoming body for create.
type ProjectCreateRequest struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

// Normalize validates and returns cleaned fields.
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
		Name:  name,
		Color: color,
	}, nil
}

// ProjectUpdateRequest handles partial updates.
type ProjectUpdateRequest struct {
	Name       *string `json:"name"`
	Color      *string `json:"color"`
	IsArchived *bool   `json:"is_archived"`
}

// Normalize ensures trimmed values.
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
	return ProjectUpdateInput{
		Name:       r.Name,
		Color:      r.Color,
		IsArchived: r.IsArchived,
	}, nil
}

// ProjectInput is a normalized representation.
type ProjectInput struct {
	Name  string
	Color string
}

// ProjectUpdateInput represents optional updates.
type ProjectUpdateInput struct {
	Name       *string
	Color      *string
	IsArchived *bool
}
