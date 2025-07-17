package middleware

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type ValidationConfig struct {
	MaxURLLength    int
	MaxSearchLength int
	MaxPageValue    int
	MaxLimitValue   int
	AllowedDomains  []string
	BlockedPatterns []string
	RequireHTTPS    bool
	ValidateReferer bool
}

func DefaultValidationConfig() *ValidationConfig {
	return &ValidationConfig{
		MaxURLLength:    2048,
		MaxSearchLength: 255,
		MaxPageValue:    10000,
		MaxLimitValue:   100,
		AllowedDomains:  []string{},
		BlockedPatterns: []string{
			`(?i)(union|select|insert|update|delete|drop|create|alter|exec|execute)`,
			`(?i)(script|javascript|vbscript|onload|onerror|onclick)`,
			`(?i)(\<|\>|%3c|%3e|&lt;|&gt;)`,
			`(?i)(cmd|command|exec|system|shell|bash|sh|powershell)`,
			`(?i)(\||&|;|` + "`" + `)`,
			`(?i)(\.\.\/|\.\.\\|\.\.\%2f|\.\.\%5c)`,
			`(?i)(alert|confirm|prompt|document\.cookie|window\.location)`,
		},
		RequireHTTPS:    false,
		ValidateReferer: false,
	}
}

type InputValidator struct {
	config         *ValidationConfig
	blockedRegexes []*regexp.Regexp
}

func NewInputValidator(config *ValidationConfig) *InputValidator {
	if config == nil {
		config = DefaultValidationConfig()
	}

	validator := &InputValidator{
		config:         config,
		blockedRegexes: make([]*regexp.Regexp, 0, len(config.BlockedPatterns)),
	}

	for _, pattern := range config.BlockedPatterns {
		if regex, err := regexp.Compile(pattern); err == nil {
			validator.blockedRegexes = append(validator.blockedRegexes, regex)
		}
	}

	return validator
}

func (v *InputValidator) ValidateURL(urlStr string) error {
	if len(urlStr) > v.config.MaxURLLength {
		return fmt.Errorf("URL length exceeds maximum of %d characters", v.config.MaxURLLength)
	}

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

	if v.config.RequireHTTPS && parsedURL.Scheme != "https" {
		return fmt.Errorf("URL must use HTTPS protocol")
	}

	if parsedURL.Host == "" {
		return fmt.Errorf("URL must contain a valid host")
	}

	if len(v.config.AllowedDomains) > 0 {
		allowed := false
		for _, domain := range v.config.AllowedDomains {
			if strings.Contains(parsedURL.Host, domain) {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("domain not allowed: %s", parsedURL.Host)
		}
	}

	for _, regex := range v.blockedRegexes {
		if regex.MatchString(urlStr) {
			return fmt.Errorf("URL contains blocked pattern")
		}
	}

	return nil
}

func (v *InputValidator) ValidateSearchQuery(query string) error {
	if len(query) > v.config.MaxSearchLength {
		return fmt.Errorf("search query length exceeds maximum of %d characters", v.config.MaxSearchLength)
	}

	for _, regex := range v.blockedRegexes {
		if regex.MatchString(query) {
			return fmt.Errorf("search query contains blocked pattern")
		}
	}

	return nil
}

func (v *InputValidator) ValidateIntegerParam(value string, paramName string, min, max int) (int, error) {
	if value == "" {
		return 0, fmt.Errorf("%s cannot be empty", paramName)
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: must be an integer", paramName)
	}

	if intValue < min {
		return 0, fmt.Errorf("%s must be at least %d", paramName, min)
	}

	if intValue > max {
		return 0, fmt.Errorf("%s must be at most %d", paramName, max)
	}

	return intValue, nil
}

func (v *InputValidator) ValidateIDParam(value string) (int, error) {
	return v.ValidateIntegerParam(value, "ID", 1, 2147483647)
}

func ValidationMiddleware() gin.HandlerFunc {
	validator := NewInputValidator(DefaultValidationConfig())

	return func(c *gin.Context) {
		path := c.Request.URL.Path
		method := c.Request.Method

		if method == "POST" && strings.Contains(path, "/urls") {
			var urlReq struct {
				URL string `json:"url"`
			}
			if err := c.ShouldBindJSON(&urlReq); err == nil {
				if err := validator.ValidateURL(urlReq.URL); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"error":   "Invalid URL",
						"message": err.Error(),
					})
					c.Abort()
					return
				}
			}
		}

		if method == "GET" && strings.Contains(path, "/urls") {
			if search := c.Query("search"); search != "" {
				if err := validator.ValidateSearchQuery(search); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"error":   "Invalid search query",
						"message": err.Error(),
					})
					c.Abort()
					return
				}
			}

			if page := c.Query("page"); page != "" {
				if _, err := validator.ValidateIntegerParam(page, "page", 1, validator.config.MaxPageValue); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"error":   "Invalid page parameter",
						"message": err.Error(),
					})
					c.Abort()
					return
				}
			}

			if limit := c.Query("limit"); limit != "" {
				if _, err := validator.ValidateIntegerParam(limit, "limit", 1, validator.config.MaxLimitValue); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"error":   "Invalid limit parameter",
						"message": err.Error(),
					})
					c.Abort()
					return
				}
			}
		}

		if id := c.Param("id"); id != "" {
			if _, err := validator.ValidateIDParam(id); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "Invalid ID parameter",
					"message": err.Error(),
				})
				c.Abort()
				return
			}
		}

		if method == "POST" && strings.Contains(path, "/bulk") {
			var bulkReq struct {
				IDs []int `json:"ids"`
			}
			if err := c.ShouldBindJSON(&bulkReq); err == nil {
				if len(bulkReq.IDs) == 0 {
					c.JSON(http.StatusBadRequest, gin.H{
						"error":   "Invalid bulk request",
						"message": "IDs array cannot be empty",
					})
					c.Abort()
					return
				}
				if len(bulkReq.IDs) > 100 {
					c.JSON(http.StatusBadRequest, gin.H{
						"error":   "Invalid bulk request",
						"message": "Cannot process more than 100 IDs at once",
					})
					c.Abort()
					return
				}
				for _, id := range bulkReq.IDs {
					if id <= 0 {
						c.JSON(http.StatusBadRequest, gin.H{
							"error":   "Invalid bulk request",
							"message": "All IDs must be positive integers",
						})
						c.Abort()
						return
					}
				}
			}
		}

		c.Next()
	}
}

func StrictValidationMiddleware() gin.HandlerFunc {
	config := DefaultValidationConfig()
	config.RequireHTTPS = true
	config.ValidateReferer = true
	config.MaxURLLength = 1000
	config.MaxSearchLength = 100

	return func(c *gin.Context) {
		if config.ValidateReferer {
			referer := c.GetHeader("Referer")
			if referer == "" {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "Missing referer header",
					"message": "Referer header is required for this operation",
				})
				c.Abort()
				return
			}
		}

		ValidationMiddleware()(c)
	}
}
