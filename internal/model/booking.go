package model

import (
	"time"
)

// BookingStatus represents the status of a booking
type BookingStatus string

const (
	BookingStatusConfirmed BookingStatus = "confirmed"
	BookingStatusCancelled BookingStatus = "cancelled"
	BookingStatusPending   BookingStatus = "pending"
)

// Booking represents a ticket booking for a concert
type Booking struct {
	ID          int64         `json:"id" db:"id"`
	ConcertID   int64         `json:"concert_id" db:"concert_id"`
	UserID      string        `json:"user_id" db:"user_id"`
	TicketCount int           `json:"ticket_count" db:"ticket_count"`
	BookingTime time.Time     `json:"booking_time" db:"booking_time"`
	Status      BookingStatus `json:"status" db:"status"`
	CreatedAt   time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at" db:"updated_at"`
}

// BookingRequest represents a request to book tickets
type BookingRequest struct {
	ConcertID   int64  `json:"concert_id" validate:"required"`
	UserID      string `json:"user_id" validate:"required"`
	TicketCount int    `json:"ticket_count" validate:"required,min=1"`
}
