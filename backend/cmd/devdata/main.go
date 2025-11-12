package main

import (
	"context"
	"errors"
	"log"

	"gorm.io/gorm"

	"chronome/internal/adapter/db/gormrepo"
	"chronome/internal/adapter/infra/config"
	"chronome/internal/adapter/infra/database"
	"chronome/internal/usecase"
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

	if _, err := userRepo.GetByEmail(ctx, seedEmail); err == nil {
		log.Printf("seed user %s already exists; skipping\n", seedEmail)
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Fatalf("checking user failed: %v", err)
	}

	authUC := usecase.NewAuthUsecase(userRepo)
	_, err = authUC.Signup(ctx, usecase.SignupParams{
		Email:       seedEmail,
		Password:    seedPassword,
		DisplayName: "Administrator",
		TimeZone:    "UTC",
	})
	if err != nil {
		log.Fatalf("failed to create seed user: %v", err)
	}

	log.Printf("created seed user %s with password %q\n", seedEmail, seedPassword)
}
