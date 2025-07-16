package services

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"website-analyzer/internal/models"
	"website-analyzer/internal/repository"
	"website-analyzer/internal/worker"

	"golang.org/x/net/html"
)

// CrawlerService interface for URL crawling operations
type CrawlerService interface {
	AddURL(urlStr string) (*models.URL, error)
	GetURLs(page, limit int, search string) ([]models.URL, int, error)
	GetURL(id int) (*models.URL, error)
	AnalyzeURL(id int) error
	DeleteURL(id int) error
	// Enhanced methods with context
	AddURLWithContext(ctx context.Context, urlStr string) (*models.URL, error)
	GetURLsWithContext(ctx context.Context, filter repository.URLFilter) ([]models.URL, int, error)
	GetURLWithContext(ctx context.Context, id int) (*models.URL, error)
	AnalyzeURLWithContext(ctx context.Context, id int) error
	DeleteURLWithContext(ctx context.Context, id int) error
	AnalyzeURLs(ctx context.Context, ids []int) error
	DeleteURLs(ctx context.Context, ids []int) error
	GetBrokenLinks(ctx context.Context, urlID int) ([]models.BrokenLink, error)
}

// CrawlerConfig holds configuration for the crawler
type CrawlerConfig struct {
	MaxConcurrentCrawls int           `envconfig:"CRAWLER_MAX_CONCURRENT" default:"10"`
	RequestTimeout      time.Duration `envconfig:"CRAWLER_REQUEST_TIMEOUT" default:"30s"`
	UserAgent           string        `envconfig:"CRAWLER_USER_AGENT" default:"WebsiteAnalyzer/1.0"`
	MaxRedirects        int           `envconfig:"CRAWLER_MAX_REDIRECTS" default:"5"`
	MaxResponseSize     int64         `envconfig:"CRAWLER_MAX_RESPONSE_SIZE" default:"10485760"` // 10MB
	RetryAttempts       int           `envconfig:"CRAWLER_RETRY_ATTEMPTS" default:"3"`
	RetryDelay          time.Duration `envconfig:"CRAWLER_RETRY_DELAY" default:"1s"`
}

// enhancedCrawlerService implements CrawlerService
type enhancedCrawlerService struct {
	urlRepo    repository.URLRepository
	workerPool *worker.WorkerPool
	httpClient *http.Client
	logger     *slog.Logger
	config     *CrawlerConfig
}

