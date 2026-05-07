package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sekigo/linkforge/internal/config"
	httpapi "github.com/sekigo/linkforge/internal/http"
	"github.com/sekigo/linkforge/internal/service"
	"github.com/sekigo/linkforge/internal/storage"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if err := run(logger); err != nil {
		logger.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	db, err := storage.NewPostgres(ctx, cfg.PostgresDSN)
	if err != nil {
		return err
	}
	defer db.Close()

	cache, err := storage.NewCache(ctx, cfg.RedisAddr)
	if err != nil {
		return err
	}
	defer cache.Close()

	shortener := service.NewShortener(db, cache, cfg.BaseURL)

	srv := httpapi.NewServer(cfg.HTTPAddr, shortener, logger)

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("http server listening", "addr", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	case err := <-serverErr:
		return err
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	return srv.Shutdown(shutdownCtx)
}
