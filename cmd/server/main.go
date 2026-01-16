package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Himany/GophKeeper/internal/config"
	"github.com/Himany/GophKeeper/internal/server"
	"github.com/Himany/GophKeeper/internal/storage/postgres"
	"go.uber.org/zap"
)

var (
	version   = "dev"
	buildTime = "unknown"
)

func main() {
	logger := initLogger()
	defer logger.Sync()

	logger.Info("Starting GophKeeper server",
		zap.String("version", version),
		zap.String("buildTime", buildTime))

	cfg, err := config.LoadServerConfig()
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	storage, err := postgres.NewPostgresStorage(context.Background(), cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	if err := storage.Ping(context.Background()); err != nil {
		logger.Fatal("Failed to ping database", zap.Error(err))
	}
	logger.Info("Database connection established successfully")

	srv, err := server.New(cfg, storage, logger)
	if err != nil {
		logger.Fatal("Failed to create server", zap.Error(err))
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		var err error
		if cfg.TLSCert != "" && cfg.TLSKey != "" {
			logger.Info("Starting HTTPS server with TLS")
			err = srv.StartTLS(cfg.TLSCert, cfg.TLSKey)
		} else {
			logger.Info("Starting HTTP server")
			err = srv.Start()
		}

		if err != nil {
			logger.Error("Server failed to start", zap.Error(err))
			quit <- syscall.SIGTERM
		}
	}()

	<-quit
	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Stop(ctx); err != nil {
		logger.Error("Server shutdown failed", zap.Error(err))
		os.Exit(1)
	}

	logger.Info("Server exited")
}

func initLogger() *zap.Logger {
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	var logger *zap.Logger
	var err error

	if logLevel == "debug" {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}

	if err != nil {
		logger = zap.NewNop()
	}

	return logger
}