// NewCrawlerService creates a new crawler service
func NewCrawlerService(db repository.URLRepository, workerPool *worker.WorkerPool, config *CrawlerConfig, logger *slog.Logger) CrawlerService {
	// Create HTTP client with optimized settings
	httpClient := &http.Client{
		Timeout: config.RequestTimeout,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
			DisableKeepAlives:   false,
			DisableCompression:  false,
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= config.MaxRedirects {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	service := &enhancedCrawlerService{
		urlRepo:    db,
		workerPool: workerPool,
		httpClient: httpClient,
		logger:     logger,
		config:     config,
	}

	// Register job handlers
	workerPool.RegisterHandler(worker.JobTypeAnalyzeURL, service.handleAnalyzeJob)
	workerPool.RegisterHandler(worker.JobTypeCrawlURL, service.handleCrawlJob)

	return service
}

// ===============================
// BACKWARD COMPATIBILITY METHODS
// ===============================

// AddURL adds a URL (backward compatibility method)
func (s *enhancedCrawlerService) AddURL(urlStr string) (*models.URL, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return s.addURL(ctx, urlStr)
}

// GetURLs retrieves URLs with pagination (backward compatibility method)
func (s *enhancedCrawlerService) GetURLs(page, limit int, search string) ([]models.URL, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := repository.URLFilter{
		Search: search,
		Page:   page,
		Limit:  limit,
	}

	return s.getURLs(ctx, filter)
}

// GetURL retrieves a single URL by ID (backward compatibility method)
func (s *enhancedCrawlerService) GetURL(id int) (*models.URL, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.getURL(ctx, id)
}

// AnalyzeURL analyzes a single URL (backward compatibility method)
func (s *enhancedCrawlerService) AnalyzeURL(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.analyzeURL(ctx, id)
}

// DeleteURL deletes a single URL (backward compatibility method)
func (s *enhancedCrawlerService) DeleteURL(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.deleteURL(ctx, id)
}

// ===============================
// ENHANCED CONTEXT METHODS
// ===============================

// AddURLWithContext adds a URL with context
func (s *enhancedCrawlerService) AddURLWithContext(ctx context.Context, urlStr string) (*models.URL, error) {
	return s.addURL(ctx, urlStr)
}

// GetURLsWithContext retrieves URLs with filtering and context
func (s *enhancedCrawlerService) GetURLsWithContext(ctx context.Context, filter repository.URLFilter) ([]models.URL, int, error) {
	return s.getURLs(ctx, filter)
}

// GetURLWithContext retrieves a URL by ID with context
func (s *enhancedCrawlerService) GetURLWithContext(ctx context.Context, id int) (*models.URL, error) {
	return s.getURL(ctx, id)
}

// AnalyzeURLWithContext analyzes a URL with context
func (s *enhancedCrawlerService) AnalyzeURLWithContext(ctx context.Context, id int) error {
	return s.analyzeURL(ctx, id)
}

// DeleteURLWithContext deletes a URL with context
func (s *enhancedCrawlerService) DeleteURLWithContext(ctx context.Context, id int) error {
	return s.deleteURL(ctx, id)
}

// AnalyzeURLs analyzes multiple URLs concurrently
func (s *enhancedCrawlerService) AnalyzeURLs(ctx context.Context, ids []int) error {
	if len(ids) == 0 {
		return nil
	}

	for _, id := range ids {
		if err := s.analyzeURL(ctx, id); err != nil {
			s.logger.Error("Failed to queue analysis job", slog.Int("url_id", id), slog.String("error", err.Error()))
		}
	}

	return nil
}

// DeleteURLs deletes multiple URLs
func (s *enhancedCrawlerService) DeleteURLs(ctx context.Context, ids []int) error {
	if len(ids) == 0 {
		return nil
	}

	if err := s.urlRepo.DeleteBatch(ctx, ids); err != nil {
		return fmt.Errorf("failed to delete URLs: %w", err)
	}

	s.logger.Info("URLs deleted successfully", slog.Int("count", len(ids)))
	return nil
}

// GetBrokenLinks retrieves broken links for a URL
func (s *enhancedCrawlerService) GetBrokenLinks(ctx context.Context, urlID int) ([]models.BrokenLink, error) {
	// TODO: Implement broken links repository
	return nil, fmt.Errorf("not implemented")
}

// ===============================
// INTERNAL IMPLEMENTATION METHODS
// ===============================

// addURL adds a URL to the system
func (s *enhancedCrawlerService) addURL(ctx context.Context, urlStr string) (*models.URL, error) {
	// Validate URL
	if err := s.validateURL(urlStr); err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Generate URL hash
	urlHash := models.GenerateURLHash(urlStr)

	// Check if URL already exists
	if existingURL, err := s.urlRepo.FindByHash(ctx, urlHash); err == nil && existingURL != nil {
		s.logger.Info("URL already exists", slog.String("url", urlStr), slog.Int("id", existingURL.ID))
		return existingURL, nil
	}

	// Create new URL
	newURL := &models.URL{
		URL:     urlStr,
		URLHash: urlHash,
		Status:  models.StatusQueued,
	}

	if err := s.urlRepo.Save(ctx, newURL); err != nil {
		return nil, fmt.Errorf("failed to save URL: %w", err)
	}

	s.logger.Info("URL added successfully", slog.String("url", urlStr), slog.Int("id", newURL.ID))
	return newURL, nil
}

// getURLs retrieves URLs with pagination and filtering
func (s *enhancedCrawlerService) getURLs(ctx context.Context, filter repository.URLFilter) ([]models.URL, int, error) {
	// Validate filter parameters
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 || filter.Limit > 100 {
		filter.Limit = 10
	}

	urls, total, err := s.urlRepo.FindAll(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to retrieve URLs: %w", err)
	}

	return urls, total, nil
}

// getURL retrieves a single URL by ID
func (s *enhancedCrawlerService) getURL(ctx context.Context, id int) (*models.URL, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid URL ID: %d", id)
	}

	url, err := s.urlRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve URL: %w", err)
	}

	return url, nil
}

// analyzeURL analyzes a single URL
func (s *enhancedCrawlerService) analyzeURL(ctx context.Context, id int) error {
	// Create analysis job
	job := worker.Job{
		ID:        fmt.Sprintf("analyze_%d_%d", id, time.Now().Unix()),
		Type:      worker.JobTypeAnalyzeURL,
		Payload:   id,
		MaxRetry:  s.config.RetryAttempts,
		CreatedAt: time.Now(),
	}

	if err := s.workerPool.AddJob(job); err != nil {
		return fmt.Errorf("failed to queue analysis job: %w", err)
	}

	return nil
}

