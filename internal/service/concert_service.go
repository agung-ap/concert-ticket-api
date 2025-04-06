package service

import (
	"context"
	"time"

	"concert-ticket-api/internal/model"
	"concert-ticket-api/internal/repository"
	"concert-ticket-api/pkg/errors"
)

// ConcertService defines the interface for concert operations
type ConcertService interface {
	// GetByID retrieves a concert by its ID
	GetByID(ctx context.Context, id int64) (*model.Concert, error)

	// ListConcerts retrieves concerts with filtering and pagination
	ListConcerts(ctx context.Context, page, pageSize int, filters map[string]interface{}) ([]*model.Concert, int, error)

	// CreateConcert creates a new concert
	CreateConcert(ctx context.Context, concert *model.Concert) (*model.Concert, error)

	// UpdateConcert updates an existing concert
	UpdateConcert(ctx context.Context, concert *model.Concert) error
}

type concertService struct {
	concertRepo repository.ConcertRepository
}

// NewConcertService creates a new implementation of ConcertService
func NewConcertService(concertRepo repository.ConcertRepository) ConcertService {
	return &concertService{
		concertRepo: concertRepo,
	}
}

// GetByID retrieves a concert by its ID
func (s *concertService) GetByID(ctx context.Context, id int64) (*model.Concert, error) {
	return s.concertRepo.GetByID(ctx, id)
}

// ListConcerts retrieves concerts with filtering and pagination
func (s *concertService) ListConcerts(ctx context.Context, page, pageSize int, filters map[string]interface{}) ([]*model.Concert, int, error) {
	if page < 1 {
		page = 1
	}

	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	// Get total count for pagination
	totalCount, err := s.concertRepo.Count(ctx, filters)
	if err != nil {
		return nil, 0, err
	}

	// Get concerts for current page
	concerts, err := s.concertRepo.List(ctx, pageSize, offset, filters)
	if err != nil {
		return nil, 0, err
	}

	return concerts, totalCount, nil
}

// CreateConcert creates a new concert
func (s *concertService) CreateConcert(ctx context.Context, concert *model.Concert) (*model.Concert, error) {
	// Validate concert data
	if err := validateConcert(concert); err != nil {
		return nil, err
	}

	// Set initial available tickets equal to total tickets
	concert.AvailableTickets = concert.TotalTickets

	return s.concertRepo.Create(ctx, concert)
}

// UpdateConcert updates an existing concert
func (s *concertService) UpdateConcert(ctx context.Context, concert *model.Concert) error {
	// Validate concert data
	if err := validateConcert(concert); err != nil {
		return err
	}

	// Check if the concert exists
	_, err := s.concertRepo.GetByID(ctx, concert.ID)
	if err != nil {
		return err
	}

	return s.concertRepo.Update(ctx, concert)
}

// validateConcert validates concert data
func validateConcert(concert *model.Concert) error {
	if concert.Name == "" {
		return errors.ErrInvalidInput("name is required")
	}

	if concert.Artist == "" {
		return errors.ErrInvalidInput("artist is required")
	}

	if concert.Venue == "" {
		return errors.ErrInvalidInput("venue is required")
	}

	if concert.ConcertDate.IsZero() {
		return errors.ErrInvalidInput("concert date is required")
	}

	if concert.TotalTickets <= 0 {
		return errors.ErrInvalidInput("total tickets must be positive")
	}

	if concert.Price < 0 {
		return errors.ErrInvalidInput("price cannot be negative")
	}

	if concert.BookingStartTime.IsZero() {
		return errors.ErrInvalidInput("booking start time is required")
	}

	if concert.BookingEndTime.IsZero() {
		return errors.ErrInvalidInput("booking end time is required")
	}

	if concert.BookingEndTime.Before(concert.BookingStartTime) {
		return errors.ErrInvalidInput("booking end time must be after booking start time")
	}

	if concert.BookingStartTime.Before(time.Now()) && concert.ID == 0 {
		return errors.ErrInvalidInput("booking start time must be in the future for new concerts")
	}

	return nil
}
