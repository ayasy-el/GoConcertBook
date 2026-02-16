package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppMode         string
	HTTPAddr        string
	JWTSecret       string
	RateLimitPerMin int
	ReservationTTL  time.Duration
	QueueThreshold  int
	WorkerPoolSize  int
	PostgresDSN     string
	RedisAddr       string
	RedisPassword   string
	KafkaBrokers    []string
	KafkaGroupID    string
}

func Load() Config {
	return Config{
		AppMode:         envOrDefault("APP_MODE", "memory"),
		HTTPAddr:        envOrDefault("HTTP_ADDR", ":8080"),
		JWTSecret:       envOrDefault("JWT_SECRET", "dev-secret"),
		RateLimitPerMin: envOrDefaultInt("RATE_LIMIT_PER_MIN", 120),
		ReservationTTL:  envOrDefaultDuration("RESERVATION_TTL", 5*time.Minute),
		QueueThreshold:  envOrDefaultInt("QUEUE_THRESHOLD", 1000),
		WorkerPoolSize:  envOrDefaultInt("WORKER_POOL_SIZE", 50),
		PostgresDSN:     envOrDefault("POSTGRES_DSN", "postgres://postgres:postgres@localhost:5432/concert?sslmode=disable"),
		RedisAddr:       envOrDefault("REDIS_ADDR", "localhost:6379"),
		RedisPassword:   envOrDefault("REDIS_PASSWORD", ""),
		KafkaBrokers:    envCSV("KAFKA_BROKERS", "localhost:9092"),
		KafkaGroupID:    envOrDefault("KAFKA_GROUP_ID", "concert-worker"),
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

func envCSV(key, fallback string) []string {
	v := envOrDefault(key, fallback)
	parts := make([]string, 0, 4)
	current := ""
	for _, ch := range v {
		if ch == ',' {
			if current != "" {
				parts = append(parts, current)
			}
			current = ""
			continue
		}
		if ch != ' ' && ch != '\t' && ch != '\n' {
			current += string(ch)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	if len(parts) == 0 {
		return []string{fallback}
	}
	return parts
}

func envOrDefaultDuration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}