// deleteURL deletes a single URL
func (s *enhancedCrawlerService) deleteURL(ctx context.Context, id int) error {
	if id <= 0 {
		return fmt.Errorf("invalid URL ID: %d", id)
	}

	if err := s.urlRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete URL: %w", err)
	}

	s.logger.Info("URL deleted successfully", slog.Int("id", id))
	return nil
}

// ===============================
// JOB HANDLERS
// ===============================

// handleAnalyzeJob handles URL analysis jobs
func (s *enhancedCrawlerService) handleAnalyzeJob(ctx context.Context, job worker.Job) (interface{}, error) {
	urlID, ok := job.Payload.(int)
	if !ok {
		return nil, fmt.Errorf("invalid job payload: expected int, got %T", job.Payload)
	}

	s.logger.Info("Starting URL analysis", slog.Int("url_id", urlID))

	// Update status to processing
	url, err := s.urlRepo.FindByID(ctx, urlID)
	if err != nil {
		return nil, fmt.Errorf("failed to find URL: %w", err)
	}

	url.Status = models.StatusProcessing
	if err := s.urlRepo.Update(ctx, url); err != nil {
		return nil, fmt.Errorf("failed to update URL status: %w", err)
	}

	// Perform analysis
	result, err := s.crawlURL(ctx, url.URL)
	if err != nil {
		// Update status to error
		url.Status = models.StatusError
		errMsg := err.Error()
		url.ErrorMessage = &errMsg
		s.urlRepo.Update(ctx, url)
		return nil, fmt.Errorf("failed to crawl URL: %w", err)
	}

	// Update URL with analysis results
	url.Title = &result.Title
	url.HTMLVersion = &result.HTMLVersion
	url.H1Count = result.HeadingCounts.H1
	url.H2Count = result.HeadingCounts.H2
	url.H3Count = result.HeadingCounts.H3
	url.H4Count = result.HeadingCounts.H4
	url.H5Count = result.HeadingCounts.H5
	url.H6Count = result.HeadingCounts.H6
	url.InternalLinksCount = result.InternalLinksCount
	url.ExternalLinksCount = result.ExternalLinksCount
	url.BrokenLinksCount = result.BrokenLinksCount
	url.HasLoginForm = result.HasLoginForm
	url.Status = models.StatusCompleted
	url.ErrorMessage = nil

	if err := s.urlRepo.Update(ctx, url); err != nil {
		return nil, fmt.Errorf("failed to update URL with results: %w", err)
	}

	s.logger.Info("URL analysis completed", slog.Int("url_id", urlID))
	return url, nil
}

// handleCrawlJob handles URL crawling jobs
func (s *enhancedCrawlerService) handleCrawlJob(ctx context.Context, job worker.Job) (interface{}, error) {
	// Implementation for dedicated crawl jobs
	return nil, fmt.Errorf("crawl job handler not implemented")
}

// ===============================
// CRAWLING LOGIC
// ===============================

