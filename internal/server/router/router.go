package router

import (
	"net/http"
	"time"

	"github.com/Himany/GophKeeper/internal/auth"
	"github.com/Himany/GophKeeper/internal/server/handlers"
	"github.com/Himany/GophKeeper/internal/server/middleware"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

func NewRouter(handler *handlers.Handler, jwtManager *auth.JWTManager, logger *zap.Logger) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RecoveryMiddleware(logger))
	r.Use(middleware.LoggerMiddleware(logger))
	r.Use(middleware.CORSMiddleware())
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(middleware.TimeoutMiddleware(30 * time.Second))
	r.Use(middleware.RateLimitMiddleware())

	r.Post("/api/v1/register", handler.Register)
	r.Post("/api/v1/login", handler.Login)
	r.Get("/api/v1/health", healthCheck)
	r.Get("/api/v1/version", versionInfo)

	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(jwtManager, logger))

		r.Get("/api/v1/entries", handler.ListEntries)
		r.Post("/api/v1/entries", handler.CreateEntry)
		r.Get("/api/v1/entries/{id}", handler.GetEntry)
		r.Put("/api/v1/entries/{id}", handler.UpdateEntry)
		r.Delete("/api/v1/entries/{id}", handler.DeleteEntry)

		r.Post("/api/v1/sync", handler.Sync)
	})

	return r
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`))
}

func versionInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"version":"1.0.0","service":"gophkeeper","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`))
}
