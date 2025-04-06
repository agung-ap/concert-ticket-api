package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"concert-ticket-api/internal/model"
	"concert-ticket-api/internal/repository"
	pkgErr "concert-ticket-api/pkg/errors"

	"github.com/jmoiron/sqlx"
)

type concertRepository struct {
	db *sqlx.DB
}

func (r *concertRepository) GetDB() *sqlx.DB {
	return r.db
}

// NewConcertRepository creates a new PostgreSQL implementation of ConcertRepository
func NewConcertRepository(db *sqlx.DB) repository.ConcertRepository {
	return &concertRepository{
		db: db,
	}
}

// GetByID retrieves a concert by its ID
func (r *concertRepository) GetByID(ctx context.Context, id int64) (*model.Concert, error) {
	query := `SELECT * FROM concerts WHERE id = $1`

	var concert model.Concert
	err := r.db.GetContext(ctx, &concert, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, pkgErr.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get concert: %w", err)
	}

	return &concert, nil
}

// List retrieves concerts with optional filtering
func (r *concertRepository) List(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]*model.Concert, error) {
	where, args := buildWhereClause(filters)

	query := fmt.Sprintf(`
		SELECT * FROM concerts
		%s
		ORDER BY concert_date
		LIMIT $%d OFFSET $%d
	`, where, len(args)+1, len(args)+2)

	args = append(args, limit, offset)

	var concerts []*model.Concert
	err := r.db.SelectContext(ctx, &concerts, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list concerts: %w", err)
	}

	return concerts, nil
}

// Count returns the total number of concerts matching the filters
func (r *concertRepository) Count(ctx context.Context, filters map[string]interface{}) (int, error) {
	where, args := buildWhereClause(filters)

	query := fmt.Sprintf(`
		SELECT COUNT(*) FROM concerts
		%s
	`, where)

	var count int
	err := r.db.GetContext(ctx, &count, query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to count concerts: %w", err)
	}

	return count, nil
}

// Create inserts a new concert
func (r *concertRepository) Create(ctx context.Context, concert *model.Concert) (*model.Concert, error) {
	query := `
		INSERT INTO concerts (
			name, artist, venue, concert_date, total_tickets, available_tickets,
			price, booking_start_time, booking_end_time
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		) RETURNING *
	`

	err := r.db.GetContext(ctx, concert, query,
		concert.Name, concert.Artist, concert.Venue, concert.ConcertDate,
		concert.TotalTickets, concert.AvailableTickets, concert.Price,
		concert.BookingStartTime, concert.BookingEndTime,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create concert: %w", err)
	}

	return concert, nil
}

// Update updates an existing concert
func (r *concertRepository) Update(ctx context.Context, concert *model.Concert) error {
	query := `
		UPDATE concerts
		SET name = $1, artist = $2, venue = $3, concert_date = $4,
			total_tickets = $5, available_tickets = $6, price = $7,
			booking_start_time = $8, booking_end_time = $9,
			version = version + 1, updated_at = NOW()
		WHERE id = $10 AND version = $11
	`

	result, err := r.db.ExecContext(ctx, query,
		concert.Name, concert.Artist, concert.Venue, concert.ConcertDate,
		concert.TotalTickets, concert.AvailableTickets, concert.Price,
		concert.BookingStartTime, concert.BookingEndTime,
		concert.ID, concert.Version,
	)
	if err != nil {
		return fmt.Errorf("failed to update concert: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return pkgErr.ErrOptimisticLockFailed
	}

	// Increment version for the caller
	concert.Version++

	return nil
}

// GetForUpdate retrieves a concert for update with row locking
func (r *concertRepository) GetForUpdate(ctx context.Context, id int64) (*model.Concert, error) {
	query := `SELECT * FROM concerts WHERE id = $1 FOR UPDATE`

	var concert model.Concert
	err := r.db.GetContext(ctx, &concert, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, pkgErr.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get concert for update: %w", err)
	}

	return &concert, nil
}

// UpdateTicketCount atomically updates the available ticket count using optimistic locking
func (r *concertRepository) UpdateTicketCount(ctx context.Context, id int64, version int, ticketCount int) error {
	query := `
		UPDATE concerts
		SET available_tickets = available_tickets - $1,
			version = version + 1,
			updated_at = NOW()
		WHERE id = $2 AND version = $3 AND available_tickets >= $1
	`

	result, err := r.db.ExecContext(ctx, query, ticketCount, id, version)
	if err != nil {
		return fmt.Errorf("failed to update ticket count: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		// Get the current available tickets to determine the error
		concert, err := r.GetByID(ctx, id)
		if err != nil {
			return fmt.Errorf("failed to get concert after update: %w", err)
		}

		if concert.Version != version {
			return pkgErr.ErrOptimisticLockFailed
		}

		if concert.AvailableTickets < ticketCount {
			return pkgErr.ErrInsufficientTickets
		}

		return pkgErr.ErrUpdateFailed
	}

	return nil
}

// Helper function to build WHERE clause from filters
func buildWhereClause(filters map[string]interface{}) (string, []interface{}) {
	if len(filters) == 0 {
		return "", nil
	}

	var conditions []string
	var args []interface{}
	i := 1

	for key, value := range filters {
		switch key {
		case "artist":
			conditions = append(conditions, fmt.Sprintf("artist ILIKE $%d", i))
			args = append(args, fmt.Sprintf("%%%v%%", value))
		case "venue":
			conditions = append(conditions, fmt.Sprintf("venue ILIKE $%d", i))
			args = append(args, fmt.Sprintf("%%%v%%", value))
		case "date_from":
			conditions = append(conditions, fmt.Sprintf("concert_date >= $%d", i))
			args = append(args, value)
		case "date_to":
			conditions = append(conditions, fmt.Sprintf("concert_date <= $%d", i))
			args = append(args, value)
		case "name":
			conditions = append(conditions, fmt.Sprintf("name ILIKE $%d", i))
			args = append(args, fmt.Sprintf("%%%v%%", value))
		case "available":
			conditions = append(conditions, fmt.Sprintf("available_tickets > 0"))
		}
		i++
	}

	return "WHERE " + strings.Join(conditions, " AND "), args
}
