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

	"gorm.io/gorm"
)

const (
	databaseStartupAttempts = 12
	databaseStartupDelay    = 5 * time.Second
)

func main() {
	// 設定は環境変数から集約し、以降の層には Config として渡す。
	cfg := config.Load()

	if cfg.Environment == "production" && (cfg.SessionSecret == config.DefaultSessionSecret || len(cfg.SessionSecret) < 32) {
		log.Fatal("SESSION_SECRET must be provided and at least 32 characters long in production")
	}

	db, err := openDatabaseWithRetry(cfg)
	if err != nil {
		log.Fatalf("database startup failed: %v", err)
	}

	// セッションは HTTP 層の関心事なので、ユースケースには渡さず handler 側で扱う。
	sessionStore, err := sess.NewSignedCookieStore(cfg.SessionSecret)
	if err != nil {
		log.Fatalf("failed to initialize session store: %v", err)
	}

	// リポジトリ
	userRepo := gormrepo.NewUserRepository(db)
	projectRepo := gormrepo.NewProjectRepository(db)
	entryRepo := gormrepo.NewEntryRepository(db)
	tagRepo := gormrepo.NewTagRepository(db)
	allocationRepo := gormrepo.NewAllocationRepository(db)

	// ユースケース
	// ユースケースは repository interface に依存し、DB 実装の詳細を知らない。
	authUC := usecase.NewAuthUsecase(userRepo)
	projectUC := usecase.NewProjectUsecase(projectRepo, cfg)
	tagUC := usecase.NewTagUsecase(tagRepo, cfg)
	entryUC := usecase.NewEntryUsecase(entryRepo, tagRepo, infTime.SystemClock{})
	reportUC := usecase.NewReportUsecase(entryRepo, projectRepo)
	allocationUC := usecase.NewAllocationUsecase(allocationRepo, infTime.SystemClock{})

	apiHandler := handler.NewAPIHandler(cfg, sessionStore, authUC, projectUC, tagUC, entryUC, reportUC, allocationUC)

	// HTTP サーバーは chi ルーターを入口にし、各 request を handler -> usecase へ流す。
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

	// SIGINT/SIGTERM 受信時は処理中の request を短時間待ってから終了する。
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
	log.Println("server stopped")
}

func openDatabaseWithRetry(cfg config.Config) (*gorm.DB, error) {
	var lastErr error
	for attempt := 1; attempt <= databaseStartupAttempts; attempt++ {
		db, err := database.Open(cfg)
		if err != nil {
			lastErr = err
		} else if err := database.Automigrate(db); err != nil {
			lastErr = err
		} else {
			return db, nil
		}

		if attempt == databaseStartupAttempts {
			break
		}
		log.Printf(
			"database startup attempt %d/%d failed: %v; retrying in %s",
			attempt,
			databaseStartupAttempts,
			lastErr,
			databaseStartupDelay,
		)
		time.Sleep(databaseStartupDelay)
	}
	return nil, lastErr
}
