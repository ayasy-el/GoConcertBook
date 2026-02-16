package usecase

import (
	"testing"
	"time"

	"concert-booking/internal/infrastructure/memory"
)

func TestEventUsecaseAvailability(t *testing.T) {
	events := memory.NewEventRepository()
	categories := memory.NewTicketCategoryRepository()
	stock := memory.NewStockService()
	u := NewEventUsecase(events, categories, stock, func() time.Time { return time.Unix(1000, 0) }, func() string { return "id-1" })

	e, err := u.CreateEvent("Coldplay", time.Now())
	if err != nil {
		t.Fatalf("create event: %v", err)
	}
	if _, err := u.CreateCategory(e.ID, "VIP", 10, 100000); err != nil {
		t.Fatalf("create category: %v", err)
	}

	av, err := u.Availability(e.ID)
	if err != nil {
		t.Fatalf("availability: %v", err)
	}
	if av["vip"] != 10 {
		t.Fatalf("expected vip=10, got %d", av["vip"])
	}
}
