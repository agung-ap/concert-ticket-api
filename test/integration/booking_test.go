package integration

import (
	"context"
	"errors"
	"testing"
	"time"

	"concert-ticket-api/internal/model"
	"concert-ticket-api/internal/repository"
	"concert-ticket-api/internal/repository/postgres"
	"concert-ticket-api/internal/service"
	pkgErr "concert-ticket-api/pkg/errors"
	"concert-ticket-api/test/testutil"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type BookingServiceTestSuite struct {
	suite.Suite
	db             *sqlx.DB
	concertRepo    repository.ConcertRepository
	bookingRepo    repository.BookingRepository
	concertService service.ConcertService
	bookingService service.BookingService
}

func (s *BookingServiceTestSuite) SetupSuite() {
	// Connect to test database
	var err error
	s.db, err = testutil.SetupTestDB()
	require.NoError(s.T(), err)

	// Initialize repositories and services
	s.concertRepo = postgres.NewConcertRepository(s.db)
	s.bookingRepo = postgres.NewBookingRepository(s.db)
	s.concertService = service.NewConcertService(s.concertRepo)
	s.bookingService = service.NewBookingService(s.bookingRepo, s.concertRepo, 3)
}

func (s *BookingServiceTestSuite) TearDownTest() {
	// Clean up database after each test
	testutil.CleanupTestDB(s.db)
}

func (s *BookingServiceTestSuite) TearDownSuite() {
	// Close database connection
	s.db.Close()
}

func (s *BookingServiceTestSuite) createTestConcert() *model.Concert {
	ctx := context.Background()

	// Create a new concert with booking window open now
	concert := &model.Concert{
		Name:             "Test Concert",
		Artist:           "Test Artist",
		Venue:            "Test Venue",
		ConcertDate:      time.Now().Add(24 * time.Hour),
		TotalTickets:     100,
		Price:            50.0,
		BookingStartTime: time.Now().Add(-1 * time.Hour), // Booking started an hour ago
		BookingEndTime:   time.Now().Add(2 * time.Hour),  // Booking ends in 2 hours
	}

	createdConcert, err := s.concertService.CreateConcert(ctx, concert)
	require.NoError(s.T(), err)

	return createdConcert
}

func (s *BookingServiceTestSuite) TestBookTickets() {
	ctx := context.Background()

	// Create a concert
	concert := s.createTestConcert()

	// Book tickets
	bookingReq := &model.BookingRequest{
		ConcertID:   concert.ID,
		UserID:      "test-user",
		TicketCount: 5,
	}

	booking, err := s.bookingService.BookTickets(ctx, bookingReq)
	require.NoError(s.T(), err)
	assert.NotZero(s.T(), booking.ID)
	assert.Equal(s.T(), concert.ID, booking.ConcertID)
	assert.Equal(s.T(), "test-user", booking.UserID)
	assert.Equal(s.T(), 5, booking.TicketCount)
	assert.Equal(s.T(), model.BookingStatusConfirmed, booking.Status)

	// Check that tickets were deducted from available tickets
	updatedConcert, err := s.concertService.GetByID(ctx, concert.ID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), concert.AvailableTickets-5, updatedConcert.AvailableTickets)
}

func (s *BookingServiceTestSuite) TestBookTooManyTickets() {
	ctx := context.Background()

	// Create a concert
	concert := s.createTestConcert()

	// Try to book more tickets than available
	bookingReq := &model.BookingRequest{
		ConcertID:   concert.ID,
		UserID:      "test-user",
		TicketCount: concert.TotalTickets + 1,
	}

	_, err := s.bookingService.BookTickets(ctx, bookingReq)
	require.Error(s.T(), err)
	assert.True(s.T(), errors.Is(err, pkgErr.ErrInsufficientTickets))

	// Check that no tickets were deducted
	updatedConcert, err := s.concertService.GetByID(ctx, concert.ID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), concert.AvailableTickets, updatedConcert.AvailableTickets)
}

func (s *BookingServiceTestSuite) TestBookClosedConcert() {
	ctx := context.Background()

	// Create a concert with booking window closed
	concert := &model.Concert{
		Name:             "Closed Concert",
		Artist:           "Test Artist",
		Venue:            "Test Venue",
		ConcertDate:      time.Now().Add(24 * time.Hour),
		TotalTickets:     100,
		Price:            50.0,
		BookingStartTime: time.Now().Add(1 * time.Hour), // Booking starts in 1 hour
		BookingEndTime:   time.Now().Add(2 * time.Hour),
	}

	createdConcert, err := s.concertService.CreateConcert(ctx, concert)
	require.NoError(s.T(), err)

	// Try to book tickets
	bookingReq := &model.BookingRequest{
		ConcertID:   createdConcert.ID,
		UserID:      "test-user",
		TicketCount: 5,
	}

	_, err = s.bookingService.BookTickets(ctx, bookingReq)
	require.Error(s.T(), err)
	assert.True(s.T(), errors.Is(err, pkgErr.ErrBookingClosed))
}

func (s *BookingServiceTestSuite) TestCancelBooking() {
	ctx := context.Background()

	// Create a concert
	concert := s.createTestConcert()

	// Book tickets
	bookingReq := &model.BookingRequest{
		ConcertID:   concert.ID,
		UserID:      "test-user",
		TicketCount: 5,
	}

	booking, err := s.bookingService.BookTickets(ctx, bookingReq)
	require.NoError(s.T(), err)

	// Cancel booking
	err = s.bookingService.CancelBooking(ctx, booking.ID, "test-user")
	require.NoError(s.T(), err)

	// Check that the booking was cancelled
	updatedBooking, err := s.bookingService.GetBookingByID(ctx, booking.ID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), model.BookingStatusCancelled, updatedBooking.Status)

	// Check that tickets were returned to available tickets
	updatedConcert, err := s.concertService.GetByID(ctx, concert.ID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), concert.AvailableTickets, updatedConcert.AvailableTickets)
}

func (s *BookingServiceTestSuite) TestCancelOtherUserBooking() {
	ctx := context.Background()

	// Create a concert
	concert := s.createTestConcert()

	// Book tickets
	bookingReq := &model.BookingRequest{
		ConcertID:   concert.ID,
		UserID:      "test-user",
		TicketCount: 5,
	}

	booking, err := s.bookingService.BookTickets(ctx, bookingReq)
	require.NoError(s.T(), err)

	// Try to cancel booking as a different user
	err = s.bookingService.CancelBooking(ctx, booking.ID, "other-user")
	require.Error(s.T(), err)
	assert.True(s.T(), errors.Is(err, pkgErr.ErrUnauthorized))
}

func (s *BookingServiceTestSuite) TestGetUserBookings() {
	ctx := context.Background()

	// Create a concert
	concert := s.createTestConcert()

	// Book multiple tickets for the same user
	for i := 0; i < 3; i++ {
		bookingReq := &model.BookingRequest{
			ConcertID:   concert.ID,
			UserID:      "test-user",
			TicketCount: 1,
		}

		_, err := s.bookingService.BookTickets(ctx, bookingReq)
		require.NoError(s.T(), err)
	}

	// Get user bookings
	bookings, err := s.bookingService.GetUserBookings(ctx, "test-user", 1, 10)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 3, len(bookings))

	// Check pagination
	limitedBookings, err := s.bookingService.GetUserBookings(ctx, "test-user", 1, 2)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 2, len(limitedBookings))
}

func TestBookingService(t *testing.T) {
	suite.Run(t, new(BookingServiceTestSuite))
}
