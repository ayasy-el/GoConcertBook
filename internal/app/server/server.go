package server

import (
	"net/http"
	"time"

	"concert-booking/internal/app/config"
	"concert-booking/internal/interface/http/router"
)

func NewHTTPServer(cfg config.Config) *http.Server {
	h := router.New()

	return &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           h,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}
