package mocks

import (
	"context"
	"sync"

	"concert-ticket-api/internal/model"
	"concert-ticket-api/internal/repository"
	"concert-ticket-api/pkg/errors"

	"github.com/jmoiron/sqlx"
)

// MockConcertRepository is a mock implementation of ConcertRepository
type MockConcertRepository struct {
	mutex    sync.RWMutex
	concerts map[int64]*model.Concert
	nextID   int64
}

func (r *MockConcertRepository) GetDB() *sqlx.DB {
	return r.GetDB()
}

// NewMockConcertRepository creates a new mock concert repository
func NewMockConcertRepository() *MockConcertRepository {
	return &MockConcertRepository{
		concerts: make(map[int64]*model.Concert),
		nextID:   1,
	}
}

// GetByID retrieves a concert by its ID
func (r *MockConcertRepository) GetByID(ctx context.Context, id int64) (*model.Concert, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	concert, ok := r.concerts[id]
	if !ok {
		return nil, errors.ErrNotFound
	}

	// Return a copy to avoid mutation
	concertCopy := *concert
	return &concertCopy, nil
}

// List retrieves concerts with optional filtering
func (r *MockConcertRepository) List(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]*model.Concert, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	result := make([]*model.Concert, 0, len(r.concerts))
	for _, concert := range r.concerts {
		// Apply filters here if needed
		// For now, we're ignoring filters for simplicity
		result = append(result, concert)
	}

	// Apply limit and offset
	if offset >= len(result) {
		return []*model.Concert{}, nil
	}

	end := offset + limit
	if end > len(result) {
		end = len(result)
	}

	return result[offset:end], nil
}

// Count returns the total number of concerts matching the filters
func (r *MockConcertRepository) Count(ctx context.Context, filters map[string]interface{}) (int, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	// For simplicity, we're ignoring filters
	return len(r.concerts), nil
}

// Create inserts a new concert
func (r *MockConcertRepository) Create(ctx context.Context, concert *model.Concert) (*model.Concert, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Assign ID
	concert.ID = r.nextID
	r.nextID++

	// Set initial version
	concert.Version = 1

	// Make a copy and store it
	concertCopy := *concert
	r.concerts[concert.ID] = &concertCopy

	return concert, nil
}

// Update updates an existing concert
func (r *MockConcertRepository) Update(ctx context.Context, concert *model.Concert) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	existing, ok := r.concerts[concert.ID]
	if !ok {
		return errors.ErrNotFound
	}

	if existing.Version != concert.Version {
		return errors.ErrOptimisticLockFailed
	}

	// Update version
	concert.Version++

	// Store updated concert
	concertCopy := *concert
	r.concerts[concert.ID] = &concertCopy

	return nil
}

// GetForUpdate retrieves a concert for update with row locking
func (r *MockConcertRepository) GetForUpdate(ctx context.Context, id int64) (*model.Concert, error) {
	// In this mock, we'll just call GetByID
	// In a real database, this would acquire a lock
	return r.GetByID(ctx, id)
}

// UpdateTicketCount atomically updates the available ticket count using optimistic locking
func (r *MockConcertRepository) UpdateTicketCount(ctx context.Context, id int64, version int, ticketCount int) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	concert, ok := r.concerts[id]
	if !ok {
		return errors.ErrNotFound
	}

	if concert.Version != version {
		return errors.ErrOptimisticLockFailed
	}

	if concert.AvailableTickets < ticketCount {
		return errors.ErrInsufficientTickets
	}

	// Update ticket count and version
	concert.AvailableTickets -= ticketCount
	concert.Version++

	return nil
}

// MockBookingRepository is a mock implementation of BookingRepository
type MockBookingRepository struct {
	mutex    sync.RWMutex
	bookings map[int64]*model.Booking
	nextID   int64
}

func (r *MockBookingRepository) GetDB() *sqlx.DB {
	return r.GetDB()
}

// NewMockBookingRepository creates a new mock booking repository
func NewMockBookingRepository() *MockBookingRepository {
	return &MockBookingRepository{
		bookings: make(map[int64]*model.Booking),
		nextID:   1,
	}
}

// GetByID retrieves a booking by its ID
func (r *MockBookingRepository) GetByID(ctx context.Context, id int64) (*model.Booking, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	booking, ok := r.bookings[id]
	if !ok {
		return nil, errors.ErrNotFound
	}

	// Return a copy
	bookingCopy := *booking
	return &bookingCopy, nil
}

// GetByUserID retrieves bookings for a user
func (r *MockBookingRepository) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*model.Booking, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var result []*model.Booking
	for _, booking := range r.bookings {
		if booking.UserID == userID {
			result = append(result, booking)
		}
	}

	// Apply limit and offset
	if offset >= len(result) {
		return []*model.Booking{}, nil
	}

	end := offset + limit
	if end > len(result) {
		end = len(result)
	}

	return result[offset:end], nil
}

// Create inserts a new booking
func (r *MockBookingRepository) Create(ctx context.Context, booking *model.Booking) (*model.Booking, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Assign ID
	booking.ID = r.nextID
	r.nextID++

	// Make a copy and store it
	bookingCopy := *booking
	r.bookings[booking.ID] = &bookingCopy

	return booking, nil
}

// Update updates an existing booking
func (r *MockBookingRepository) Update(ctx context.Context, booking *model.Booking) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	_, ok := r.bookings[booking.ID]
	if !ok {
		return errors.ErrNotFound
	}

	// Store updated booking
	bookingCopy := *booking
	r.bookings[booking.ID] = &bookingCopy

	return nil
}

// CountByUserAndConcert counts bookings by a user for a specific concert
func (r *MockBookingRepository) CountByUserAndConcert(ctx context.Context, userID string, concertID int64) (int, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	count := 0
	for _, booking := range r.bookings {
		if booking.UserID == userID && booking.ConcertID == concertID && booking.Status == model.BookingStatusConfirmed {
			count++
		}
	}

	return count, nil
}

// CreateWithTicketUpdate creates a booking and updates ticket count in a transaction
func (r *MockBookingRepository) CreateWithTicketUpdate(ctx context.Context, booking *model.Booking, concertVersion int) error {
	// This would normally be implemented with a proper transaction
	// For the mock, we just create the booking
	_, err := r.Create(ctx, booking)
	return err
}

// Ensure the mocks implement the interfaces
var _ repository.ConcertRepository = (*MockConcertRepository)(nil)
var _ repository.BookingRepository = (*MockBookingRepository)(nil)
