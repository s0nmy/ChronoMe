package usecase

import (
	"context"
	"errors"
	"math"
	"sort"
	"time"

	"github.com/google/uuid"

	"chronome/internal/domain/entity"
	"chronome/internal/domain/repository"
	"chronome/internal/usecase/dto"
)

const allocationEpsilon = 1e-9

// AllocationConstraintError は制約違反の分配失敗を表す。
type AllocationConstraintError struct {
	Message string
}

func (e AllocationConstraintError) Error() string {
	return e.Message
}

// AllocationUsecase は分配ロジックと永続化を担当する。
type AllocationUsecase struct {
	repo repository.AllocationRepository
}

func NewAllocationUsecase(repo repository.AllocationRepository) *AllocationUsecase {
	return &AllocationUsecase{repo: repo}
}

// AllocationResult は API に返す結果。
type AllocationResult struct {
	RequestID    uuid.UUID        `json:"request_id"`
	TotalMinutes int              `json:"total_minutes"`
	Allocations  []AllocationItem `json:"allocations"`
}

// AllocationItem は分配結果の1行。
type AllocationItem struct {
	TaskID           string  `json:"task_id"`
	Ratio            float64 `json:"ratio"`
	AllocatedMinutes int     `json:"allocated_minutes"`
}

type allocationDistribution struct {
	TaskID           string
	Ratio            float64
	AllocatedMinutes int
	MinMinutes       *int
	MaxMinutes       *int
}

// Allocate は分配計算と保存を行う。
func (u *AllocationUsecase) Allocate(ctx context.Context, input dto.AllocationRequest) (AllocationResult, error) {
	data, err := input.Normalize()
	if err != nil {
		return AllocationResult{}, err
	}

	allocations, err := distributeAllocations(data)
	if err != nil {
		return AllocationResult{}, AllocationConstraintError{Message: err.Error()}
	}

	requestID := uuid.New()
	now := time.Now().UTC()
	request := &entity.AllocationRequest{
		ID:           requestID,
		TotalMinutes: data.TotalMinutes,
		CreatedAt:    now,
	}
	allocationEntities := make([]entity.TaskAllocation, 0, len(allocations))
	responseItems := make([]AllocationItem, 0, len(allocations))
	for _, allocation := range allocations {
		allocationEntities = append(allocationEntities, entity.TaskAllocation{
			RequestID:        requestID,
			TaskID:           allocation.TaskID,
			Ratio:            allocation.Ratio,
			AllocatedMinutes: allocation.AllocatedMinutes,
			MinMinutes:       allocation.MinMinutes,
			MaxMinutes:       allocation.MaxMinutes,
			CreatedAt:        now,
			UpdatedAt:        now,
		})
		responseItems = append(responseItems, AllocationItem{
			TaskID:           allocation.TaskID,
			Ratio:            allocation.Ratio,
			AllocatedMinutes: allocation.AllocatedMinutes,
		})
	}

	if err := u.repo.Create(ctx, request, allocationEntities); err != nil {
		return AllocationResult{}, err
	}

	return AllocationResult{
		RequestID:    requestID,
		TotalMinutes: data.TotalMinutes,
		Allocations:  responseItems,
	}, nil
}

type allocationState struct {
	TaskID     string
	Ratio      float64
	MinMinutes int
	MaxMinutes *int
	Allocation int
	Remainder  float64
	Normalized float64
	Index      int
}

