package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Himany/GophKeeper/internal/auth"
	"github.com/Himany/GophKeeper/internal/config"
	"github.com/Himany/GophKeeper/internal/crypto"
	"github.com/Himany/GophKeeper/internal/server/handlers"
	"github.com/Himany/GophKeeper/internal/server/router"
	"github.com/Himany/GophKeeper/internal/storage"
	"go.uber.org/zap"
)

type Server struct {
	httpServer *http.Server
	logger     *zap.Logger
	storage    storage.Storage
}

func New(cfg *config.ServerConfig, storage storage.Storage, logger *zap.Logger) (*Server, error) {
	jwtManager := auth.NewJWTManager(cfg.JWTSecret, cfg.JWTExpiry)

	encryptor, err := crypto.NewEncryptor(cfg.EncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create encryptor: %w", err)
	}

	handler := handlers.NewHandler(storage, jwtManager, encryptor, logger)

	httpRouter := router.NewRouter(handler, jwtManager, logger)

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	httpServer := &http.Server{
		Addr:           addr,
		Handler:        httpRouter,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	return &Server{
		httpServer: httpServer,
		logger:     logger,
		storage:    storage,
	}, nil
}

func (s *Server) Start() error {
	s.logger.Info("Starting HTTP server", zap.String("addr", s.httpServer.Addr))

	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}

	return nil
}

func (s *Server) StartTLS(certFile, keyFile string) error {
	s.logger.Info("Starting HTTPS server", zap.String("addr", s.httpServer.Addr))

	if err := s.httpServer.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start HTTPS server: %w", err)
	}

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Stopping HTTP server")

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown HTTP server: %w", err)
	}

	if err := s.storage.Close(); err != nil {
		s.logger.Error("Failed to close storage", zap.Error(err))
	}

	s.logger.Info("Server stopped successfully")
	return nil
}
