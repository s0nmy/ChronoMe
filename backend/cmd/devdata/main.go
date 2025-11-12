package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"chronome/internal/adapter/db/gormrepo"
	"chronome/internal/adapter/infra/config"
	"chronome/internal/adapter/infra/database"
	infTime "chronome/internal/adapter/infra/time"
	"chronome/internal/domain/entity"
	"chronome/internal/usecase"
	"chronome/internal/usecase/dto"
)

const (
	seedEmail    = "admin@example.com"
	seedPassword = "password"
)

func main() {
	ctx := context.Background()
	cfg := config.Load()

	db, err := database.Open(cfg)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	if err := database.Automigrate(db); err != nil {
		log.Fatalf("automigrate failed: %v", err)
	}

	userRepo := gormrepo.NewUserRepository(db)
	demoUser, err := ensureDemoUser(ctx, userRepo)
	if err != nil {
		log.Fatalf("failed to prepare demo user: %v", err)
	}

	if err := seedDemoData(ctx, db, demoUser); err != nil {
		log.Fatalf("failed to seed demo data: %v", err)
	}

	log.Printf("demo data ready for %s / %s\n", seedEmail, seedPassword)
}

func ensureDemoUser(ctx context.Context, userRepo *gormrepo.UserRepository) (*entity.User, error) {
	existing, err := userRepo.GetByEmail(ctx, seedEmail)
	switch {
	case err == nil:
		log.Printf("user %s already exists, refreshing demo data\n", seedEmail)
		return existing, nil
	case !errors.Is(err, gorm.ErrRecordNotFound):
		return nil, fmt.Errorf("lookup user: %w", err)
	default:
		authUC := usecase.NewAuthUsecase(userRepo)
		user, signupErr := authUC.Signup(ctx, usecase.SignupParams{
			Email:       seedEmail,
			Password:    seedPassword,
			DisplayName: "Administrator",
			TimeZone:    "Asia/Tokyo",
		})
		if signupErr != nil {
			return nil, signupErr
		}
		log.Printf("created seed user %s\n", seedEmail)
		return user, nil
	}
}

