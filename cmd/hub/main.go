package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/avozda/global-latency-tracker/internal/hub/handlers"
	"github.com/avozda/global-latency-tracker/internal/hub/tools"

	"github.com/go-chi/chi"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetReportCaller(true)

	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Fatal(err)
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}

	if err := tools.RunMigrations(dsn); err != nil {
		log.Errorf("apply migrations: %v", err)
	} else {
		log.Info("Database migrations applied")
	}

	db, err := tools.OpenDatabase(dsn)
	if err != nil {
		log.Errorf("open database: %v", err)
	}
	if db != nil {
		defer db.Close()
	}

	retention, cleanupInterval, err := retentionConfig()
	if err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	retentionDone := startProbeResultsRetention(ctx, db, retention, cleanupInterval)

	var r chi.Router = chi.NewRouter()
	handlers.RegisterRoutes(r, db)

	srv := &http.Server{
		Addr:              ":" + port(),
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	idleClosed := make(chan struct{})
	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Error(err)
		}
		close(idleClosed)
	}()

	log.Info("Starting server on ", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}

	<-idleClosed
	<-retentionDone
	log.Info("Server stopped")
}

func port() string {
	if p := os.Getenv("PORT"); p != "" {
		return p
	}
	return "8080"
}
