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

	"searcher-app/internal/models"
	"searcher-app/internal/repository"
	"searcher-app/internal/worker"

	"golang.org/x/net/html"
)

type CrawlerService interface {
	AddURL(urlStr string) (*models.URL, error)
	GetURLs(page, limit int, search string) ([]models.URL, int, error)
	GetURL(id int) (*models.URL, error)
	AnalyzeURL(id int) error
	DeleteURL(id int) error
	AddURLWithContext(ctx context.Context, urlStr string) (*models.URL, error)
	GetURLsWithContext(ctx context.Context, filter repository.URLFilter) ([]models.URL, int, error)
	GetURLWithContext(ctx context.Context, id int) (*models.URL, error)
	AnalyzeURLWithContext(ctx context.Context, id int) error
	DeleteURLWithContext(ctx context.Context, id int) error
	AnalyzeURLs(ctx context.Context, ids []int) error
	DeleteURLs(ctx context.Context, ids []int) error
	GetBrokenLinks(ctx context.Context, urlID int) ([]models.BrokenLink, error)
}

type CrawlerConfig struct {
	MaxConcurrentCrawls int           `envconfig:"CRAWLER_MAX_CONCURRENT" default:"10"`
	RequestTimeout      time.Duration `envconfig:"CRAWLER_REQUEST_TIMEOUT" default:"30s"`
	UserAgent           string        `envconfig:"CRAWLER_USER_AGENT" default:"WebsiteAnalyzer/1.0"`
	MaxRedirects        int           `envconfig:"CRAWLER_MAX_REDIRECTS" default:"5"`
	MaxResponseSize     int64         `envconfig:"CRAWLER_MAX_RESPONSE_SIZE" default:"10485760"`
	RetryAttempts       int           `envconfig:"CRAWLER_RETRY_ATTEMPTS" default:"3"`
	RetryDelay          time.Duration `envconfig:"CRAWLER_RETRY_DELAY" default:"1s"`
}

type enhancedCrawlerService struct {
	urlRepo    repository.URLRepository
	workerPool *worker.WorkerPool
	httpClient *http.Client
	logger     *slog.Logger
	config     *CrawlerConfig
}

func NewCrawlerService(db repository.URLRepository, workerPool *worker.WorkerPool, config *CrawlerConfig, logger *slog.Logger) CrawlerService {
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

	workerPool.RegisterHandler(worker.JobTypeAnalyzeURL, service.handleAnalyzeJob)
	workerPool.RegisterHandler(worker.JobTypeCrawlURL, service.handleCrawlJob)

	return service
}


func (s *enhancedCrawlerService) AddURL(urlStr string) (*models.URL, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return s.addURL(ctx, urlStr)
}

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

func (s *enhancedCrawlerService) GetURL(id int) (*models.URL, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.getURL(ctx, id)
}

func (s *enhancedCrawlerService) AnalyzeURL(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.analyzeURL(ctx, id)
}

func (s *enhancedCrawlerService) DeleteURL(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.deleteURL(ctx, id)
}


func (s *enhancedCrawlerService) AddURLWithContext(ctx context.Context, urlStr string) (*models.URL, error) {
	return s.addURL(ctx, urlStr)
}

func (s *enhancedCrawlerService) GetURLsWithContext(ctx context.Context, filter repository.URLFilter) ([]models.URL, int, error) {
	return s.getURLs(ctx, filter)
}

func (s *enhancedCrawlerService) GetURLWithContext(ctx context.Context, id int) (*models.URL, error) {
	return s.getURL(ctx, id)
}

func (s *enhancedCrawlerService) AnalyzeURLWithContext(ctx context.Context, id int) error {
	return s.analyzeURL(ctx, id)
}

func (s *enhancedCrawlerService) DeleteURLWithContext(ctx context.Context, id int) error {
	return s.deleteURL(ctx, id)
}

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

func (s *enhancedCrawlerService) GetBrokenLinks(ctx context.Context, urlID int) ([]models.BrokenLink, error) {
	if urlID <= 0 {
		return nil, fmt.Errorf("invalid URL ID: %d", urlID)
	}

	brokenLinks, err := s.urlRepo.FindBrokenLinksByURLID(ctx, urlID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve broken links: %w", err)
	}

	return brokenLinks, nil
}


