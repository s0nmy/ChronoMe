package dto

import "strings"

// TagCreateRequest はタグ作成ペイロードを検証する。
type TagCreateRequest struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

// Normalize はデフォルト設定と検証を行う。
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

// TagUpdateRequest は部分更新を扱う。
type TagUpdateRequest struct {
	Name  *string `json:"name"`
	Color *string `json:"color"`
}

// Normalize は更新フィールドをトリムして検証する。
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

// TagInput は正規化済みペイロード。
type TagInput struct {
	Name  string
	Color string
}

// TagUpdateInput は任意フィールドを保持する。
type TagUpdateInput struct {
	Name  *string
	Color *string
}
