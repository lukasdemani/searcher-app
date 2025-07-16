package handlers

import (
	"context"
	"math"
	"net/http"
	"strconv"
	"time"

	"seacher-app/internal/models"
	"seacher-app/internal/services"

	"github.com/gin-gonic/gin"
)

type URLHandler struct {
	crawlerService services.CrawlerService
	wsHandler      *WebSocketHandler
}

func NewURLHandler(crawlerService services.CrawlerService, wsHandler *WebSocketHandler) *URLHandler {
	return &URLHandler{
		crawlerService: crawlerService,
		wsHandler:      wsHandler,
	}
}

func (h *URLHandler) GetURLs(c *gin.Context) {
	// Create context with timeout for database operations
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	search := c.DefaultQuery("search", "")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Check if context is already cancelled
	select {
	case <-ctx.Done():
		c.JSON(http.StatusRequestTimeout, models.ErrorResponse{Error: "Request timeout"})
		return
	default:
	}

	urls, total, err := h.crawlerService.GetURLs(page, limit, search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	response := models.PaginatedResponse{
		Data:       urls,
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}

	c.JSON(http.StatusOK, response)
}

func (h *URLHandler) CreateURL(c *gin.Context) {
	// Create context with timeout for URL creation
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	var req models.URLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	// Check if context is already cancelled
	select {
	case <-ctx.Done():
		c.JSON(http.StatusRequestTimeout, models.ErrorResponse{Error: "Request timeout"})
		return
	default:
	}

	url, err := h.crawlerService.AddURL(req.URL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	// Start analysis in background with context and broadcast status updates
	go func() {
		// Create background context for analysis (not cancelled by request)
		bgCtx, bgCancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer bgCancel()

		// Broadcast processing status
		h.wsHandler.BroadcastStatusUpdate(url.ID, "processing", nil)

		// Perform analysis
		err := h.crawlerService.AnalyzeURL(url.ID)

		// Get updated URL data
		updatedURL, getErr := h.crawlerService.GetURL(url.ID)
		if getErr != nil {
			h.wsHandler.BroadcastStatusUpdate(url.ID, "error", map[string]string{"error": getErr.Error()})
			return
		}

		// Broadcast final status
		if err != nil {
			h.wsHandler.BroadcastStatusUpdate(url.ID, "error", updatedURL)
		} else {
			h.wsHandler.BroadcastStatusUpdate(url.ID, "completed", updatedURL)
		}

		// Check if background context was cancelled
		select {
		case <-bgCtx.Done():
			h.wsHandler.BroadcastStatusUpdate(url.ID, "timeout", map[string]string{"error": "Analysis timeout"})
		default:
		}
	}()

	c.JSON(http.StatusCreated, models.SuccessResponse{
		Message: "URL added successfully",
		Data:    url,
	})
}

func (h *URLHandler) GetURL(c *gin.Context) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid URL ID"})
		return
	}

	// Check if context is already cancelled
	select {
	case <-ctx.Done():
		c.JSON(http.StatusRequestTimeout, models.ErrorResponse{Error: "Request timeout"})
		return
	default:
	}

	url, err := h.crawlerService.GetURL(id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "URL not found"})
		return
	}

	c.JSON(http.StatusOK, url)
}

func (h *URLHandler) AnalyzeURL(c *gin.Context) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid URL ID"})
		return
	}

	// Check if context is already cancelled
	select {
	case <-ctx.Done():
		c.JSON(http.StatusRequestTimeout, models.ErrorResponse{Error: "Request timeout"})
		return
	default:
	}

	// Start analysis in background with WebSocket updates
	go func() {
		// Create background context for analysis (not cancelled by request)
		bgCtx, bgCancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer bgCancel()

		// Broadcast processing status
		h.wsHandler.BroadcastStatusUpdate(id, "processing", nil)

		// Perform analysis
		err := h.crawlerService.AnalyzeURL(id)

		// Get updated URL data
		updatedURL, getErr := h.crawlerService.GetURL(id)
		if getErr != nil {
			h.wsHandler.BroadcastStatusUpdate(id, "error", map[string]string{"error": getErr.Error()})
			return
		}

		// Broadcast final status
		if err != nil {
			h.wsHandler.BroadcastStatusUpdate(id, "error", updatedURL)
		} else {
			h.wsHandler.BroadcastStatusUpdate(id, "completed", updatedURL)
		}

		// Check if background context was cancelled
		select {
		case <-bgCtx.Done():
			h.wsHandler.BroadcastStatusUpdate(id, "timeout", map[string]string{"error": "Analysis timeout"})
		default:
		}
	}()

	c.JSON(http.StatusOK, models.SuccessResponse{Message: "URL analysis started"})
}

