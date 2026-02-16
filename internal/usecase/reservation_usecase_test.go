package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"concert-booking/internal/domain/entity"
	"concert-booking/internal/domain/service"
	"concert-booking/internal/infrastructure/memory"
)

func TestReserveAndConfirm(t *testing.T) {
	categories := memory.NewTicketCategoryRepository()
	stock := memory.NewStockService()
	reservations := memory.NewReservationRepository()
	bookings := memory.NewBookingRepository()
	producer := memory.NewEventProducer()

	eventID := "event-1"
	_ = categories.Create(entity.TicketCategory{ID: "cat-1", EventID: eventID, Name: "VIP", TotalStock: 3, Price: 1000})
	_ = stock.InitStock(context.Background(), eventID, "VIP", 3)

	idSeq := 0
	u := NewReservationUsecase(categories, reservations, bookings, stock, producer, time.Now, func() string {
		idSeq++
		if idSeq == 1 {
			return "res-1"
		}
		return "book-1"
	}, 5*time.Minute, 100, 10)

	res, err := u.Reserve(context.Background(), "user-1", eventID, "vip", 2)
	if err != nil {
		t.Fatalf("reserve failed: %v", err)
	}
	if res.ID != "res-1" {
		t.Fatalf("unexpected reservation id %s", res.ID)
	}

	book, err := u.Confirm(context.Background(), res.ID, true)
	if err != nil {
		t.Fatalf("confirm failed: %v", err)
	}
	if book.ID != "book-1" {
		t.Fatalf("unexpected booking id %s", book.ID)
	}
}

func TestReserveOutOfStock(t *testing.T) {
	categories := memory.NewTicketCategoryRepository()
	stock := memory.NewStockService()
	reservations := memory.NewReservationRepository()
	bookings := memory.NewBookingRepository()
	producer := memory.NewEventProducer()

	eventID := "event-1"
	_ = categories.Create(entity.TicketCategory{ID: "cat-1", EventID: eventID, Name: "REGULAR", TotalStock: 1, Price: 1000})
	_ = stock.InitStock(context.Background(), eventID, "REGULAR", 1)

	u := NewReservationUsecase(categories, reservations, bookings, stock, producer, time.Now, func() string { return "res-1" }, 5*time.Minute, 100, 10)
	_, err := u.Reserve(context.Background(), "user-1", eventID, "REGULAR", 2)
	if !errors.Is(err, service.ErrOutOfStock) {
		t.Fatalf("expected out of stock, got %v", err)
	}
}
