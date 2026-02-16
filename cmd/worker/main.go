package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"concert-booking/internal/app/config"
	"concert-booking/internal/domain/entity"
	kafkainfra "concert-booking/internal/infrastructure/kafka"
	"concert-booking/internal/infrastructure/postgres"
	"concert-booking/internal/observability/metrics"
)

func main() {
	cfg := config.Load()
	db, err := connectPostgresWithRetry(cfg.PostgresDSN, 10, 2*time.Second)
	if err != nil {
		log.Fatalf("postgres connect failed: %v", err)
	}
	defer db.Close()
	if err := waitForDependency(10, 2*time.Second, func() error {
		return kafkainfra.EnsureTopics(context.Background(), cfg.KafkaBrokers, []string{"ticket.reserved", "ticket.confirmed", "ticket.expired"}, 3, 1)
	}); err != nil {
		log.Fatalf("kafka topic ensure failed: %v", err)
	}

	reservationRepo := postgres.NewReservationRepository(db)
	consumer := kafkainfra.NewConsumer(cfg.KafkaBrokers, cfg.KafkaGroupID, "ticket.reserved")
	defer consumer.Close()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	metricsStop := make(chan struct{})
	go func() {
		t := time.NewTicker(5 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-metricsStop:
				return
			case <-t.C:
				stats := consumer.Stats()
				metrics.SetKafkaLag(stats.Lag)
			}
		}
	}()

	httpSrv := &http.Server{Addr: ":9091", Handler: http.HandlerFunc(metrics.Handler)}
	go func() {
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("worker metrics server error: %v", err)
		}
	}()

	for {
		msg, err := consumer.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				close(metricsStop)
				_ = httpSrv.Shutdown(context.Background())
				log.Println("worker shutting down")
				return
			}
			log.Printf("consume error: %v", err)
			continue
		}
		var res entity.Reservation
		if err := json.Unmarshal(msg.Value, &res); err != nil {
			log.Printf("invalid message: %v", err)
			continue
		}
		if err := reservationRepo.Upsert(res); err != nil {
			log.Printf("upsert reservation failed: %v", err)
			continue
		}
		log.Printf("reservation persisted: %s", res.ID)
	}
}

func connectPostgresWithRetry(dsn string, attempts int, delay time.Duration) (*sql.DB, error) {
	var lastErr error
	for i := 0; i < attempts; i++ {
		db, err := postgres.NewDB(dsn)
		if err == nil {
			return db, nil
		}
		lastErr = err
		time.Sleep(delay)
	}
	return nil, lastErr
}

func waitForDependency(attempts int, delay time.Duration, fn func() error) error {
	var lastErr error
	for i := 0; i < attempts; i++ {
		if err := fn(); err == nil {
			return nil
		} else {
			lastErr = err
		}
		time.Sleep(delay)
	}
	if lastErr == nil {
		return errors.New("dependency unavailable")
	}
	return lastErr
}
