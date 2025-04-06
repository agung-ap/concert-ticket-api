// test/load/booking_load_test.go
package load

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"concert-ticket-api/internal/model"
	"concert-ticket-api/internal/service"
	pkgErr "concert-ticket-api/pkg/errors"
	_ "concert-ticket-api/pkg/logger"
	"concert-ticket-api/test/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// LoadTest simulates high concurrency booking scenarios
func TestConcurrentBookings(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	// Initialize logger
	//log := logger.NewLogger("info")

	// Initialize mock repositories
	concertRepo := mocks.NewMockConcertRepository()
	bookingRepo := mocks.NewMockBookingRepository()

	// Initialize services
	concertService := service.NewConcertService(concertRepo)
	bookingService := service.NewBookingService(bookingRepo, concertRepo, 3) // Use 3 retries

	// Create a test concert with a limited number of tickets
	ctx := context.Background()
	concert := &model.Concert{
		Name:             "Load Test Concert",
		Artist:           "Load Test Artist",
		Venue:            "Load Test Venue",
		ConcertDate:      time.Now().Add(24 * time.Hour),
		TotalTickets:     1000, // Limited to 1000 tickets
		Price:            50.0,
		BookingStartTime: time.Now().Add(-1 * time.Minute), // Booking started a minute ago
		BookingEndTime:   time.Now().Add(1 * time.Hour),    // Booking ends in an hour
	}

	createdConcert, err := concertService.CreateConcert(ctx, concert)
	require.NoError(t, err)

	// Number of concurrent users
	concurrentUsers := 500
	ticketsPerUser := 1 // Each user books 1 ticket
	expectedTotalBookings := concurrentUsers * ticketsPerUser

	// Make sure we don't exceed available tickets
	require.LessOrEqual(t, expectedTotalBookings, createdConcert.TotalTickets)

	// Create a wait group to synchronize goroutines
	var wg sync.WaitGroup
	wg.Add(concurrentUsers)

	// Channel to collect results
	results := make(chan error, concurrentUsers)

	// Start time
	startTime := time.Now()

	// Launch concurrent booking requests
	for i := 0; i < concurrentUsers; i++ {
		go func(userID int) {
			defer wg.Done()

			// Create booking request
			bookingReq := &model.BookingRequest{
				ConcertID:   createdConcert.ID,
				UserID:      fmt.Sprintf("loadtest-user-%d", userID),
				TicketCount: ticketsPerUser,
			}

			// Book tickets
			_, err := bookingService.BookTickets(ctx, bookingReq)
			results <- err
		}(i)
	}

	// Wait for all goroutines to finish
	wg.Wait()
	close(results)

	// End time
	duration := time.Since(startTime)

	// Analyze results
	success := 0
	failures := 0
	for err := range results {
		if err == nil {
			success++
		} else {
			failures++
			t.Logf("Booking failure: %v", err)
		}
	}

	// Get updated concert
	updatedConcert, err := concertService.GetByID(ctx, createdConcert.ID)
	require.NoError(t, err)

	// Calculate tickets booked
	ticketsBooked := createdConcert.AvailableTickets - updatedConcert.AvailableTickets

	// Log results
	t.Logf("Load Test Results:")
	t.Logf("Total Users: %d", concurrentUsers)
	t.Logf("Tickets Per User: %d", ticketsPerUser)
	t.Logf("Successful Bookings: %d", success)
	t.Logf("Failed Bookings: %d", failures)
	t.Logf("Tickets Booked: %d", ticketsBooked)
	t.Logf("Duration: %v", duration)
	t.Logf("Throughput: %.2f bookings/second", float64(success)/duration.Seconds())

	// Verify no overbooking
	assert.LessOrEqual(t, ticketsBooked, createdConcert.TotalTickets)

	// Verify high success rate (at least 95%)
	successRate := float64(success) / float64(concurrentUsers) * 100
	assert.GreaterOrEqual(t, successRate, 95.0, "Expected at least 95%% success rate, got %.2f%%", successRate)

	// Verify performance (at least 100 bookings/second)
	throughput := float64(success) / duration.Seconds()
	assert.GreaterOrEqual(t, throughput, 100.0, "Expected at least 100 bookings/second, got %.2f", throughput)
}

