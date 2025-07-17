package handlers

import (
	"context"
	"math"
	"net/http"
	"strconv"
	"time"

	"searcher-app/internal/models"
	"searcher-app/internal/services"

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
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	var req models.URLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

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

	go func() {
		bgCtx, bgCancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer bgCancel()

		h.wsHandler.BroadcastStatusUpdate(url.ID, "processing", nil)

		err := h.crawlerService.AnalyzeURL(url.ID)

		updatedURL, getErr := h.crawlerService.GetURL(url.ID)
		if getErr != nil {
			h.wsHandler.BroadcastStatusUpdate(url.ID, "error", map[string]string{"error": getErr.Error()})
			return
		}

		if err != nil {
			h.wsHandler.BroadcastStatusUpdate(url.ID, "error", updatedURL)
		} else {
			h.wsHandler.BroadcastStatusUpdate(url.ID, "completed", updatedURL)
		}

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
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid URL ID"})
		return
	}

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
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid URL ID"})
		return
	}

	select {
	case <-ctx.Done():
		c.JSON(http.StatusRequestTimeout, models.ErrorResponse{Error: "Request timeout"})
		return
	default:
	}

	go func() {
		bgCtx, bgCancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer bgCancel()

		h.wsHandler.BroadcastStatusUpdate(id, "processing", nil)

		err := h.crawlerService.AnalyzeURL(id)

		updatedURL, getErr := h.crawlerService.GetURL(id)
		if getErr != nil {
			h.wsHandler.BroadcastStatusUpdate(id, "error", map[string]string{"error": getErr.Error()})
			return
		}

		if err != nil {
			h.wsHandler.BroadcastStatusUpdate(id, "error", updatedURL)
		} else {
			h.wsHandler.BroadcastStatusUpdate(id, "completed", updatedURL)
		}

		select {
		case <-bgCtx.Done():
			h.wsHandler.BroadcastStatusUpdate(id, "timeout", map[string]string{"error": "Analysis timeout"})
		default:
		}
	}()

	c.JSON(http.StatusOK, models.SuccessResponse{Message: "URL analysis started"})
}

func (h *URLHandler) DeleteURL(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid URL ID"})
		return
	}

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

	h.wsHandler.BroadcastStatusUpdate(id, "deleted", nil)

	c.JSON(http.StatusOK, models.SuccessResponse{Message: "URL deleted successfully"})
}

func (h *URLHandler) BulkAnalyze(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var req models.BulkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	select {
	case <-ctx.Done():
		c.JSON(http.StatusRequestTimeout, models.ErrorResponse{Error: "Request timeout"})
		return
	default:
	}

	for _, id := range req.IDs {
		go func(urlID int) {
			bgCtx, bgCancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer bgCancel()

			h.wsHandler.BroadcastStatusUpdate(urlID, "processing", nil)

			err := h.crawlerService.AnalyzeURL(urlID)

			updatedURL, getErr := h.crawlerService.GetURL(urlID)
			if getErr != nil {
				h.wsHandler.BroadcastStatusUpdate(urlID, "error", map[string]string{"error": getErr.Error()})
				return
			}

			if err != nil {
				h.wsHandler.BroadcastStatusUpdate(urlID, "error", updatedURL)
			} else {
				h.wsHandler.BroadcastStatusUpdate(urlID, "completed", updatedURL)
			}

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
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	var req models.BulkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	select {
	case <-ctx.Done():
		c.JSON(http.StatusRequestTimeout, models.ErrorResponse{Error: "Request timeout"})
		return
	default:
	}

	deleted := 0
	for _, id := range req.IDs {
		select {
		case <-ctx.Done():
			c.JSON(http.StatusRequestTimeout, models.ErrorResponse{Error: "Request timeout during bulk deletion"})
			return
		default:
		}

		if err := h.crawlerService.DeleteURL(id); err == nil {
			deleted++
			h.wsHandler.BroadcastStatusUpdate(id, "deleted", nil)
		}
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Bulk deletion completed",
		Data:    map[string]int{"deleted": deleted, "total": len(req.IDs)},
	})
}

func (h *URLHandler) GetBrokenLinks(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid URL ID"})
		return
	}

	select {
	case <-ctx.Done():
		c.JSON(http.StatusRequestTimeout, models.ErrorResponse{Error: "Request timeout"})
		return
	default:
	}

	brokenLinks, err := h.crawlerService.GetBrokenLinks(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, brokenLinks)
}
