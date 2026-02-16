package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPAddr         string
	JWTSecret        string
	RateLimitPerMin  int
	ReservationTTL   time.Duration
	QueueThreshold   int
	WorkerPoolSize   int
}

func Load() Config {
	return Config{
		HTTPAddr:        envOrDefault("HTTP_ADDR", ":8080"),
		JWTSecret:       envOrDefault("JWT_SECRET", "dev-secret"),
		RateLimitPerMin: envOrDefaultInt("RATE_LIMIT_PER_MIN", 120),
		ReservationTTL:  5 * time.Minute,
		QueueThreshold:  envOrDefaultInt("QUEUE_THRESHOLD", 1000),
		WorkerPoolSize:  envOrDefaultInt("WORKER_POOL_SIZE", 50),
	}
}

func envOrDefault(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

func envOrDefaultInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
