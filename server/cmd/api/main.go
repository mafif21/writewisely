// Command api is the WriteWisely HTTP API server.
//
// @title        WriteWisely API
// @version      0.1.0
// @description  AI-powered IELTS Writing checker (Task 1 Academic and Task 2).
// @BasePath     /api/v1
package main

//go:generate go tool swag init --dir ./,../../internal/platform/server,../../internal/feature --generalInfo main.go --output ../../docs --outputTypes go,json,yaml

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	_ "writewisely/server/docs" // registers the generated OpenAPI spec for /swagger
	"writewisely/server/internal/common/constants"
	"writewisely/server/internal/platform/config"
	"writewisely/server/internal/platform/database"
	"writewisely/server/internal/platform/server"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	if err := run(); err != nil {
		slog.Error("server exited with error", "err", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := database.NewPool(ctx, cfg.Database.DSN())
	if err != nil {
		return err
	}
	defer pool.Close()

	pingCtx, cancel := context.WithTimeout(ctx, constants.StartupPingTimeout)
	if err := pool.Ping(pingCtx); err != nil {
		slog.Error("database ping failed", "err", err)
	} else {
		slog.Info("database connected")
	}
	cancel()

	app := server.New(pool)

	errCh := make(chan error, 1)
	go func() {
		slog.Info("server listening", "port", cfg.Port, "env", cfg.AppEnv)
		errCh <- app.Listen(":" + cfg.Port)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		slog.Info("shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), constants.ShutdownTimeout)
		defer cancel()
		if err := app.ShutdownWithContext(shutdownCtx); err != nil && !errors.Is(err, context.Canceled) {
			return err
		}
		return nil
	}
}