// TestConcurrencyRaceCondition specifically tests for race conditions
func TestConcurrencyRaceCondition(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping race condition test in short mode")
	}

	// Initialize logger
	//log := logger.NewLogger("info")

	// Initialize mock repositories
	concertRepo := mocks.NewMockConcertRepository()
	bookingRepo := mocks.NewMockBookingRepository()

	// Initialize services
	concertService := service.NewConcertService(concertRepo)
	bookingService := service.NewBookingService(bookingRepo, concertRepo, 3) // Use 3 retries

	// Create a test concert with very limited tickets
	ctx := context.Background()
	concert := &model.Concert{
		Name:             "Race Condition Test Concert",
		Artist:           "Test Artist",
		Venue:            "Test Venue",
		ConcertDate:      time.Now().Add(24 * time.Hour),
		TotalTickets:     20, // Very limited tickets
		Price:            50.0,
		BookingStartTime: time.Now().Add(-1 * time.Minute),
		BookingEndTime:   time.Now().Add(1 * time.Hour),
	}

	createdConcert, err := concertService.CreateConcert(ctx, concert)
	require.NoError(t, err)

	// Number of concurrent users (intentionally more than available tickets)
	concurrentUsers := 40 // Double the number of available tickets
	ticketsPerUser := 1   // Each user books 1 ticket

	// Create a wait group to synchronize goroutines
	var wg sync.WaitGroup
	wg.Add(concurrentUsers)

	// Channel to collect results
	results := make(chan error, concurrentUsers)

	// Launch concurrent booking requests
	for i := 0; i < concurrentUsers; i++ {
		go func(userID int) {
			defer wg.Done()

			// Create booking request
			bookingReq := &model.BookingRequest{
				ConcertID:   createdConcert.ID,
				UserID:      fmt.Sprintf("racetest-user-%d", userID),
				TicketCount: ticketsPerUser,
			}

			// Book tickets
			_, err := bookingService.BookTickets(ctx, bookingReq)
			results <- err
		}(i)
	}

	// Wait for all goroutines to finish
	wg.Wait()
	close(results)

	// Analyze results
	success := 0
	insufficientTickets := 0
	otherErrors := 0

	for err := range results {
		if err == nil {
			success++
		} else if errors.Is(err, pkgErr.ErrInsufficientTickets) {
			insufficientTickets++
		} else {
			otherErrors++
			t.Logf("Unexpected error: %v", err)
		}
	}

	// Get updated concert
	updatedConcert, err := concertService.GetByID(ctx, createdConcert.ID)
	require.NoError(t, err)

	// Calculate tickets booked
	ticketsBooked := createdConcert.AvailableTickets - updatedConcert.AvailableTickets

	// Log results
	t.Logf("Race Condition Test Results:")
	t.Logf("Total Users: %d", concurrentUsers)
	t.Logf("Available Tickets: %d", createdConcert.TotalTickets)
	t.Logf("Successful Bookings: %d", success)
	t.Logf("Insufficient Ticket Errors: %d", insufficientTickets)
	t.Logf("Other Errors: %d", otherErrors)
	t.Logf("Tickets Booked: %d", ticketsBooked)

	// Verify no overbooking
	assert.Equal(t, success, ticketsBooked, "Tickets booked should match successful bookings")
	assert.Equal(t, createdConcert.TotalTickets, ticketsBooked, "All available tickets should be booked")

	// Verify we got the expected number of insufficient tickets errors
	assert.Equal(t, concurrentUsers-createdConcert.TotalTickets, insufficientTickets,
		"Expected %d insufficient ticket errors", concurrentUsers-createdConcert.TotalTickets)

	// Verify no unexpected errors
	assert.Zero(t, otherErrors, "Expected no other errors")
}
