package router

import (
	"net/http"

	"concert-booking/internal/interface/http/handler"
	"concert-booking/internal/interface/http/middleware"
)

type Dependencies struct {
	HealthHandler *handler.HealthHandler
	EventHandler  *handler.EventHandler
	Auth          *middleware.AuthMiddleware
	RateLimiter   *middleware.RateLimiter
}

func New(dep Dependencies) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", dep.HealthHandler.Handle)
	mux.Handle("POST /events", dep.RateLimiter.Limit(dep.Auth.RequireRole("admin", http.HandlerFunc(dep.EventHandler.CreateEvent))))
	mux.Handle("POST /events/{id}/ticket-category", dep.RateLimiter.Limit(dep.Auth.RequireRole("admin", http.HandlerFunc(dep.EventHandler.CreateCategory))))
	mux.Handle("GET /events/{id}/availability", dep.RateLimiter.Limit(http.HandlerFunc(dep.EventHandler.Availability)))

	return mux
}
