package server

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"time"

	"concert-booking/internal/app/config"
	kafkainfra "concert-booking/internal/infrastructure/kafka"
	"concert-booking/internal/infrastructure/memory"
	"concert-booking/internal/infrastructure/postgres"
	redisinfra "concert-booking/internal/infrastructure/redis"
	"concert-booking/internal/interface/http/handler"
	"concert-booking/internal/interface/http/middleware"
	"concert-booking/internal/interface/http/router"
	"concert-booking/internal/observability/metrics"
	"concert-booking/internal/usecase"
)

func NewHTTPServer(cfg config.Config) *http.Server {
	var (
		eventUsecase       *usecase.EventUsecase
		reservationUsecase *usecase.ReservationUsecase
		cleanup            []func()
	)

	if cfg.AppMode == "production" {
		db, err := postgres.NewDB(cfg.PostgresDSN)
		if err != nil {
			log.Fatalf("postgres connect failed: %v", err)
		}
		cleanup = append(cleanup, func() { _ = db.Close() })

		stock := redisinfra.NewStockService(cfg.RedisAddr, cfg.RedisPassword)
		if err := stock.Ping(context.Background()); err != nil {
			log.Fatalf("redis connect failed: %v", err)
		}

		producer := kafkainfra.NewProducer(cfg.KafkaBrokers)
		cleanup = append(cleanup, func() { _ = producer.Close() })

		eventRepo := postgres.NewEventRepository(db)
		categoryRepo := postgres.NewTicketCategoryRepository(db)
		reservationRepo := postgres.NewReservationRepository(db)
		bookingRepo := postgres.NewBookingRepository(db)

		eventUsecase = usecase.NewEventUsecase(eventRepo, categoryRepo, stock, time.Now, newID)
		reservationUsecase = usecase.NewReservationUsecase(categoryRepo, reservationRepo, bookingRepo, stock, producer, time.Now, newID, cfg.ReservationTTL, cfg.QueueThreshold, cfg.WorkerPoolSize, false)

		collectorStop := make(chan struct{})
		go metrics.StartInfraCollectors(db, stock.Client(), 5*time.Second, collectorStop)
		cleanup = append(cleanup, func() { close(collectorStop) })
	} else {
		eventRepo := memory.NewEventRepository()
		categoryRepo := memory.NewTicketCategoryRepository()
		reservationRepo := memory.NewReservationRepository()
		bookingRepo := memory.NewBookingRepository()
		stock := memory.NewStockService()
		producer := memory.NewEventProducer()

		eventUsecase = usecase.NewEventUsecase(eventRepo, categoryRepo, stock, time.Now, newID)
		reservationUsecase = usecase.NewReservationUsecase(categoryRepo, reservationRepo, bookingRepo, stock, producer, time.Now, newID, cfg.ReservationTTL, cfg.QueueThreshold, cfg.WorkerPoolSize, true)
	}

	h := router.New(router.Dependencies{
		HealthHandler:      handler.NewHealthHandler(),
		EventHandler:       handler.NewEventHandler(eventUsecase),
		ReservationHandler: handler.NewReservationHandler(reservationUsecase),
		Auth:               middleware.NewAuthMiddleware(cfg.JWTSecret),
		RateLimiter:        middleware.NewRateLimiter(cfg.RateLimitPerMin, time.Minute),
	})

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           h,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	reaperCtx, cancel := context.WithCancel(context.Background())
	go reservationUsecase.StartExpiryReaper(reaperCtx, 2*time.Second, 100)

	srv.RegisterOnShutdown(func() {
		cancel()
		for _, fn := range cleanup {
			fn()
		}
	})

	return srv
}

func newID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return time.Now().Format("20060102150405.000000000")
	}
	return hex.EncodeToString(b)
}
