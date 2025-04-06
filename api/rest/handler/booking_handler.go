package handler

import (
	"errors"
	"net/http"
	"strconv"

	"concert-ticket-api/internal/model"
	"concert-ticket-api/internal/service"
	pkgErr "concert-ticket-api/pkg/errors"

	"github.com/gin-gonic/gin"
)

// BookingHandler handles HTTP requests related to bookings
type BookingHandler struct {
	bookingService service.BookingService
}

// NewBookingHandler creates a new BookingHandler
func NewBookingHandler(bookingService service.BookingService) *BookingHandler {
	return &BookingHandler{
		bookingService: bookingService,
	}
}

// RegisterRoutes registers the routes for this handler
func (h *BookingHandler) RegisterRoutes(router *gin.Engine) {
	bookingGroup := router.Group("/api/v1/bookings")
	{
		bookingGroup.POST("", h.BookTickets)
		bookingGroup.GET("", h.GetUserBookings)
		bookingGroup.GET("/:id", h.GetBooking)
		bookingGroup.POST("/:id/cancel", h.CancelBooking)
	}
}

// BookTickets handles POST /api/v1/bookings requests
func (h *BookingHandler) BookTickets(c *gin.Context) {
	var req model.BookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid booking data"})
		return
	}

	// In a real app, userID would come from auth middleware
	// For this exercise, we'll use the one in the request

	booking, err := h.bookingService.BookTickets(c.Request.Context(), &req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorMsg := "Failed to book tickets"

		// Map specific errors to appropriate HTTP status codes
		switch {
		case errors.Is(err, pkgErr.ErrInvalidInput("")):
			statusCode = http.StatusBadRequest
			if errWithMsg, ok := err.(*pkgErr.ErrorWithMessage); ok {
				errorMsg = errWithMsg.Message()
			} else {
				errorMsg = "Invalid booking request"
			}
		case errors.Is(err, pkgErr.ErrNotFound):
			statusCode = http.StatusNotFound
			errorMsg = "Concert not found"
		case errors.Is(err, pkgErr.ErrBookingClosed):
			statusCode = http.StatusBadRequest
			errorMsg = "Booking is not open for this concert"
		case errors.Is(err, pkgErr.ErrInsufficientTickets):
			statusCode = http.StatusBadRequest
			errorMsg = "Not enough tickets available"
		case errors.Is(err, pkgErr.ErrOptimisticLockFailed):
			statusCode = http.StatusConflict
			errorMsg = "Booking conflict, please try again"
		}

		c.JSON(statusCode, gin.H{"error": errorMsg})
		return
	}

	c.JSON(http.StatusCreated, booking)
}

// GetBooking handles GET /api/v1/bookings/:id requests
func (h *BookingHandler) GetBooking(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid booking ID"})
		return
	}

	booking, err := h.bookingService.GetBookingByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, pkgErr.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get booking"})
		return
	}

	// In a real app, check if the booking belongs to the authenticated user
	// For this exercise, we'll skip this check

	c.JSON(http.StatusOK, booking)
}

// GetUserBookings handles GET /api/v1/bookings requests
func (h *BookingHandler) GetUserBookings(c *gin.Context) {
	// In a real app, userID would come from auth middleware
	// For this exercise, we'll use a query parameter
	userID := c.Query("userID")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	bookings, err := h.bookingService.GetUserBookings(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user bookings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": bookings,
		"meta": gin.H{
			"page":     page,
			"pageSize": pageSize,
		},
	})
}

// CancelBooking handles POST /api/v1/bookings/:id/cancel requests
func (h *BookingHandler) CancelBooking(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid booking ID"})
		return
	}

	// In a real app, userID would come from auth middleware
	// For this exercise, we'll use a JSON request body
	var req struct {
		UserID string `json:"userID"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	err = h.bookingService.CancelBooking(c.Request.Context(), id, req.UserID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorMsg := "Failed to cancel booking"

		// Map specific errors to appropriate HTTP status codes
		switch {
		case errors.Is(err, pkgErr.ErrNotFound):
			statusCode = http.StatusNotFound
			errorMsg = "Booking not found"
		case errors.Is(err, pkgErr.ErrUnauthorized):
			statusCode = http.StatusForbidden
			errorMsg = "You are not authorized to cancel this booking"
		case errors.Is(err, pkgErr.ErrBookingAlreadyCancelled):
			statusCode = http.StatusBadRequest
			errorMsg = "Booking is already cancelled"
		}

		c.JSON(statusCode, gin.H{"error": errorMsg})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Booking cancelled successfully"})
}
