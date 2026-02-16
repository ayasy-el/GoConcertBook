package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"concert-booking/internal/interface/http/dto"
	"concert-booking/internal/usecase"
)

type EventHandler struct {
	usecase *usecase.EventUsecase
}

func NewEventHandler(usecase *usecase.EventUsecase) *EventHandler {
	return &EventHandler{usecase: usecase}
}

// CreateEvent godoc
// @Summary Create event
// @Tags events
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateEventRequest true "Create event payload"
// @Success 201 {object} entity.Event
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /events [post]
func (h *EventHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	date, err := time.Parse(time.RFC3339, req.Date)
	if err != nil {
		http.Error(w, "invalid date format", http.StatusBadRequest)
		return
	}
	e, err := h.usecase.CreateEvent(req.Name, date)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, usecase.ErrInvalidInput) {
			status = http.StatusBadRequest
		}
		http.Error(w, err.Error(), status)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(e)
}

// CreateCategory godoc
// @Summary Create ticket category
// @Tags events
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Event ID"
// @Param request body dto.CreateCategoryRequest true "Create category payload"
// @Success 201 {object} entity.TicketCategory
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /events/{id}/ticket-category [post]
func (h *EventHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	eventID := strings.TrimSpace(r.PathValue("id"))
	var req dto.CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	c, err := h.usecase.CreateCategory(eventID, req.Name, req.TotalStock, req.Price)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, usecase.ErrInvalidInput) {
			status = http.StatusBadRequest
		}
		if errors.Is(err, usecase.ErrNotFound) {
			status = http.StatusNotFound
		}
		http.Error(w, err.Error(), status)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(c)
}

// Availability godoc
// @Summary Get realtime ticket availability
// @Tags events
// @Produce json
// @Param id path string true "Event ID"
// @Success 200 {object} map[string]int
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /events/{id}/availability [get]
func (h *EventHandler) Availability(w http.ResponseWriter, r *http.Request) {
	eventID := strings.TrimSpace(r.PathValue("id"))
	availability, err := h.usecase.Availability(eventID)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, usecase.ErrInvalidInput) {
			status = http.StatusBadRequest
		}
		if errors.Is(err, usecase.ErrNotFound) {
			status = http.StatusNotFound
		}
		http.Error(w, err.Error(), status)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(availability)
}
