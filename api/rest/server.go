package rest

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"concert-ticket-api/api/rest/handler"
	"concert-ticket-api/api/rest/middleware"
	"concert-ticket-api/internal/service"
	"concert-ticket-api/pkg/logger"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Server represents a REST API server
type Server struct {
	router         *gin.Engine
	httpServer     *http.Server
	concertHandler *handler.ConcertHandler
	bookingHandler *handler.BookingHandler
	logger         logger.Logger
}

// NewServer creates a new REST API server
func NewServer(
	concertService service.ConcertService,
	bookingService service.BookingService,
	logger logger.Logger,
	port int,
) *Server {
	// Create Gin router
	router := gin.New()

	// Set up middleware
	router.Use(gin.Recovery())
	router.Use(middleware.RequestLogger(logger))
	router.Use(middleware.RateLimiter(500)) // 500 requests per second

	// Set up CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Create handlers
	concertHandler := handler.NewConcertHandler(concertService)
	bookingHandler := handler.NewBookingHandler(bookingService)

	// Register routes
	concertHandler.RegisterRoutes(router)
	bookingHandler.RegisterRoutes(router)

	// Add health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "up"})
	})

	// Create HTTP server
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return &Server{
		router:         router,
		httpServer:     httpServer,
		concertHandler: concertHandler,
		bookingHandler: bookingHandler,
		logger:         logger,
	}
}

// Start starts the server
func (s *Server) Start() error {
	s.logger.Info("Starting REST API server on %s", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down REST API server")
	return s.httpServer.Shutdown(ctx)
}
