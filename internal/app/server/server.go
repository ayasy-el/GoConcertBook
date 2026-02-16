package server

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"concert-booking/internal/app/config"
	"concert-booking/internal/infrastructure/memory"
	"concert-booking/internal/interface/http/handler"
	"concert-booking/internal/interface/http/middleware"
	"concert-booking/internal/interface/http/router"
	"concert-booking/internal/usecase"
)

func NewHTTPServer(cfg config.Config) *http.Server {
	eventRepo := memory.NewEventRepository()
	categoryRepo := memory.NewTicketCategoryRepository()
	eventUsecase := usecase.NewEventUsecase(eventRepo, categoryRepo, time.Now, newID)

	h := router.New(router.Dependencies{
		HealthHandler: handler.NewHealthHandler(),
		EventHandler:  handler.NewEventHandler(eventUsecase),
		Auth:          middleware.NewAuthMiddleware(cfg.JWTSecret),
		RateLimiter:   middleware.NewRateLimiter(cfg.RateLimitPerMin, time.Minute),
	})

	return &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           h,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}

func newID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return time.Now().Format("20060102150405.000000000")
	}
	return hex.EncodeToString(b)
}
