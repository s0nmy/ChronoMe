package usecase

import (
	"context"
	"sort"
	"time"

	"github.com/google/uuid"

	"chronome/internal/domain/entity"
	"chronome/internal/domain/repository"
)

// ReportUsecase はダッシュボード用にエントリ集計を行う。
const unassignedProjectKey = "unassigned"

type ReportUsecase struct {
	entries  repository.EntryRepository
	projects repository.ProjectRepository
}

// ReportRange はユーザーのローカル時間で期間を保持する。
type ReportRange struct {
	Start    time.Time
	End      time.Time
	Location *time.Location
}

func (r ReportRange) utcBounds() (time.Time, time.Time) {
	return r.Start.In(time.UTC), r.End.In(time.UTC)
}

func (r ReportRange) location() *time.Location {
	if r.Location != nil {
		return r.Location
	}
	return time.UTC
}

func (r ReportRange) dayCount() int {
	count := 0
	for d := r.Start; d.Before(r.End); d = d.AddDate(0, 0, 1) {
		count++
	}
	return count
}

type DailyReport struct {
	Date         string         `json:"date"`
	TotalSeconds int64          `json:"total_seconds"`
	Entries      []entity.Entry `json:"entries"`
}

type ReportDay struct {
	Date         string `json:"date"`
	TotalSeconds int64  `json:"total_seconds"`
}

type WeeklyReport struct {
	WeekStart    string             `json:"week_start"`
	TotalSeconds int64              `json:"total_seconds"`
	Days         []ReportDay        `json:"days"`
	Projects     []ProjectBreakdown `json:"projects"`
	Tags         []TagBreakdown     `json:"tags"`
}

type MonthlyReport struct {
	Month        string             `json:"month"`
	TotalSeconds int64              `json:"total_seconds"`
	Days         []ReportDay        `json:"days"`
	Weeks        []ReportWeek       `json:"weeks"`
	Projects     []ProjectBreakdown `json:"projects"`
	Tags         []TagBreakdown     `json:"tags"`
	DaysInMonth  int                `json:"days_in_month"`
}

type ReportWeek struct {
	WeekStart    string `json:"week_start"`
	TotalSeconds int64  `json:"total_seconds"`
}

type ProjectBreakdown struct {
	ProjectID    *uuid.UUID `json:"project_id,omitempty"`
	Name         string     `json:"name"`
	Color        string     `json:"color"`
	TotalSeconds int64      `json:"total_seconds"`
}

type TagBreakdown struct {
	TagID        uuid.UUID `json:"tag_id"`
	Name         string    `json:"name"`
	Color        string    `json:"color"`
	TotalSeconds int64     `json:"total_seconds"`
}

func NewReportUsecase(entries repository.EntryRepository, projects repository.ProjectRepository) *ReportUsecase {
	return &ReportUsecase{entries: entries, projects: projects}
}

func (u *ReportUsecase) Daily(ctx context.Context, userID uuid.UUID, rr ReportRange) (DailyReport, error) {
	from, to := rr.utcBounds()
	entries, err := u.entries.ListByUser(ctx, userID, repository.EntryFilter{From: &from, To: &to})
	if err != nil {
		return DailyReport{}, err
	}
	var total int64
	for _, entry := range entries {
		total += entry.DurationSec
	}
	return DailyReport{
		Date:         rr.Start.Format("2006-01-02"),
		TotalSeconds: total,
		Entries:      entries,
	}, nil
}

func (u *ReportUsecase) Weekly(ctx context.Context, userID uuid.UUID, rr ReportRange) (WeeklyReport, error) {
	from, to := rr.utcBounds()
	entries, err := u.entries.ListByUser(ctx, userID, repository.EntryFilter{From: &from, To: &to})
	if err != nil {
		return WeeklyReport{}, err
	}
	total := int64(0)
	dayTotals := map[string]int64{}
	loc := rr.location()
	projectTotals := make(map[uuid.UUID]int64)
	var unassignedTotal int64
	tagTotals := make(map[uuid.UUID]int64)
	tagMeta := make(map[uuid.UUID]entity.Tag)
	for _, entry := range entries {
		localStarted := entry.StartedAt.In(loc)
		if localStarted.Before(rr.Start) || !localStarted.Before(rr.End) {
			continue
		}
		dayKey := localStarted.Format("2006-01-02")
		dayTotals[dayKey] += entry.DurationSec
		total += entry.DurationSec
		if entry.ProjectID == nil {
			unassignedTotal += entry.DurationSec
		} else {
			projectTotals[*entry.ProjectID] += entry.DurationSec
		}
		for _, tag := range entry.Tags {
			tagTotals[tag.ID] += entry.DurationSec
			if _, ok := tagMeta[tag.ID]; !ok {
				tagMeta[tag.ID] = tag
			}
		}
	}
	days := make([]ReportDay, 7)
	for i := 0; i < 7; i++ {
		day := rr.Start.AddDate(0, 0, i)
		key := day.Format("2006-01-02")
		days[i] = ReportDay{Date: key, TotalSeconds: dayTotals[key]}
	}
	projectBreakdown := u.buildProjectBreakdown(ctx, userID, projectTotals, unassignedTotal)
	tagBreakdown := buildTagBreakdown(tagTotals, tagMeta)
	return WeeklyReport{
		WeekStart:    rr.Start.Format("2006-01-02"),
		TotalSeconds: total,
		Days:         days,
		Projects:     projectBreakdown,
		Tags:         tagBreakdown,
	}, nil
}

