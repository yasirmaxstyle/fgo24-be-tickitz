package models

import (
	"time"
)

type Screen struct {
	ScreenID   int       `json:"screen_id" db:"screen_id"`
	CinemaID   int       `json:"cinema_id" db:"cinema_id"`
	Name       string    `json:"name" db:"name"`
	TotalSeats int       `json:"total_seats" db:"total_seats"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}
