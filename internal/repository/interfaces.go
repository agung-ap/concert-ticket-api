package repository

import (
	"context"

	"concert-ticket-api/internal/model"

	"github.com/jmoiron/sqlx"
)

// ConcertRepository defines the interface for concert data access
type ConcertRepository interface {
	GetDB() *sqlx.DB

	// GetByID retrieves a concert by its ID
	GetByID(ctx context.Context, id int64) (*model.Concert, error)

	// List retrieves concerts with optional filtering
	List(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]*model.Concert, error)

	// Count returns the total number of concerts matching the filters
	Count(ctx context.Context, filters map[string]interface{}) (int, error)

	// Create inserts a new concert
	Create(ctx context.Context, concert *model.Concert) (*model.Concert, error)

	// Update updates an existing concert
	Update(ctx context.Context, concert *model.Concert) error

	// GetForUpdate retrieves a concert for update with row locking
	GetForUpdate(ctx context.Context, id int64) (*model.Concert, error)

	// UpdateTicketCount atomically updates the available ticket count using optimistic locking
	UpdateTicketCount(ctx context.Context, id int64, version int, ticketCount int) error
}

// BookingRepository defines the interface for booking data access
type BookingRepository interface {
	GetDB() *sqlx.DB

	// GetByID retrieves a booking by its ID
	GetByID(ctx context.Context, id int64) (*model.Booking, error)

	// GetByUserID retrieves bookings for a user
	GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*model.Booking, error)

	// Create inserts a new booking
	Create(ctx context.Context, booking *model.Booking) (*model.Booking, error)

	// Update updates an existing booking
	Update(ctx context.Context, booking *model.Booking) error

	// CountByUserAndConcert counts bookings by a user for a specific concert
	CountByUserAndConcert(ctx context.Context, userID string, concertID int64) (int, error)

	// CreateWithTicketUpdate creates a booking and updates ticket count in a transaction
	CreateWithTicketUpdate(ctx context.Context, booking *model.Booking, concertVersion int) error
}
