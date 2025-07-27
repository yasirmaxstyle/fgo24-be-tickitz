package dto

import "time"

type CreateTransactionRequest struct {
	ShowtimeID        int      `json:"showtime_id" binding:"required"`
	SeatNumbers       []string `json:"seat_numbers" binding:"required"`
	RecipientEmail    string   `json:"recipient_email" binding:"required,email"`
	RecipientFullName string   `json:"recipient_full_name" binding:"required"`
	RecipientPhone    string   `json:"recipient_phone_number" binding:"required"`
	PaymentMethodID   int      `json:"payment_method_id" binding:"required"`
}

type ProcessPaymentRequest struct {
	TransactionCode string `json:"transaction_code" binding:"required"`
	PaymentProof    string `json:"payment_proof,omitempty"`
}

type TransactionResult struct {
	Transaction TransactionResponse `json:"transaction"`
	Tickets     []TicketResponse    `json:"tickets"`
}

type TransactionResponse struct {
	TransactionID     int        `json:"transaction_id"`
	TransactionCode   string     `json:"transaction_code"`
	RecipientEmail    string     `json:"recipient_email"`
	RecipientFullName string     `json:"recipient_full_name"`
	RecipientPhone    string     `json:"recipient_phone_number"`
	TotalSeats        int        `json:"total_seats"`
	TotalAmount       float64    `json:"total_amount"`
	Status            string     `json:"status"`
	CreatedAt         time.Time  `json:"created_at"`
	ExpiresAt         time.Time  `json:"expires_at"`
	PaidAt            *time.Time `json:"paid_at"`
	CreatedBy         int        `json:"created_by"`
	PaymentMethodID   int        `json:"payment_method_id"`
}

type TicketResponse struct {
	TicketID      int       `json:"ticket_id"`
	TicketCode    string    `json:"ticket_code"`
	ShowtimeID    int       `json:"showtime_id"`
	SeatNumber    string    `json:"seat_number"`
	Status        string    `json:"status"`
	TransactionID int       `json:"transaction_id"`
	CreatedAt     time.Time `json:"created_at"`
}

type ShowtimeResponse struct {
	ShowtimeID   int       `json:"showtime_id"`
	ShowDatetime time.Time `json:"show_datetime"`
	Price        float64   `json:"price"`
}

type CinemaResponse struct {
	CinemaID int    `json:"cinema_id"`
	Name     string `json:"name"`
	Location string `json:"location"`
}

type TransactionListResponse struct {
	TransactionID   int              `json:"transaction_id"`
	TransactionCode string           `json:"transaction_code"`
	Status          string           `json:"status"`
	TotalAmount     float64          `json:"total_amount"`
	ExpiresAt       time.Time        `json:"expires_at"`
	CreatedAt       time.Time        `json:"created_at"`
	Movie           MovieResponse    `json:"movie"`
	Showtime        ShowtimeResponse `json:"showtime"`
	Cinema          CinemaResponse   `json:"cinema"`
	Seats           []string         `json:"seats"`
}
