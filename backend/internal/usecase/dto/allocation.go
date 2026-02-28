package dto

import "strings"

// AllocationRequest は分配 API の入力ペイロードを表す。
type AllocationRequest struct {
	TotalMinutes int                     `json:"total_minutes"`
	Tasks        []AllocationTaskRequest `json:"tasks"`
}

// AllocationTaskRequest は各タスクの分配条件。
type AllocationTaskRequest struct {
	TaskID     string  `json:"task_id"`
	Ratio      float64 `json:"ratio"`
	MinMinutes *int    `json:"min_minutes,omitempty"`
	MaxMinutes *int    `json:"max_minutes,omitempty"`
}

// AllocationRequestData は正規化後の入力。
type AllocationRequestData struct {
	TotalMinutes int
	Tasks        []AllocationTaskData
}

// AllocationTaskData は正規化されたタスク情報。
type AllocationTaskData struct {
	TaskID     string
	Ratio      float64
	MinMinutes *int
	MaxMinutes *int
}

// Normalize は入力を検証して整形する。
func (r AllocationRequest) Normalize() (AllocationRequestData, error) {
	if r.TotalMinutes <= 0 {
		return AllocationRequestData{}, ValidationError{Field: "total_minutes", Message: "must be positive"}
	}
	if len(r.Tasks) == 0 {
		return AllocationRequestData{}, ValidationError{Field: "tasks", Message: "must include at least one task"}
	}
	seen := make(map[string]struct{}, len(r.Tasks))
	tasks := make([]AllocationTaskData, 0, len(r.Tasks))
	for _, task := range r.Tasks {
		id := strings.TrimSpace(task.TaskID)
		if id == "" {
			return AllocationRequestData{}, ValidationError{Field: "task_id", Message: "is required"}
		}
		if _, exists := seen[id]; exists {
			return AllocationRequestData{}, ValidationError{Field: "tasks", Message: "task_id must be unique"}
		}
		seen[id] = struct{}{}
		if task.Ratio <= 0 {
			return AllocationRequestData{}, ValidationError{Field: "ratio", Message: "must be positive"}
		}
		if task.MinMinutes != nil && *task.MinMinutes < 0 {
			return AllocationRequestData{}, ValidationError{Field: "min_minutes", Message: "must be >= 0"}
		}
		if task.MaxMinutes != nil && *task.MaxMinutes <= 0 {
			return AllocationRequestData{}, ValidationError{Field: "max_minutes", Message: "must be positive"}
		}
		if task.MinMinutes != nil && task.MaxMinutes != nil && *task.MinMinutes > *task.MaxMinutes {
			return AllocationRequestData{}, ValidationError{Field: "max_minutes", Message: "must be >= min_minutes"}
		}
		tasks = append(tasks, AllocationTaskData{
			TaskID:     id,
			Ratio:      task.Ratio,
			MinMinutes: task.MinMinutes,
			MaxMinutes: task.MaxMinutes,
		})
	}
	return AllocationRequestData{TotalMinutes: r.TotalMinutes, Tasks: tasks}, nil
}