func distributeAllocations(input dto.AllocationRequestData) ([]allocationDistribution, error) {
	if len(input.Tasks) == 0 {
		return nil, errors.New("at least one task is required")
	}
	ratioSum := 0.0
	for _, task := range input.Tasks {
		ratioSum += task.Ratio
	}
	if ratioSum <= 0 {
		return nil, errors.New("sum of ratios must be greater than zero")
	}

	states := make([]allocationState, 0, len(input.Tasks))
	allocated := 0
	maxSum := 0
	allMaxBounded := true
	for i, task := range input.Tasks {
		minMinutes := 0
		if task.MinMinutes != nil {
			minMinutes = *task.MinMinutes
		}
		if task.MaxMinutes != nil && minMinutes > *task.MaxMinutes {
			return nil, errors.New("min_minutes cannot exceed max_minutes")
		}
		allocated += minMinutes
		if task.MaxMinutes != nil {
			maxSum += *task.MaxMinutes
		} else {
			allMaxBounded = false
		}
		states = append(states, allocationState{
			TaskID:     task.TaskID,
			Ratio:      task.Ratio,
			MinMinutes: minMinutes,
			MaxMinutes: task.MaxMinutes,
			Allocation: minMinutes,
			Remainder:  0,
			Normalized: task.Ratio / ratioSum,
			Index:      i,
		})
	}
	if allocated > input.TotalMinutes {
		return nil, errors.New("total_minutes is smaller than the sum of min_minutes")
	}
	if allMaxBounded && input.TotalMinutes > maxSum {
		return nil, errors.New("total_minutes exceeds the sum of max_minutes")
	}

	remainingPool := input.TotalMinutes - allocated
	if remainingPool == 0 {
		return mapAllocations(states), nil
	}

	carried := 0
	for i := range states {
		task := &states[i]
		desired := float64(remainingPool) * task.Normalized
		capacity := remainingPool
		if task.MaxMinutes != nil {
			capRemaining := *task.MaxMinutes - task.Allocation
			if capRemaining < 0 {
				capRemaining = 0
			}
			capacity = capRemaining
		}
		if capacity <= 0 {
			task.Remainder = -1
			continue
		}
		baseAdd := int(math.Floor(desired))
		if baseAdd > capacity {
			baseAdd = capacity
		}
		task.Allocation += baseAdd
		carried += baseAdd
		task.Remainder = desired - float64(baseAdd)
	}

	remaining := input.TotalMinutes - (allocated + carried)
	for remaining > 0 {
		eligible := make([]*allocationState, 0, len(states))
		for i := range states {
			task := &states[i]
			if task.MaxMinutes != nil && task.Allocation >= *task.MaxMinutes {
				continue
			}
			eligible = append(eligible, task)
		}
		if len(eligible) == 0 {
			return nil, errors.New("unable to satisfy max constraints with provided total_minutes")
		}
		sort.Slice(eligible, func(i, j int) bool {
			ai := eligible[i]
			aj := eligible[j]
			if math.Abs(aj.Remainder-ai.Remainder) > allocationEpsilon {
				return aj.Remainder > ai.Remainder
			}
			if math.Abs(aj.Normalized-ai.Normalized) > allocationEpsilon {
				return aj.Normalized > ai.Normalized
			}
			return ai.Index < aj.Index
		})

		if len(eligible) == 1 {
			task := eligible[0]
			available := remaining
			if task.MaxMinutes != nil {
				available = *task.MaxMinutes - task.Allocation
			}
			if available <= 0 {
				return nil, errors.New("unable to satisfy max constraints with provided total_minutes")
			}
			chunk := available
			if chunk > remaining {
				chunk = remaining
			}
			task.Allocation += chunk
			remaining -= chunk
			continue
		}

		distributed := 0
		for _, task := range eligible {
			if remaining == 0 {
				break
			}
			if task.MaxMinutes != nil && task.Allocation >= *task.MaxMinutes {
				continue
			}
			task.Allocation++
			remaining--
			distributed++
		}
		if distributed == 0 {
			return nil, errors.New("unable to distribute remaining minutes due to max constraints")
		}
	}

	return mapAllocations(states), nil
}

func mapAllocations(states []allocationState) []allocationDistribution {
	results := make([]allocationDistribution, 0, len(states))
	for _, task := range states {
		results = append(results, allocationDistribution{
			TaskID:           task.TaskID,
			Ratio:            task.Ratio,
			AllocatedMinutes: task.Allocation,
			MinMinutes:       task.MinMinutesPtr(),
			MaxMinutes:       task.MaxMinutes,
		})
	}
	return results
}

func (s allocationState) MinMinutesPtr() *int {
	if s.MinMinutes == 0 {
		return nil
	}
	value := s.MinMinutes
	return &value
}