func (h *URLHandler) DeleteURL(c *gin.Context) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid URL ID"})
		return
	}

	// Check if context is already cancelled
	select {
	case <-ctx.Done():
		c.JSON(http.StatusRequestTimeout, models.ErrorResponse{Error: "Request timeout"})
		return
	default:
	}

	err = h.crawlerService.DeleteURL(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	// Broadcast deletion
	h.wsHandler.BroadcastStatusUpdate(id, "deleted", nil)

	c.JSON(http.StatusOK, models.SuccessResponse{Message: "URL deleted successfully"})
}

func (h *URLHandler) BulkAnalyze(c *gin.Context) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var req models.BulkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	// Check if context is already cancelled
	select {
	case <-ctx.Done():
		c.JSON(http.StatusRequestTimeout, models.ErrorResponse{Error: "Request timeout"})
		return
	default:
	}

	for _, id := range req.IDs {
		go func(urlID int) {
			// Create background context for analysis (not cancelled by request)
			bgCtx, bgCancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer bgCancel()

			// Broadcast processing status
			h.wsHandler.BroadcastStatusUpdate(urlID, "processing", nil)

			// Perform analysis
			err := h.crawlerService.AnalyzeURL(urlID)

			// Get updated URL data
			updatedURL, getErr := h.crawlerService.GetURL(urlID)
			if getErr != nil {
				h.wsHandler.BroadcastStatusUpdate(urlID, "error", map[string]string{"error": getErr.Error()})
				return
			}

			// Broadcast final status
			if err != nil {
				h.wsHandler.BroadcastStatusUpdate(urlID, "error", updatedURL)
			} else {
				h.wsHandler.BroadcastStatusUpdate(urlID, "completed", updatedURL)
			}

			// Check if background context was cancelled
			select {
			case <-bgCtx.Done():
				h.wsHandler.BroadcastStatusUpdate(urlID, "timeout", map[string]string{"error": "Analysis timeout"})
			default:
			}
		}(id)
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Bulk analysis started",
		Data:    map[string]int{"count": len(req.IDs)},
	})
}

func (h *URLHandler) BulkDelete(c *gin.Context) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	var req models.BulkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	// Check if context is already cancelled
	select {
	case <-ctx.Done():
		c.JSON(http.StatusRequestTimeout, models.ErrorResponse{Error: "Request timeout"})
		return
	default:
	}

	deleted := 0
	for _, id := range req.IDs {
		// Check context before each deletion
		select {
		case <-ctx.Done():
			c.JSON(http.StatusRequestTimeout, models.ErrorResponse{Error: "Request timeout during bulk deletion"})
			return
		default:
		}

		if err := h.crawlerService.DeleteURL(id); err == nil {
			deleted++
			// Broadcast deletion
			h.wsHandler.BroadcastStatusUpdate(id, "deleted", nil)
		}
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Bulk deletion completed",
		Data:    map[string]int{"deleted": deleted, "total": len(req.IDs)},
	})
}

func (h *URLHandler) GetBrokenLinks(c *gin.Context) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	_, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid URL ID"})
		return
	}

	// Check if context is already cancelled
	select {
	case <-ctx.Done():
		c.JSON(http.StatusRequestTimeout, models.ErrorResponse{Error: "Request timeout"})
		return
	default:
	}

	c.JSON(http.StatusOK, []interface{}{})
}