func (s *enhancedCrawlerService) addURL(ctx context.Context, urlStr string) (*models.URL, error) {
	if err := s.validateURL(urlStr); err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	urlHash := models.GenerateURLHash(urlStr)

	if existingURL, err := s.urlRepo.FindByHash(ctx, urlHash); err == nil && existingURL != nil {
		s.logger.Info("URL already exists", slog.String("url", urlStr), slog.Int("id", existingURL.ID))
		return existingURL, nil
	}

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

func (s *enhancedCrawlerService) getURLs(ctx context.Context, filter repository.URLFilter) ([]models.URL, int, error) {
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

func (s *enhancedCrawlerService) analyzeURL(ctx context.Context, id int) error {
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


func (s *enhancedCrawlerService) handleAnalyzeJob(ctx context.Context, job worker.Job) (interface{}, error) {
	urlID, ok := job.Payload.(int)
	if !ok {
		return nil, fmt.Errorf("invalid job payload: expected int, got %T", job.Payload)
	}

	s.logger.Info("Starting URL analysis", slog.Int("url_id", urlID))

	url, err := s.urlRepo.FindByID(ctx, urlID)
	if err != nil {
		return nil, fmt.Errorf("failed to find URL: %w", err)
	}

	url.Status = models.StatusProcessing
	if err := s.urlRepo.Update(ctx, url); err != nil {
		return nil, fmt.Errorf("failed to update URL status: %w", err)
	}

	result, err := s.crawlURL(ctx, url.URL)
	if err != nil {
		url.Status = models.StatusError
		errMsg := err.Error()
		url.ErrorMessage = &errMsg
		s.urlRepo.Update(ctx, url)
		return nil, fmt.Errorf("failed to crawl URL: %w", err)
	}

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

	if len(result.BrokenLinks) > 0 {
		if err := s.urlRepo.DeleteBrokenLinksByURLID(ctx, urlID); err != nil {
			s.logger.Error("Failed to clear existing broken links", slog.Int("url_id", urlID), slog.String("error", err.Error()))
		}

		for _, brokenLink := range result.BrokenLinks {
			brokenLink.URLID = urlID
			if err := s.urlRepo.SaveBrokenLink(ctx, &brokenLink); err != nil {
				s.logger.Error("Failed to save broken link", slog.Int("url_id", urlID), slog.String("link", brokenLink.LinkURL), slog.String("error", err.Error()))
			}
		}
	}

	s.logger.Info("URL analysis completed", slog.Int("url_id", urlID))
	return url, nil
}

func (s *enhancedCrawlerService) handleCrawlJob(ctx context.Context, job worker.Job) (interface{}, error) {
	return nil, fmt.Errorf("crawl job handler not implemented")
}


func (s *enhancedCrawlerService) crawlURL(ctx context.Context, urlStr string) (*models.URLAnalysisResult, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", s.config.UserAgent)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d %s", resp.StatusCode, resp.Status)
	}

	limitedReader := &io.LimitedReader{R: resp.Body, N: s.config.MaxResponseSize}

	doc, err := html.Parse(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	result := &models.URLAnalysisResult{
		HTMLVersion:   s.detectHTMLVersion(doc),
		HeadingCounts: models.HeadingCounts{},
		BrokenLinks:   []models.BrokenLink{},
	}

	baseURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse base URL: %w", err)
	}

	links := s.collectLinks(doc, baseURL)

	s.analyzeHTMLNode(doc, result, baseURL)

	s.analyzeLinks(ctx, links, result, baseURL)

	return result, nil
}

func (s *enhancedCrawlerService) validateURL(urlStr string) error {
	if strings.TrimSpace(urlStr) == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	parsedURL, err := url.ParseRequestURI(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("URL must use HTTP or HTTPS protocol")
	}

	if parsedURL.Host == "" {
		return fmt.Errorf("URL must contain a valid host")
	}

	return nil
}

func (s *enhancedCrawlerService) detectHTMLVersion(doc *html.Node) string {
	var findDoctype func(*html.Node) string
	findDoctype = func(n *html.Node) string {
		if n.Type == html.DoctypeNode {
			doctype := strings.ToLower(n.Data)
			
			if doctype == "html" {
				return "HTML5"
			}
			
			if strings.Contains(doctype, "html 4.01") && strings.Contains(doctype, "strict") {
				return "HTML 4.01 Strict"
			}
			
			if strings.Contains(doctype, "html 4.01") && strings.Contains(doctype, "transitional") {
				return "HTML 4.01 Transitional"
			}
			
			if strings.Contains(doctype, "html 4.01") && strings.Contains(doctype, "frameset") {
				return "HTML 4.01 Frameset"
			}
			
			if strings.Contains(doctype, "xhtml 1.0") && strings.Contains(doctype, "strict") {
				return "XHTML 1.0 Strict"
			}
			
			if strings.Contains(doctype, "xhtml 1.0") && strings.Contains(doctype, "transitional") {
				return "XHTML 1.0 Transitional"
			}
			
			if strings.Contains(doctype, "xhtml 1.0") && strings.Contains(doctype, "frameset") {
				return "XHTML 1.0 Frameset"
			}
			
			if strings.Contains(doctype, "xhtml 1.1") {
				return "XHTML 1.1"
			}
			
			if strings.Contains(doctype, "xhtml") {
				return "XHTML"
			}
			if strings.Contains(doctype, "html") {
				return "HTML 4.01"
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
	
	if s.hasXMLDeclaration(doc) {
		return "XHTML"
	}
	
	return "HTML5"
}

func (s *enhancedCrawlerService) hasXMLDeclaration(doc *html.Node) bool {
	return false
}

func (s *enhancedCrawlerService) analyzeHTMLNode(n *html.Node, result *models.URLAnalysisResult, baseURL *url.URL) {
	if n.Type == html.ElementNode {
		switch strings.ToLower(n.Data) {
		case "title":
			if result.Title == "" {
				result.Title = s.extractTextContent(n)
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
		case "form":
			if s.isLoginForm(n) {
				result.HasLoginForm = true
			}
		case "input":
			if s.isLoginInput(n) {
				result.HasLoginForm = true
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		s.analyzeHTMLNode(c, result, baseURL)
	}
}

func (s *enhancedCrawlerService) extractTextContent(n *html.Node) string {
	var text strings.Builder
	var extract func(*html.Node)
	extract = func(node *html.Node) {
		if node.Type == html.TextNode {
			text.WriteString(node.Data)
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			extract(c)
		}
	}
	extract(n)
	return strings.TrimSpace(text.String())
}

func (s *enhancedCrawlerService) collectLinks(doc *html.Node, baseURL *url.URL) []string {
	var links []string
	var collect func(*html.Node)
	collect = func(n *html.Node) {
		if n.Type == html.ElementNode && strings.ToLower(n.Data) == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" && attr.Val != "" {
					href := strings.TrimSpace(attr.Val)
					if href != "" && !strings.HasPrefix(href, "#") && 
					   !strings.HasPrefix(href, "javascript:") && 
					   !strings.HasPrefix(href, "mailto:") &&
					   !strings.HasPrefix(href, "tel:") {
						links = append(links, href)
					}
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			collect(c)
		}
	}
	collect(doc)
	return links
}

func (s *enhancedCrawlerService) analyzeLinks(ctx context.Context, links []string, result *models.URLAnalysisResult, baseURL *url.URL) {
	uniqueLinks := make(map[string]bool)
	
	for _, link := range links {
		parsedLink, err := url.Parse(link)
		if err != nil {
			continue
		}
		
		resolvedURL := baseURL.ResolveReference(parsedLink)
		resolvedStr := resolvedURL.String()
		
		if uniqueLinks[resolvedStr] {
			continue
		}
		uniqueLinks[resolvedStr] = true
		
		if resolvedURL.Host == baseURL.Host {
			result.InternalLinksCount++
		} else {
			result.ExternalLinksCount++
		}
		
		if statusCode, err := s.checkLinkStatus(ctx, resolvedStr); err != nil || statusCode >= 400 {
			result.BrokenLinksCount++
			brokenLink := models.BrokenLink{
				LinkURL:    resolvedStr,
				StatusCode: statusCode,
			}
			if err != nil {
				errMsg := err.Error()
				brokenLink.ErrorMessage = &errMsg
			}
			result.BrokenLinks = append(result.BrokenLinks, brokenLink)
		}
	}
}

func (s *enhancedCrawlerService) checkLinkStatus(ctx context.Context, linkURL string) (int, error) {
	linkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(linkCtx, "HEAD", linkURL, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", s.config.UserAgent)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		getCtx, getCancel := context.WithTimeout(ctx, 3*time.Second)
		defer getCancel()
		
		getReq, getErr := http.NewRequestWithContext(getCtx, "GET", linkURL, nil)
		if getErr != nil {
			return 0, fmt.Errorf("failed to create GET request: %w", getErr)
		}
		
		getReq.Header.Set("User-Agent", s.config.UserAgent)
		getResp, getErr := s.httpClient.Do(getReq)
		if getErr != nil {
			return 0, fmt.Errorf("failed to fetch URL: %w", err)
		}
		defer getResp.Body.Close()
		
		return getResp.StatusCode, nil
	}
	defer resp.Body.Close()

	return resp.StatusCode, nil
}

func (s *enhancedCrawlerService) isLoginForm(n *html.Node) bool {
	for _, attr := range n.Attr {
		if attr.Key == "id" || attr.Key == "class" || attr.Key == "name" {
			value := strings.ToLower(attr.Val)
			loginPatterns := []string{
				"login", "signin", "sign-in", "auth", "authentication",
				"user", "account", "credential", "password", "login-form",
				"signin-form", "auth-form", "user-form",
			}
			for _, pattern := range loginPatterns {
				if strings.Contains(value, pattern) {
					return true
				}
			}
		}
	}
	
	hasPasswordInput := false
	hasUsernameInput := false
	
	var checkInputs func(*html.Node)
	checkInputs = func(node *html.Node) {
		if node.Type == html.ElementNode && strings.ToLower(node.Data) == "input" {
			inputType := ""
			inputName := ""
			inputId := ""
			
			for _, attr := range node.Attr {
				switch attr.Key {
				case "type":
					inputType = strings.ToLower(attr.Val)
				case "name":
					inputName = strings.ToLower(attr.Val)
				case "id":
					inputId = strings.ToLower(attr.Val)
				}
			}
			
			if inputType == "password" {
				hasPasswordInput = true
			}
			
			usernamePatterns := []string{
				"username", "user", "email", "login", "account",
				"userid", "user_id", "user-id", "mail",
			}
			for _, pattern := range usernamePatterns {
				if strings.Contains(inputName, pattern) || strings.Contains(inputId, pattern) {
					hasUsernameInput = true
					break
				}
			}
			
			if inputType == "email" {
				hasUsernameInput = true
			}
		}
		
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			checkInputs(c)
		}
	}
	
	checkInputs(n)
	
	return hasPasswordInput && hasUsernameInput
}

func (s *enhancedCrawlerService) isLoginInput(n *html.Node) bool {
	inputType := ""
	inputName := ""
	inputId := ""
	
	for _, attr := range n.Attr {
		switch attr.Key {
		case "type":
			inputType = strings.ToLower(attr.Val)
		case "name":
			inputName = strings.ToLower(attr.Val)
		case "id":
			inputId = strings.ToLower(attr.Val)
		}
	}
	
	if inputType == "password" {
		return true
	}
	
	if inputType == "email" {
		return true
	}
	
	loginPatterns := []string{
		"password", "username", "user", "email", "login",
		"signin", "auth", "credential", "account",
	}
	
	for _, pattern := range loginPatterns {
		if strings.Contains(inputName, pattern) || strings.Contains(inputId, pattern) {
			return true
		}
	}
	
	return false
}
