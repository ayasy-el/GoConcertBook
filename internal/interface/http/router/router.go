package router

import (
	"net/http"

	"concert-booking/internal/interface/http/handler"
)

func New() http.Handler {
	mux := http.NewServeMux()
	health := handler.NewHealthHandler()

	mux.HandleFunc("GET /health", health.Handle)

	return mux
}
