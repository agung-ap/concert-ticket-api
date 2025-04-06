package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"concert-ticket-api/internal/model"
	"concert-ticket-api/internal/repository"
	pkgErr "concert-ticket-api/pkg/errors"

	"github.com/jmoiron/sqlx"
)

type bookingRepository struct {
	db *sqlx.DB
}

func (r *bookingRepository) GetDB() *sqlx.DB {
	return r.db
}

// NewBookingRepository creates a new PostgreSQL implementation of BookingRepository
func NewBookingRepository(db *sqlx.DB) repository.BookingRepository {
	return &bookingRepository{
		db: db,
	}
}

// GetByID retrieves a booking by its ID
func (r *bookingRepository) GetByID(ctx context.Context, id int64) (*model.Booking, error) {
	query := `SELECT b.id, b.concert_id, b.user_id, b.ticket_count, b.booking_time, b.status, b.created_at, b.updated_at
		FROM bookings b WHERE b.id = $1`

	var booking model.Booking
	err := r.db.GetContext(ctx, &booking, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, pkgErr.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get booking: %w", err)
	}

	return &booking, nil
}

// GetByUserID retrieves bookings for a user
func (r *bookingRepository) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*model.Booking, error) {
	query := `
		SELECT b.id, b.concert_id, b.user_id, b.ticket_count, b.booking_time, b.status, b.created_at, b.updated_at
		FROM bookings b
		JOIN concerts c ON b.concert_id = c.id
		WHERE b.user_id = $1
		ORDER BY b.booking_time DESC
		LIMIT $2 OFFSET $3
	`

	var bookings []*model.Booking
	err := r.db.SelectContext(ctx, &bookings, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get user bookings: %w", err)
	}

	return bookings, nil
}

// Create inserts a new booking
func (r *bookingRepository) Create(ctx context.Context, booking *model.Booking) (*model.Booking, error) {
	query := `
		INSERT INTO bookings (
			concert_id, user_id, ticket_count, status
		) VALUES (
			$1, $2, $3, $4
		) RETURNING *
	`

	err := r.db.GetContext(ctx, booking, query,
		booking.ConcertID, booking.UserID, booking.TicketCount, booking.Status,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create booking: %w", err)
	}

	return booking, nil
}

// Update updates an existing booking
func (r *bookingRepository) Update(ctx context.Context, booking *model.Booking) error {
	query := `
		UPDATE bookings
		SET concert_id = $1, user_id = $2, ticket_count = $3, status = $4, updated_at = NOW()
		WHERE id = $5
	`

	_, err := r.db.ExecContext(ctx, query,
		booking.ConcertID, booking.UserID, booking.TicketCount, booking.Status, booking.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update booking: %w", err)
	}

	return nil
}

// CountByUserAndConcert counts bookings by a user for a specific concert
func (r *bookingRepository) CountByUserAndConcert(ctx context.Context, userID string, concertID int64) (int, error) {
	query := `
		SELECT COUNT(id) 
		FROM bookings 
		WHERE user_id = $1 AND concert_id = $2 AND status = 'confirmed'
	`

	var count int
	err := r.db.GetContext(ctx, &count, query, userID, concertID)
	if err != nil {
		return 0, fmt.Errorf("failed to count user bookings for concert: %w", err)
	}

	return count, nil
}

// CreateWithTicketUpdate creates a booking and updates ticket count in a transaction
func (r *bookingRepository) CreateWithTicketUpdate(ctx context.Context, booking *model.Booking, concertVersion int) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Defer a rollback in case anything fails
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			// Log the rollback error, but don't return it
			// We're more interested in the original error
		}
	}()

	// First, get the concert with row lock
	var concert model.Concert
	getConcertQuery := `SELECT id, name, artist, venue, concert_date, total_tickets, available_tickets, price, 
    	booking_start_time, booking_end_time, version, created_at, updated_at 
		FROM concerts WHERE id = $1 FOR UPDATE`
	err = tx.GetContext(ctx, &concert, getConcertQuery, booking.ConcertID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return pkgErr.ErrNotFound
		}
		return fmt.Errorf("failed to get concert for booking: %w", err)
	}

	// Check if the concert version matches
	if concert.Version != concertVersion {
		return pkgErr.ErrOptimisticLockFailed
	}

	// Check if concert is open for booking
	if !concert.IsBookingOpen() {
		return pkgErr.ErrBookingClosed
	}

	// Check if there are enough tickets
	if concert.AvailableTickets < booking.TicketCount {
		return pkgErr.ErrInsufficientTickets
	}

	// Update the ticket count
	updateTicketQuery := `
		UPDATE concerts
		SET available_tickets = available_tickets - $1,
			version = version + 1,
			updated_at = NOW()
		WHERE id = $2
	`
	_, err = tx.ExecContext(ctx, updateTicketQuery, booking.TicketCount, booking.ConcertID)
	if err != nil {
		return fmt.Errorf("failed to update ticket count: %w", err)
	}

	// Create the booking
	createBookingQuery := `
		INSERT INTO bookings (
			concert_id, user_id, ticket_count, status
		) VALUES (
			$1, $2, $3, $4
		) RETURNING id, booking_time, created_at, updated_at
	`

	err = tx.GetContext(ctx, booking, createBookingQuery,
		booking.ConcertID, booking.UserID, booking.TicketCount, booking.Status,
	)
	if err != nil {
		return fmt.Errorf("failed to create booking: %w", err)
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
