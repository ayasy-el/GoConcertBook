package dto

type ReserveRequest struct {
	EventID  string `json:"event_id"`
	Category string `json:"category"`
	Qty      int    `json:"qty"`
}

type ConfirmRequest struct {
	ReservationID string `json:"reservation_id"`
	PaymentOK     bool   `json:"payment_ok"`
}
