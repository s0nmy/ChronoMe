package dto

import "strings"

// TagCreateRequest validates tag creation payload.
type TagCreateRequest struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

// Normalize ensures defaults and validation.
func (r TagCreateRequest) Normalize(defaultColor string) (TagInput, error) {
	name := strings.TrimSpace(r.Name)
	if name == "" {
		return TagInput{}, ValidationError{Field: "name", Message: "is required"}
	}
	color := strings.TrimSpace(r.Color)
	if color == "" {
		color = defaultColor
	}
	if len(color) != 7 || color[0] != '#' {
		return TagInput{}, ValidationError{Field: "color", Message: "must be #RRGGBB"}
	}
	return TagInput{Name: name, Color: color}, nil
}

// TagUpdateRequest handles partial updates.
type TagUpdateRequest struct {
	Name  *string `json:"name"`
	Color *string `json:"color"`
}

// Normalize trims and validates update fields.
func (r TagUpdateRequest) Normalize() (TagUpdateInput, error) {
	if r.Name != nil {
		trimmed := strings.TrimSpace(*r.Name)
		if trimmed == "" {
			return TagUpdateInput{}, ValidationError{Field: "name", Message: "is required"}
		}
		r.Name = &trimmed
	}
	if r.Color != nil {
		trimmed := strings.TrimSpace(*r.Color)
		if trimmed == "" {
			return TagUpdateInput{}, ValidationError{Field: "color", Message: "is required"}
		}
		if len(trimmed) != 7 || trimmed[0] != '#' {
			return TagUpdateInput{}, ValidationError{Field: "color", Message: "must be #RRGGBB"}
		}
		r.Color = &trimmed
	}
	return TagUpdateInput{Name: r.Name, Color: r.Color}, nil
}

// TagInput is normalized payload.
type TagInput struct {
	Name  string
	Color string
}

// TagUpdateInput stores optional fields.
type TagUpdateInput struct {
	Name  *string
	Color *string
}
