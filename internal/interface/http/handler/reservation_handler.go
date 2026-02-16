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
	"concert-booking/internal/observability/metrics"
	"concert-booking/internal/usecase"
)

type ReservationHandler struct {
	usecase *usecase.ReservationUsecase
}

func NewReservationHandler(usecase *usecase.ReservationUsecase) *ReservationHandler {
	return &ReservationHandler{usecase: usecase}
}

// Reserve godoc
// @Summary Reserve ticket
// @Tags reservation
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.ReserveRequest true "Reserve payload"
// @Success 201 {object} entity.Reservation
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 429 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /reserve [post]
func (h *ReservationHandler) Reserve(w http.ResponseWriter, r *http.Request) {
	var req dto.ReserveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	userID := strings.TrimSpace(r.Header.Get("X-User-ID"))
	reservation, err := h.usecase.Reserve(r.Context(), userID, req.EventID, req.Category, req.Qty)
	if err != nil {
		metrics.IncReservationFailed()
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
	metrics.IncReservationSuccess()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(reservation)
}

// Confirm godoc
// @Summary Confirm reservation (payment simulation)
// @Tags reservation
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.ConfirmRequest true "Confirm payload"
// @Success 200 {object} entity.Booking
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 402 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /confirm [post]
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