func (u *ReportUsecase) Monthly(ctx context.Context, userID uuid.UUID, rr ReportRange) (MonthlyReport, error) {
	from, to := rr.utcBounds()
	entries, err := u.entries.ListByUser(ctx, userID, repository.EntryFilter{From: &from, To: &to})
	if err != nil {
		return MonthlyReport{}, err
	}
	daysInMonth := rr.dayCount()
	dayTotals := make([]int64, daysInMonth)
	weekTotals := map[string]int64{}
	projectTotals := make(map[uuid.UUID]int64)
	var unassignedTotal int64
	total := int64(0)
	tagTotals := make(map[uuid.UUID]int64)
	tagMeta := make(map[uuid.UUID]entity.Tag)
	loc := rr.location()
	for _, entry := range entries {
		localDate := entry.StartedAt.In(loc)
		if localDate.Before(rr.Start) || !localDate.Before(rr.End) {
			continue
		}
		dayIndex := int(localDate.Sub(rr.Start).Hours() / 24)
		if dayIndex >= 0 && dayIndex < len(dayTotals) {
			dayTotals[dayIndex] += entry.DurationSec
		}
		total += entry.DurationSec
		weekStart := startOfWeek(localDate)
		key := weekStart.Format("2006-01-02")
		weekTotals[key] += entry.DurationSec
		if entry.ProjectID == nil {
			unassignedTotal += entry.DurationSec
		} else {
			projectTotals[*entry.ProjectID] += entry.DurationSec
		}
		for _, tag := range entry.Tags {
			tagTotals[tag.ID] += entry.DurationSec
			if _, ok := tagMeta[tag.ID]; !ok {
				tagMeta[tag.ID] = tag
			}
		}
	}
	days := make([]ReportDay, daysInMonth)
	for i := 0; i < daysInMonth; i++ {
		day := rr.Start.AddDate(0, 0, i)
		days[i] = ReportDay{
			Date:         day.Format("2006-01-02"),
			TotalSeconds: dayTotals[i],
		}
	}
	var weeks []ReportWeek
	for key, value := range weekTotals {
		weeks = append(weeks, ReportWeek{WeekStart: key, TotalSeconds: value})
	}
	projectBreakdown := u.buildProjectBreakdown(ctx, userID, projectTotals, unassignedTotal)
	tagBreakdown := buildTagBreakdown(tagTotals, tagMeta)
	sort.Slice(weeks, func(i, j int) bool {
		return weeks[i].WeekStart < weeks[j].WeekStart
	})
	return MonthlyReport{
		Month:        rr.Start.Format("2006-01"),
		TotalSeconds: total,
		Days:         days,
		Weeks:        weeks,
		Projects:     projectBreakdown,
		Tags:         tagBreakdown,
		DaysInMonth:  daysInMonth,
	}, nil
}

func startOfWeek(t time.Time) time.Time {
	t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	return t.AddDate(0, 0, -(weekday - 1))
}

func (u *ReportUsecase) buildProjectBreakdown(ctx context.Context, userID uuid.UUID, totals map[uuid.UUID]int64, unassigned int64) []ProjectBreakdown {
	meta := make(map[uuid.UUID]entity.Project)
	if projects, err := u.projects.ListByUser(ctx, userID); err == nil {
		for _, project := range projects {
			meta[project.ID] = project
		}
	}
	var breakdown []ProjectBreakdown
	for id, total := range totals {
		proj := meta[id]
		name := proj.Name
		if name == "" {
			name = "Deleted project"
		}
		idCopy := id
		breakdown = append(breakdown, ProjectBreakdown{
			ProjectID:    &idCopy,
			Name:         name,
			Color:        proj.Color,
			TotalSeconds: total,
		})
	}
	if unassigned > 0 {
		breakdown = append(breakdown, ProjectBreakdown{
			Name:         "Unassigned",
			Color:        "",
			TotalSeconds: unassigned,
		})
	}
	sort.Slice(breakdown, func(i, j int) bool {
		return breakdown[i].Name < breakdown[j].Name
	})
	return breakdown
}

func buildTagBreakdown(totals map[uuid.UUID]int64, meta map[uuid.UUID]entity.Tag) []TagBreakdown {
	var breakdown []TagBreakdown
	for id, total := range totals {
		tag := meta[id]
		breakdown = append(breakdown, TagBreakdown{
			TagID:        id,
			Name:         tag.Name,
			Color:        tag.Color,
			TotalSeconds: total,
		})
	}
	sort.Slice(breakdown, func(i, j int) bool {
		return breakdown[i].Name < breakdown[j].Name
	})
	return breakdown
}
