package models

import (
	"time"
)

type Transaction struct {
	TransactionID     int        `json:"transaction_id" db:"transaction_id"`
	TransactionCode   string     `json:"transaction_code" db:"transaction_code"`
	RecipientEmail    string     `json:"recipient_email" db:"recipient_email"`
	RecipientFullName string     `json:"recipient_full_name" db:"recipient_full_name"`
	RecipientPhone    string     `json:"recipient_phone_number" db:"recipient_phone_number"`
	TotalSeats        int        `json:"total_seats" db:"total_seats"`
	TotalAmount       float64    `json:"total_amount" db:"total_amount"`
	Status            string     `json:"status" db:"status"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	ExpiresAt         time.Time  `json:"expires_at" db:"expires_at"`
	PaidAt            *time.Time `json:"paid_at" db:"paid_at"`
	CreatedBy         int        `json:"created_by" db:"created_by"`
	PaymentMethodID   int        `json:"payment_method_id" db:"payment_method_id"`
}

type Ticket struct {
	TicketID      int       `json:"ticket_id" db:"ticket_id"`
	TicketCode    string    `json:"ticket_code" db:"ticket_code"`
	ShowtimeID    int       `json:"showtime_id" db:"showtime_id"`
	SeatNumber    string    `json:"seat_number" db:"seat_number"`
	Price         float64   `json:"price" db:"price"`
	Status        string    `json:"status" db:"status"`
	TransactionID int       `json:"transaction_id" db:"transaction_id"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

type Showtime struct {
	ShowtimeID   int       `json:"showtime_id" db:"showtime_id"`
	MovieID      int       `json:"movie_id" db:"movie_id"`
	ScreenID     int       `json:"screen_id" db:"screen_id"`
	ShowDatetime time.Time `json:"show_datetime" db:"show_datetime"`
	BasePrice    float64   `json:"base_price" db:"base_price"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

type PaymentMethod struct {
	PaymentMethodID int    `json:"payment_method_id" db:"payment_method_id"`
	Name            string `json:"name" db:"name"`
	Code            string `json:"code" db:"code"`
	IsActive        bool   `json:"is_active" db:"is_active"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

type Cinema struct {
	CinemaID    int       `json:"cinema_id" db:"cinema_id"`
	Name        string    `json:"name" db:"name"`
	ImagePath   string    `json:"image_path" db:"image_path"`
	Location    string    `json:"location" db:"location"`
	TotalSeats  int       `json:"total_seats" db:"total_seats"`
	Address     string    `json:"address" db:"address"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type TransactionJoinRow struct {
	TransactionID   int        `db:"transaction_id"`
	TransactionCode string     `db:"transaction_code"`
	Status          string     `db:"status"`
	TotalAmount     float64    `db:"total_amount"`
	ExpiresAt       time.Time  `db:"expires_at"`
	CreatedAt       time.Time  `db:"created_at"`
	SeatNumber      *string    `db:"seat_number"`
	ShowtimeID      *int       `db:"showtime_id"`
	ShowDatetime    *time.Time `db:"show_datetime"`
	TicketPrice     *float64   `db:"ticket_price"`
	MovieID         *int       `db:"movie_id"`
	MovieTitle      *string    `db:"movie_title"`
	CinemaID        *int       `db:"cinema_id"`
	CinemaName      *string    `db:"cinema_name"`
	CinemaLocation  *string    `db:"cinema_location"`
}
