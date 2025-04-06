package model

import (
	"time"
)

// Concert represents a concert event with ticket information
type Concert struct {
	ID               int64     `json:"id" db:"id"`
	Name             string    `json:"name" db:"name"`
	Artist           string    `json:"artist" db:"artist"`
	Venue            string    `json:"venue" db:"venue"`
	ConcertDate      time.Time `json:"concert_date" db:"concert_date"`
	TotalTickets     int       `json:"total_tickets" db:"total_tickets"`
	AvailableTickets int       `json:"available_tickets" db:"available_tickets"`
	Price            float64   `json:"price" db:"price"`
	BookingStartTime time.Time `json:"booking_start_time" db:"booking_start_time"`
	BookingEndTime   time.Time `json:"booking_end_time" db:"booking_end_time"`
	Version          int       `json:"version" db:"version"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

// IsBookingOpen checks if booking is currently open for this concert
func (c *Concert) IsBookingOpen() bool {
	now := time.Now()
	return now.After(c.BookingStartTime) && now.Before(c.BookingEndTime)
}

// HasAvailableTickets checks if the concert has enough available tickets
func (c *Concert) HasAvailableTickets(count int) bool {
	return c.AvailableTickets >= count
}