func seedDemoData(ctx context.Context, db *gorm.DB, user *entity.User) error {
	projectRepo := gormrepo.NewProjectRepository(db)
	tagRepo := gormrepo.NewTagRepository(db)
	entryRepo := gormrepo.NewEntryRepository(db)
	entryUC := usecase.NewEntryUsecase(entryRepo, tagRepo, infTime.SystemClock{})

	log.Println("cleaning previous demo data...")
	if err := db.Where("user_id = ?", user.ID).Delete(&entity.Entry{}).Error; err != nil {
		return fmt.Errorf("delete entries: %w", err)
	}
	if err := db.Where("user_id = ?", user.ID).Delete(&entity.Project{}).Error; err != nil {
		return fmt.Errorf("delete projects: %w", err)
	}
	if err := db.Where("user_id = ?", user.ID).Delete(&entity.Tag{}).Error; err != nil {
		return fmt.Errorf("delete tags: %w", err)
	}

	projectDefs := []struct {
		Name  string
		Color string
		Desc  string
	}{
		{"ChronoMe UI Revamp", "#3B82F6", "Dashboard / timer UI improvements"},
		{"API Stabilization", "#22C55E", "Backend refactors & performance"},
		{"Learning & Research", "#A855F7", "Personal skill-up time"},
		{"Customer Support", "#F97316", "User communication & fixes"},
	}
	projectIDs := make(map[string]uuid.UUID, len(projectDefs))
	for _, def := range projectDefs {
		project := &entity.Project{
			ID:          uuid.New(),
			UserID:      user.ID,
			Name:        def.Name,
			Color:       def.Color,
			Description: def.Desc,
		}
		if err := project.Validate(); err != nil {
			return err
		}
		if err := projectRepo.Create(ctx, project); err != nil {
			return fmt.Errorf("create project %s: %w", def.Name, err)
		}
		projectIDs[def.Name] = project.ID
	}

	tagDefs := []struct {
		Name  string
		Color string
	}{
		{"Design", "#F97316"},
		{"Implementation", "#2563EB"},
		{"Code Review", "#0EA5E9"},
		{"Learning", "#10B981"},
		{"Support", "#EF4444"},
	}
	tagIDs := make(map[string]uuid.UUID, len(tagDefs))
	for _, def := range tagDefs {
		tag := &entity.Tag{
			ID:     uuid.New(),
			UserID: user.ID,
			Name:   def.Name,
			Color:  def.Color,
		}
		if err := tag.Validate(); err != nil {
			return err
		}
		if err := tagRepo.Create(ctx, tag); err != nil {
			return fmt.Errorf("create tag %s: %w", def.Name, err)
		}
		tagIDs[def.Name] = tag.ID
	}

	loc, err := time.LoadLocation(defaultLocation(user.TimeZone))
	if err != nil {
		loc = time.UTC
	}
	now := time.Now().In(loc)
	startOfWeek := now.AddDate(0, 0, -int(now.Weekday())+1).Truncate(24 * time.Hour)

	entryDefs := []struct {
		Title       string
		ProjectName string
		DayOffset   int
		StartHour   int
		Duration    time.Duration
		Notes       string
		Tags        []string
	}{
		{"Dashboard レイアウト調整", "ChronoMe UI Revamp", 0, 9, 2 * time.Hour, "Sidebar nav polishing", []string{"Design"}},
		{"タイマー動作の最適化", "ChronoMe UI Revamp", 0, 11, 90 * time.Minute, "Fixed pause/resume glitch", []string{"Implementation"}},
		{"API レスポンス監視", "API Stabilization", -1, 14, 2 * time.Hour, "Tracing slow queries", []string{"Implementation", "Code Review"}},
		{"OAuth 認証コードレビュー", "API Stabilization", -2, 10, 90 * time.Minute, "Reviewed PR#142", []string{"Code Review"}},
		{"Clean Architecture 読書", "Learning & Research", -3, 16, 2 * time.Hour, "Chapter 5-6 notes", []string{"Learning"}},
	}

	for _, def := range entryDefs {
		projectID, ok := projectIDs[def.ProjectName]
		if !ok {
			return fmt.Errorf("unknown project %s", def.ProjectName)
		}
		start := startOfWeek.AddDate(0, 0, def.DayOffset).Add(time.Duration(def.StartHour) * time.Hour)
		if err := createEntry(ctx, entryUC, user.ID, projectID, start, def.Duration, def.Title, def.Notes, def.Tags, tagIDs); err != nil {
			return err
		}
	}

	dailyTemplates := []struct {
		Project string
		Title   string
		Notes   string
		Start   int
		Tags    []string
	}{
		{"ChronoMe UI Revamp", "UI モック調整", "Tweaked command palette", 9, []string{"Design"}},
		{"API Stabilization", "API パフォーマンス改善", "Profiled /reports endpoints", 13, []string{"Implementation"}},
		{"Learning & Research", "技術調査", "Read Go generics articles", 16, []string{"Learning"}},
		{"Customer Support", "ユーザー問い合わせ対応", "Helped customers with exports", 11, []string{"Support"}},
	}

	for dayOffset := -21; dayOffset <= 0; dayOffset++ {
		for i, tmpl := range dailyTemplates {
			if dayOffset%3 == 0 && i > 2 {
				continue
			}
			projectID, ok := projectIDs[tmpl.Project]
			if !ok {
				continue
			}
			start := startOfWeek.AddDate(0, 0, dayOffset).Add(time.Duration(tmpl.Start+i) * time.Hour)
			duration := time.Duration(60+15*i) * time.Minute
			title := fmt.Sprintf("%s (%s)", tmpl.Title, start.Format("01/02"))
			if err := createEntry(ctx, entryUC, user.ID, projectID, start, duration, title, tmpl.Notes, tmpl.Tags, tagIDs); err != nil {
				return err
			}
		}

		if dayOffset%4 == 0 {
			projectID := projectIDs["ChronoMe UI Revamp"]
			start := startOfWeek.AddDate(0, 0, dayOffset).Add(20 * time.Hour)
			if err := createEntry(ctx, entryUC, user.ID, projectID, start, 45*time.Minute, "ナイトリービルド確認", "Ran nightly smoke tests", []string{"Implementation"}, tagIDs); err != nil {
				return err
			}
		}
	}

	log.Println("seeded demo projects, tags, and rich timeline entries.")
	return nil
}

func defaultLocation(tz string) string {
	if tz == "" {
		return "UTC"
	}
	return tz
}

func createEntry(ctx context.Context, uc *usecase.EntryUsecase, userID uuid.UUID, projectID uuid.UUID, start time.Time, duration time.Duration, title, notes string, tagNames []string, tagIDs map[string]uuid.UUID) error {
	end := start.Add(duration)
	startStr := start.Format(time.RFC3339)
	endStr := end.Format(time.RFC3339)
	projectIDStr := projectID.String()
	tagStrs := make([]string, 0, len(tagNames))
	for _, name := range tagNames {
		if id, ok := tagIDs[name]; ok {
			tagStrs = append(tagStrs, id.String())
		}
	}
	req := dto.EntryCreateRequest{
		Title:     title,
		Notes:     notes,
		ProjectID: &projectIDStr,
		StartedAt: &startStr,
		EndedAt:   &endStr,
		TagIDs:    tagStrs,
	}
	if _, err := uc.Create(ctx, userID, req); err != nil {
		return fmt.Errorf("create entry %s: %w", title, err)
	}
	return nil
}
