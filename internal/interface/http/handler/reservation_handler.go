package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"concert-booking/internal/domain/service"
	"concert-booking/internal/interface/http/dto"
	"concert-booking/internal/usecase"
)

type ReservationHandler struct {
	usecase *usecase.ReservationUsecase
}

func NewReservationHandler(usecase *usecase.ReservationUsecase) *ReservationHandler {
	return &ReservationHandler{usecase: usecase}
}

func (h *ReservationHandler) Reserve(w http.ResponseWriter, r *http.Request) {
	var req dto.ReserveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	userID := strings.TrimSpace(r.Header.Get("X-User-ID"))
	reservation, err := h.usecase.Reserve(r.Context(), userID, req.EventID, req.Category, req.Qty)
	if err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, usecase.ErrInvalidInput):
			status = http.StatusBadRequest
		case errors.Is(err, usecase.ErrQueueFull):
			status = http.StatusTooManyRequests
		case errors.Is(err, usecase.ErrNotFound):
			status = http.StatusNotFound
		case errors.Is(err, service.ErrOutOfStock):
			status = http.StatusConflict
		}
		http.Error(w, err.Error(), status)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(reservation)
}

func (h *ReservationHandler) Confirm(w http.ResponseWriter, r *http.Request) {
	var req dto.ConfirmRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	booking, err := h.usecase.Confirm(r.Context(), req.ReservationID, req.PaymentOK)
	if err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, usecase.ErrInvalidInput):
			status = http.StatusBadRequest
		case errors.Is(err, usecase.ErrNotFound):
			status = http.StatusNotFound
		case err.Error() == "payment failed":
			status = http.StatusPaymentRequired
		}
		http.Error(w, err.Error(), status)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(booking)
}

func (h *ReservationHandler) StartExpiryReaper(stop <-chan struct{}, interval time.Duration, batch int) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			_ = h.usecase.ReleaseExpired(context.Background(), time.Now(), batch)
		}
	}
}
