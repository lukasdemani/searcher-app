package models

import (
	"crypto/sha256"
	"fmt"
	"time"
)

type URLStatus string

const (
	StatusQueued     URLStatus = "queued"
	StatusProcessing URLStatus = "processing"
	StatusCompleted  URLStatus = "completed"
	StatusError      URLStatus = "error"
)

type URL struct {
	ID                  int       `json:"id" db:"id"`
	URL                 string    `json:"url" db:"url"`
	URLHash             string    `json:"-" db:"url_hash"`
	Title               *string   `json:"title" db:"title"`
	HTMLVersion         *string   `json:"html_version" db:"html_version"`
	H1Count             int       `json:"h1_count" db:"h1_count"`
	H2Count             int       `json:"h2_count" db:"h2_count"`
	H3Count             int       `json:"h3_count" db:"h3_count"`
	H4Count             int       `json:"h4_count" db:"h4_count"`
	H5Count             int       `json:"h5_count" db:"h5_count"`
	H6Count             int       `json:"h6_count" db:"h6_count"`
	InternalLinksCount  int       `json:"internal_links_count" db:"internal_links_count"`
	ExternalLinksCount  int       `json:"external_links_count" db:"external_links_count"`
	BrokenLinksCount    int       `json:"broken_links_count" db:"broken_links_count"`
	HasLoginForm        bool      `json:"has_login_form" db:"has_login_form"`
	Status              URLStatus `json:"status" db:"status"`
	ErrorMessage        *string   `json:"error_message" db:"error_message"`
	CreatedAt           time.Time `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time `json:"updated_at" db:"updated_at"`
}

type BrokenLink struct {
	ID           int     `json:"id" db:"id"`
	URLID        int     `json:"url_id" db:"url_id"`
	LinkURL      string  `json:"link_url" db:"link_url"`
	StatusCode   int     `json:"status_code" db:"status_code"`
	ErrorMessage *string `json:"error_message" db:"error_message"`
}

type URLRequest struct {
	URL string `json:"url" binding:"required,url"`
}

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	Total      int         `json:"total"`
	TotalPages int         `json:"total_pages"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type BulkRequest struct {
	IDs []int `json:"ids" binding:"required,min=1"`
}

type HeadingCounts struct {
	H1 int `json:"h1"`
	H2 int `json:"h2"`
	H3 int `json:"h3"`
	H4 int `json:"h4"`
	H5 int `json:"h5"`
	H6 int `json:"h6"`
}

type URLAnalysisResult struct {
	Title              string        `json:"title"`
	HTMLVersion        string        `json:"html_version"`
	HeadingCounts      HeadingCounts `json:"heading_counts"`
	InternalLinksCount int           `json:"internal_links_count"`
	ExternalLinksCount int           `json:"external_links_count"`
	BrokenLinksCount   int           `json:"broken_links_count"`
	HasLoginForm       bool          `json:"has_login_form"`
	BrokenLinks        []BrokenLink  `json:"broken_links"`
}

func GenerateURLHash(url string) string {
	hash := sha256.Sum256([]byte(url))
	return fmt.Sprintf("%x", hash)
} 