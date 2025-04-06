package integration

import (
	"context"
	"testing"
	"time"

	"concert-ticket-api/internal/model"
	"concert-ticket-api/internal/repository"
	"concert-ticket-api/internal/repository/postgres"
	"concert-ticket-api/internal/service"
	"concert-ticket-api/test/testutil"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ConcertServiceTestSuite struct {
	suite.Suite
	db             *sqlx.DB
	concertRepo    repository.ConcertRepository
	concertService service.ConcertService
}

func (s *ConcertServiceTestSuite) SetupSuite() {
	// Connect to test database
	var err error
	s.db, err = testutil.SetupTestDB()
	require.NoError(s.T(), err)

	// Initialize repositories and services
	s.concertRepo = postgres.NewConcertRepository(s.db)
	s.concertService = service.NewConcertService(s.concertRepo)
}

func (s *ConcertServiceTestSuite) TearDownTest() {
	// Clean up database after each test
	testutil.CleanupTestDB(s.db)
}

func (s *ConcertServiceTestSuite) TearDownSuite() {
	// Close database connection
	s.db.Close()
}

func (s *ConcertServiceTestSuite) TestCreateConcert() {
	ctx := context.Background()

	// Create a new concert
	concert := &model.Concert{
		Name:             "Test Concert",
		Artist:           "Test Artist",
		Venue:            "Test Venue",
		ConcertDate:      time.Now().Add(24 * time.Hour),
		TotalTickets:     100,
		Price:            50.0,
		BookingStartTime: time.Now().Add(1 * time.Hour),
		BookingEndTime:   time.Now().Add(2 * time.Hour),
	}

	createdConcert, err := s.concertService.CreateConcert(ctx, concert)
	require.NoError(s.T(), err)
	assert.NotZero(s.T(), createdConcert.ID)
	assert.Equal(s.T(), concert.Name, createdConcert.Name)
	assert.Equal(s.T(), concert.Artist, createdConcert.Artist)
	assert.Equal(s.T(), concert.Venue, createdConcert.Venue)
	assert.Equal(s.T(), concert.TotalTickets, createdConcert.TotalTickets)
	assert.Equal(s.T(), concert.TotalTickets, createdConcert.AvailableTickets)
	assert.Equal(s.T(), concert.Price, createdConcert.Price)
	assert.NotZero(s.T(), createdConcert.CreatedAt)
	assert.NotZero(s.T(), createdConcert.UpdatedAt)
}

func (s *ConcertServiceTestSuite) TestGetConcert() {
	ctx := context.Background()

	// Create a new concert
	concert := &model.Concert{
		Name:             "Test Concert",
		Artist:           "Test Artist",
		Venue:            "Test Venue",
		ConcertDate:      time.Now().Add(24 * time.Hour),
		TotalTickets:     100,
		Price:            50.0,
		BookingStartTime: time.Now().Add(1 * time.Hour),
		BookingEndTime:   time.Now().Add(2 * time.Hour),
	}

	createdConcert, err := s.concertService.CreateConcert(ctx, concert)
	require.NoError(s.T(), err)

	// Get the concert
	retrievedConcert, err := s.concertService.GetByID(ctx, createdConcert.ID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), createdConcert.ID, retrievedConcert.ID)
	assert.Equal(s.T(), createdConcert.Name, retrievedConcert.Name)
}

func (s *ConcertServiceTestSuite) TestListConcerts() {
	ctx := context.Background()

	// Create multiple concerts
	concerts := []*model.Concert{
		{
			Name:             "Concert 1",
			Artist:           "Artist 1",
			Venue:            "Venue 1",
			ConcertDate:      time.Now().Add(24 * time.Hour),
			TotalTickets:     100,
			Price:            50.0,
			BookingStartTime: time.Now().Add(1 * time.Hour),
			BookingEndTime:   time.Now().Add(2 * time.Hour),
		},
		{
			Name:             "Concert 2",
			Artist:           "Artist 2",
			Venue:            "Venue 2",
			ConcertDate:      time.Now().Add(48 * time.Hour),
			TotalTickets:     200,
			Price:            75.0,
			BookingStartTime: time.Now().Add(3 * time.Hour),
			BookingEndTime:   time.Now().Add(4 * time.Hour),
		},
	}

	for _, c := range concerts {
		_, err := s.concertService.CreateConcert(ctx, c)
		require.NoError(s.T(), err)
	}

	// List concerts
	listedConcerts, count, err := s.concertService.ListConcerts(ctx, 1, 10, nil)
	require.NoError(s.T(), err)
	assert.GreaterOrEqual(s.T(), len(listedConcerts), 2)
	assert.GreaterOrEqual(s.T(), count, 2)

	// Test filtering by artist
	filteredConcerts, filteredCount, err := s.concertService.ListConcerts(ctx, 1, 10, map[string]interface{}{
		"artist": "Artist 1",
	})
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 1, len(filteredConcerts))
	assert.Equal(s.T(), 1, filteredCount)
	assert.Equal(s.T(), "Artist 1", filteredConcerts[0].Artist)
}

func (s *ConcertServiceTestSuite) TestUpdateConcert() {
	ctx := context.Background()

	// Create a new concert
	concert := &model.Concert{
		Name:             "Test Concert",
		Artist:           "Test Artist",
		Venue:            "Test Venue",
		ConcertDate:      time.Now().Add(24 * time.Hour),
		TotalTickets:     100,
		Price:            50.0,
		BookingStartTime: time.Now().Add(1 * time.Hour),
		BookingEndTime:   time.Now().Add(2 * time.Hour),
	}

	createdConcert, err := s.concertService.CreateConcert(ctx, concert)
	require.NoError(s.T(), err)

	// Update the concert
	createdConcert.Name = "Updated Concert"
	createdConcert.Price = 75.0

	err = s.concertService.UpdateConcert(ctx, createdConcert)
	require.NoError(s.T(), err)

	// Get the updated concert
	updatedConcert, err := s.concertService.GetByID(ctx, createdConcert.ID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "Updated Concert", updatedConcert.Name)
	assert.Equal(s.T(), 75.0, updatedConcert.Price)
}

func TestConcertService(t *testing.T) {
	suite.Run(t, new(ConcertServiceTestSuite))
}
