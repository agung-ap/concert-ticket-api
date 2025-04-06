package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"concert-ticket-api/internal/model"
	"concert-ticket-api/internal/service"
	pkgErr "concert-ticket-api/pkg/errors"

	"github.com/gin-gonic/gin"
)

// ConcertHandler handles HTTP requests related to concerts
type ConcertHandler struct {
	concertService service.ConcertService
}

// NewConcertHandler creates a new ConcertHandler
func NewConcertHandler(concertService service.ConcertService) *ConcertHandler {
	return &ConcertHandler{
		concertService: concertService,
	}
}

// RegisterRoutes registers the routes for this handler
func (h *ConcertHandler) RegisterRoutes(router *gin.Engine) {
	concertGroup := router.Group("/api/v1/concerts")
	{
		concertGroup.GET("", h.ListConcerts)
		concertGroup.GET("/:id", h.GetConcert)
		concertGroup.POST("", h.CreateConcert)
		concertGroup.PUT("/:id", h.UpdateConcert)
	}
}

// GetConcert handles GET /api/v1/concerts/:id requests
func (h *ConcertHandler) GetConcert(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid concert ID"})
		return
	}

	concert, err := h.concertService.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, pkgErr.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Concert not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get concert"})
		return
	}

	c.JSON(http.StatusOK, concert)
}

// ListConcerts handles GET /api/v1/concerts requests
func (h *ConcertHandler) ListConcerts(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	// Parse filter parameters
	filters := make(map[string]interface{})

	if artist := c.Query("artist"); artist != "" {
		filters["artist"] = artist
	}

	if venue := c.Query("venue"); venue != "" {
		filters["venue"] = venue
	}

	if name := c.Query("name"); name != "" {
		filters["name"] = name
	}

	if dateFromStr := c.Query("dateFrom"); dateFromStr != "" {
		dateFrom, err := time.Parse(time.RFC3339, dateFromStr)
		if err == nil {
			filters["date_from"] = dateFrom
		}
	}

	if dateToStr := c.Query("dateTo"); dateToStr != "" {
		dateTo, err := time.Parse(time.RFC3339, dateToStr)
		if err == nil {
			filters["date_to"] = dateTo
		}
	}

	if availableOnly := c.Query("availableOnly"); availableOnly == "true" {
		filters["available"] = true
	}

	concerts, totalCount, err := h.concertService.ListConcerts(c.Request.Context(), page, pageSize, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list concerts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": concerts,
		"meta": gin.H{
			"page":       page,
			"pageSize":   pageSize,
			"totalCount": totalCount,
			"totalPages": (totalCount + pageSize - 1) / pageSize,
		},
	})
}

// CreateConcert handles POST /api/v1/concerts requests
func (h *ConcertHandler) CreateConcert(c *gin.Context) {
	var concert model.Concert
	if err := c.ShouldBindJSON(&concert); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid concert data"})
		return
	}

	createdConcert, err := h.concertService.CreateConcert(c.Request.Context(), &concert)
	if err != nil {
		if errWithMsg, ok := err.(*pkgErr.ErrorWithMessage); ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": errWithMsg.Message()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create concert"})
		return
	}

	c.JSON(http.StatusCreated, createdConcert)
}

// UpdateConcert handles PUT /api/v1/concerts/:id requests
func (h *ConcertHandler) UpdateConcert(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid concert ID"})
		return
	}

	var concert model.Concert
	if err := c.ShouldBindJSON(&concert); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid concert data"})
		return
	}

	concert.ID = id

	err = h.concertService.UpdateConcert(c.Request.Context(), &concert)
	if err != nil {
		if errors.Is(err, pkgErr.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Concert not found"})
			return
		}
		if errWithMsg, ok := err.(*pkgErr.ErrorWithMessage); ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": errWithMsg.Message()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update concert"})
		return
	}

	c.JSON(http.StatusOK, concert)
}
