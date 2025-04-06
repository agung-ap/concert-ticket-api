package service

import (
	"concert-ticket-api/internal/model"
	"concert-ticket-api/internal/repository"
	pkgErr "concert-ticket-api/pkg/errors"
	"context"
	"errors"
	"fmt"
	"time"
)

// BookingService defines the interface for booking operations
type BookingService interface {
	// GetBookingByID retrieves a booking by its ID
	GetBookingByID(ctx context.Context, id int64) (*model.Booking, error)

	// GetUserBookings retrieves bookings for a user
	GetUserBookings(ctx context.Context, userID string, page, pageSize int) ([]*model.Booking, error)

	// BookTickets books tickets for a concert
	BookTickets(ctx context.Context, req *model.BookingRequest) (*model.Booking, error)

	// CancelBooking cancels a booking
	CancelBooking(ctx context.Context, bookingID int64, userID string) error
}

type bookingService struct {
	bookingRepo repository.BookingRepository
	concertRepo repository.ConcertRepository
	maxRetries  int
}

// NewBookingService creates a new implementation of BookingService
func NewBookingService(
	bookingRepo repository.BookingRepository,
	concertRepo repository.ConcertRepository,
	maxRetries int,
) BookingService {
	if maxRetries <= 0 {
		maxRetries = 3 // Default to 3 retries
	}

	return &bookingService{
		bookingRepo: bookingRepo,
		concertRepo: concertRepo,
		maxRetries:  maxRetries,
	}
}

// GetBookingByID retrieves a booking by its ID
func (s *bookingService) GetBookingByID(ctx context.Context, id int64) (*model.Booking, error) {
	return s.bookingRepo.GetByID(ctx, id)
}

// GetUserBookings retrieves bookings for a user
func (s *bookingService) GetUserBookings(ctx context.Context, userID string, page, pageSize int) ([]*model.Booking, error) {
	if page < 1 {
		page = 1
	}

	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	return s.bookingRepo.GetByUserID(ctx, userID, pageSize, offset)
}

// BookTickets books tickets for a concert
func (s *bookingService) BookTickets(ctx context.Context, req *model.BookingRequest) (*model.Booking, error) {
	// Validate booking request
	if err := validateBookingRequest(req); err != nil {
		return nil, err
	}

	// Get the concert
	concert, err := s.concertRepo.GetByID(ctx, req.ConcertID)
	if err != nil {
		return nil, err
	}

	// Check if booking is open
	if !concert.IsBookingOpen() {
		return nil, pkgErr.ErrBookingClosed
	}

	// Check if there are enough tickets
	if !concert.HasAvailableTickets(req.TicketCount) {
		return nil, pkgErr.ErrInsufficientTickets
	}

	// Create booking with retries for handling concurrent requests
	booking := &model.Booking{
		ConcertID:   req.ConcertID,
		UserID:      req.UserID,
		TicketCount: req.TicketCount,
		Status:      model.BookingStatusConfirmed,
		BookingTime: time.Now(),
	}

	var lastErr error

	// Retry loop for concurrent booking attempts
	for attempt := 0; attempt < s.maxRetries; attempt++ {
		// Get the latest concert state with FOR UPDATE lock
		concertForUpdate, err := s.concertRepo.GetForUpdate(ctx, req.ConcertID)
		if err != nil {
			return nil, err
		}

		// Check if booking is still open
		if !concertForUpdate.IsBookingOpen() {
			return nil, pkgErr.ErrBookingClosed
		}

		// Check if there are still enough tickets
		if !concertForUpdate.HasAvailableTickets(req.TicketCount) {
			return nil, pkgErr.ErrInsufficientTickets
		}

		// Create booking and update ticket count in a transaction
		err = s.bookingRepo.CreateWithTicketUpdate(ctx, booking, concertForUpdate.Version)
		if err == nil {
			// Success!
			return booking, nil
		}

		// If we encounter a version conflict, we'll retry
		if errors.Is(err, pkgErr.ErrOptimisticLockFailed) {
			lastErr = err
			// Add a small delay before retrying to reduce contention
			time.Sleep(time.Duration(attempt+1) * 10 * time.Millisecond)
			continue
		}

		// For other errors, return immediately
		return nil, err
	}

	// If we get here, we've exhausted our retries
	return nil, fmt.Errorf("failed to book tickets after %d attempts: %w", s.maxRetries, lastErr)
}

// CancelBooking cancels a booking
func (s *bookingService) CancelBooking(ctx context.Context, bookingID int64, userID string) error {
	// Get the booking
	booking, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return err
	}

	// Check if the booking belongs to the user
	if booking.UserID != userID {
		return pkgErr.ErrUnauthorized
	}

	// Check if the booking is already cancelled
	if booking.Status == model.BookingStatusCancelled {
		return pkgErr.ErrBookingAlreadyCancelled
	}

	// Start a transaction
	tx, err := s.concertRepo.GetDB().BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Update booking status
	booking.Status = model.BookingStatusCancelled
	if err = s.bookingRepo.Update(ctx, booking); err != nil {
		return err
	}

	// Return the tickets to the available pool
	concert, err := s.concertRepo.GetForUpdate(ctx, booking.ConcertID)
	if err != nil {
		return err
	}

	concert.AvailableTickets += booking.TicketCount
	if err = s.concertRepo.Update(ctx, concert); err != nil {
		return err
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// validateBookingRequest validates booking request data
func validateBookingRequest(req *model.BookingRequest) error {
	if req.ConcertID <= 0 {
		return pkgErr.ErrInvalidInput("concert_id is required")
	}

	if req.UserID == "" {
		return pkgErr.ErrInvalidInput("user_id is required")
	}

	if req.TicketCount <= 0 {
		return pkgErr.ErrInvalidInput("ticket_count must be positive")
	}

	if req.TicketCount > 10 {
		return pkgErr.ErrInvalidInput("cannot book more than 10 tickets at once")
	}

	return nil
}
