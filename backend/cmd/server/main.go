package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"chronome/internal/adapter/db/gormrepo"
	"chronome/internal/adapter/http/handler"
	"chronome/internal/adapter/infra/config"
	"chronome/internal/adapter/infra/database"
	sess "chronome/internal/adapter/infra/session"
	infTime "chronome/internal/adapter/infra/time"
	"chronome/internal/usecase"
)

func main() {
	cfg := config.Load()

	if cfg.Environment == "production" && (cfg.SessionSecret == config.DefaultSessionSecret || len(cfg.SessionSecret) < 32) {
		log.Fatal("SESSION_SECRET must be provided and at least 32 characters long in production")
	}

	db, err := database.Open(cfg)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	if err := database.Automigrate(db); err != nil {
		log.Fatalf("automigrate failed: %v", err)
	}

	sessionStore, err := sess.NewSignedCookieStore(cfg.SessionSecret)
	if err != nil {
		log.Fatalf("failed to initialize session store: %v", err)
	}

	// リポジトリ
	userRepo := gormrepo.NewUserRepository(db)
	projectRepo := gormrepo.NewProjectRepository(db)
	entryRepo := gormrepo.NewEntryRepository(db)
	tagRepo := gormrepo.NewTagRepository(db)

	// ユースケース
	authUC := usecase.NewAuthUsecase(userRepo)
	projectUC := usecase.NewProjectUsecase(projectRepo, cfg)
	tagUC := usecase.NewTagUsecase(tagRepo, cfg)
	entryUC := usecase.NewEntryUsecase(entryRepo, tagRepo, infTime.SystemClock{})
	reportUC := usecase.NewReportUsecase(entryRepo, projectRepo)

	apiHandler := handler.NewAPIHandler(cfg, sessionStore, authUC, projectUC, tagUC, entryUC, reportUC)

	srv := &http.Server{
		Addr:    cfg.Address,
		Handler: apiHandler.Router(),
	}

	go func() {
		log.Printf("ChronoMe backend listening on %s", cfg.Address)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	<-shutdown

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
	log.Println("server stopped")
}
