package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/config"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/logger"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/router"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/storage/db"
)

func run() error {
	log := logger.Init()

	cfg := config.Load()

	database, err := db.NewPostgresConnection(cfg.DB)
	if err != nil {
		log.Error("Failed to connect to database", "error", err)
		return err
	}
	defer database.Close()

	log.Info("Connected to database successfully")

	addr := ":" + cfg.Server.Port
	srvRouter, err := router.New(cfg, database)
	if err != nil {
		log.Error("Failed to init repo", "error", err)
		return err
	}

	srv := &http.Server{
		Handler: srvRouter,
		Addr:    addr,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	serverErrors := make(chan error, 1)

	go func() {
		log.Info("Server started", "host", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Server failed", "error", err)
			serverErrors <- err
		}
	}()

	select {
	case <-stop:
		log.Info("Shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Error("Server forced to shutdown", "error", err)
			return err
		}

		log.Info("Server stopped gracefully")
		return nil
	case err := <-serverErrors:
		return err
	}
}

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}