// crawlURL performs the actual URL crawling and analysis
func (s *enhancedCrawlerService) crawlURL(ctx context.Context, urlStr string) (*models.URLAnalysisResult, error) {
	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set user agent
	req.Header.Set("User-Agent", s.config.UserAgent)

	// Perform request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

		// Check response status
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("HTTP error: %d %s", resp.StatusCode, resp.Status)
		}
	
		// Limit response size
		limitedReader := &io.LimitedReader{R: resp.Body, N: s.config.MaxResponseSize}
	
		// Parse HTML
		doc, err := html.Parse(limitedReader)
		if err != nil {
			return nil, fmt.Errorf("failed to parse HTML: %w", err)
		}
	
		// Analyze document
		result := &models.URLAnalysisResult{
			HTMLVersion: s.detectHTMLVersion(doc),
		}
	
		baseURL, _ := url.Parse(urlStr)
		s.analyzeHTMLNode(doc, result, baseURL.String())
	
		return result, nil
	}
	
	// validateURL validates a URL string
	func (s *enhancedCrawlerService) validateURL(urlStr string) error {
		// Basic validation
		if strings.TrimSpace(urlStr) == "" {
			return fmt.Errorf("URL cannot be empty")
		}
	
		// Parse URL
		parsedURL, err := url.ParseRequestURI(urlStr)
		if err != nil {
			return fmt.Errorf("invalid URL format: %w", err)
		}
	
		// Scheme validation
		if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
			return fmt.Errorf("URL must use HTTP or HTTPS protocol")
		}
	
		// Host validation
		if parsedURL.Host == "" {
			return fmt.Errorf("URL must contain a valid host")
		}
	
		return nil
	}
	
	// detectHTMLVersion detects HTML version from document
	func (s *enhancedCrawlerService) detectHTMLVersion(doc *html.Node) string {
		// Look for DOCTYPE
		var findDoctype func(*html.Node) string
		findDoctype = func(n *html.Node) string {
			if n.Type == html.DoctypeNode {
				if strings.Contains(strings.ToLower(n.Data), "html") {
					if strings.Contains(n.Data, "XHTML") {
						return "XHTML"
					}
					if strings.Contains(n.Data, "HTML 4") {
						return "HTML 4.01"
					}
					return "HTML5"
				}
			}
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if result := findDoctype(c); result != "" {
					return result
				}
			}
			return ""
		}
	
		if version := findDoctype(doc); version != "" {
			return version
		}
		return "HTML5" // Default assumption
	}
	
	// analyzeHTMLNode recursively analyzes HTML nodes
	func (s *enhancedCrawlerService) analyzeHTMLNode(n *html.Node, result *models.URLAnalysisResult, baseURL string) {
		if n.Type == html.ElementNode {
			switch strings.ToLower(n.Data) {
			case "title":
				if n.FirstChild != nil && n.FirstChild.Type == html.TextNode {
					result.Title = strings.TrimSpace(n.FirstChild.Data)
				}
			case "h1":
				result.HeadingCounts.H1++
			case "h2":
				result.HeadingCounts.H2++
			case "h3":
				result.HeadingCounts.H3++
			case "h4":
				result.HeadingCounts.H4++
			case "h5":
				result.HeadingCounts.H5++
			case "h6":
				result.HeadingCounts.H6++
			case "a":
				s.analyzeLink(n, result, baseURL)
			case "input":
				s.analyzeInput(n, result)
			case "form":
				s.analyzeForm(n, result)
			}
		}
	
		// Recursively analyze child nodes
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			s.analyzeHTMLNode(c, result, baseURL)
		}
	}
	
	// analyzeLink analyzes link elements
	func (s *enhancedCrawlerService) analyzeLink(n *html.Node, result *models.URLAnalysisResult, baseURL string) {
		for _, attr := range n.Attr {
			if attr.Key == "href" && attr.Val != "" {
				linkURL := strings.TrimSpace(attr.Val)
	
				// Skip anchors, javascript, and mailto links
				if strings.HasPrefix(linkURL, "#") ||
					strings.HasPrefix(linkURL, "javascript:") ||
					strings.HasPrefix(linkURL, "mailto:") {
					continue
				}
	
				// Parse link URL
				parsedLink, err := url.Parse(linkURL)
				if err != nil {
					continue
				}
	
				// Parse base URL
				parsedBase, err := url.Parse(baseURL)
				if err != nil {
					continue
				}
	
				// Resolve relative URLs
				resolvedURL := parsedBase.ResolveReference(parsedLink)
	
				// Classify as internal or external
				if resolvedURL.Host == parsedBase.Host {
					result.InternalLinksCount++
				} else {
					result.ExternalLinksCount++
				}
	
				break
			}
		}
	}
	
	// analyzeInput analyzes input elements for login forms
	func (s *enhancedCrawlerService) analyzeInput(n *html.Node, result *models.URLAnalysisResult) {
		for _, attr := range n.Attr {
			if attr.Key == "type" && (attr.Val == "password" || attr.Val == "email") {
				result.HasLoginForm = true
				break
			}
		}
	}
	
	// analyzeForm analyzes form elements
	func (s *enhancedCrawlerService) analyzeForm(n *html.Node, result *models.URLAnalysisResult) {
		// Check for login-related attributes
		for _, attr := range n.Attr {
			if attr.Key == "id" || attr.Key == "class" || attr.Key == "name" {
				value := strings.ToLower(attr.Val)
				loginPatterns := []string{"login", "signin", "sign-in", "auth", "user"}
				for _, pattern := range loginPatterns {
					if strings.Contains(value, pattern) {
						result.HasLoginForm = true
						return
					}
				}
			}
		}
	}