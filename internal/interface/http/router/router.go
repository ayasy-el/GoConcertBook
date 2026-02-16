package router

import (
	"net/http"

	"concert-booking/internal/interface/http/handler"
	"concert-booking/internal/interface/http/middleware"
	"concert-booking/internal/observability/metrics"
)

type Dependencies struct {
	HealthHandler      *handler.HealthHandler
	EventHandler       *handler.EventHandler
	ReservationHandler *handler.ReservationHandler
	Auth               *middleware.AuthMiddleware
	RateLimiter        *middleware.RateLimiter
}

func New(dep Dependencies) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", dep.HealthHandler.Handle)
	mux.HandleFunc("GET /metrics", metrics.Handler)
	mux.Handle("POST /events", dep.RateLimiter.Limit(dep.Auth.RequireRole("admin", http.HandlerFunc(dep.EventHandler.CreateEvent))))
	mux.Handle("POST /events/{id}/ticket-category", dep.RateLimiter.Limit(dep.Auth.RequireRole("admin", http.HandlerFunc(dep.EventHandler.CreateCategory))))
	mux.Handle("GET /events/{id}/availability", dep.RateLimiter.Limit(http.HandlerFunc(dep.EventHandler.Availability)))
	mux.Handle("POST /reserve", dep.RateLimiter.Limit(dep.Auth.RequireRole("user", http.HandlerFunc(dep.ReservationHandler.Reserve))))
	mux.Handle("POST /confirm", dep.RateLimiter.Limit(dep.Auth.RequireRole("user", http.HandlerFunc(dep.ReservationHandler.Confirm))))

	return middleware.Instrument(mux)
}
