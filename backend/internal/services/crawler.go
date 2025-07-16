package services

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/html"
	"seacher-app/internal/models"
)

type CrawlerService struct {
	db *sql.DB
}

func NewCrawlerService(db *sql.DB) *CrawlerService {
	return &CrawlerService{db: db}
}

func (s *CrawlerService) AddURL(urlStr string) (*models.URL, error) {
	urlHash := models.GenerateURLHash(urlStr)
	
	var existingID int
	checkQuery := "SELECT id FROM urls WHERE url_hash = ?"
	err := s.db.QueryRow(checkQuery, urlHash).Scan(&existingID)
	if err == nil {
		return s.GetURL(existingID)
	} else if err != sql.ErrNoRows {
		return nil, err
	}
	
	query := "INSERT INTO urls (url, url_hash, status) VALUES (?, ?, ?)"
	result, err := s.db.Exec(query, urlStr, urlHash, models.StatusQueued)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &models.URL{
		ID:      int(id),
		URL:     urlStr,
		URLHash: urlHash,
		Status:  models.StatusQueued,
	}, nil
}

func (s *CrawlerService) GetURLs(page, limit int, search string) ([]models.URL, int, error) {
	var urls []models.URL
	var total int

	countQuery := "SELECT COUNT(*) FROM urls WHERE url LIKE ?"
	searchPattern := "%" + search + "%"
	err := s.db.QueryRow(countQuery, searchPattern).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	query := `
		SELECT id, url, url_hash, title, html_version, h1_count, h2_count, h3_count, h4_count, h5_count, h6_count,
		       internal_links_count, external_links_count, broken_links_count, has_login_form, status,
		       error_message, created_at, updated_at
		FROM urls 
		WHERE url LIKE ? 
		ORDER BY created_at DESC 
		LIMIT ? OFFSET ?`

	rows, err := s.db.Query(query, searchPattern, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var url models.URL
		var title, htmlVersion, errorMessage sql.NullString
		err := rows.Scan(
			&url.ID, &url.URL, &url.URLHash, &title, &htmlVersion,
			&url.H1Count, &url.H2Count, &url.H3Count, &url.H4Count, &url.H5Count, &url.H6Count,
			&url.InternalLinksCount, &url.ExternalLinksCount, &url.BrokenLinksCount,
			&url.HasLoginForm, &url.Status, &errorMessage, &url.CreatedAt, &url.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		if title.Valid {
			url.Title = &title.String
		}
		if htmlVersion.Valid {
			url.HTMLVersion = &htmlVersion.String
		}
		if errorMessage.Valid {
			url.ErrorMessage = &errorMessage.String
		}

		urls = append(urls, url)
	}

	return urls, total, nil
}

func (s *CrawlerService) GetURL(id int) (*models.URL, error) {
	var url models.URL
	var title, htmlVersion, errorMessage sql.NullString
	query := `
		SELECT id, url, url_hash, title, html_version, h1_count, h2_count, h3_count, h4_count, h5_count, h6_count,
		       internal_links_count, external_links_count, broken_links_count, has_login_form, status,
		       error_message, created_at, updated_at
		FROM urls WHERE id = ?`

	err := s.db.QueryRow(query, id).Scan(
		&url.ID, &url.URL, &url.URLHash, &title, &htmlVersion,
		&url.H1Count, &url.H2Count, &url.H3Count, &url.H4Count, &url.H5Count, &url.H6Count,
		&url.InternalLinksCount, &url.ExternalLinksCount, &url.BrokenLinksCount,
		&url.HasLoginForm, &url.Status, &errorMessage, &url.CreatedAt, &url.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	if title.Valid {
		url.Title = &title.String
	}
	if htmlVersion.Valid {
		url.HTMLVersion = &htmlVersion.String
	}
	if errorMessage.Valid {
		url.ErrorMessage = &errorMessage.String
	}

	return &url, nil
}

func (s *CrawlerService) AnalyzeURL(id int) error {
	_, err := s.db.Exec("UPDATE urls SET status = ? WHERE id = ?", models.StatusProcessing, id)
	if err != nil {
		return err
	}

	url, err := s.GetURL(id)
	if err != nil {
		return err
	}

	result, err := s.crawlURL(url.URL)
	if err != nil {
		_, updateErr := s.db.Exec("UPDATE urls SET status = ?, error_message = ? WHERE id = ?",
			models.StatusError, err.Error(), id)
		if updateErr != nil {
			return updateErr
		}
		return err
	}

	updateQuery := `
		UPDATE urls SET 
			title = ?, html_version = ?, h1_count = ?, h2_count = ?, h3_count = ?, h4_count = ?, h5_count = ?, h6_count = ?,
			internal_links_count = ?, external_links_count = ?, broken_links_count = ?, has_login_form = ?,
			status = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`

	_, err = s.db.Exec(updateQuery,
		result.Title, result.HTMLVersion,
		result.HeadingCounts.H1, result.HeadingCounts.H2, result.HeadingCounts.H3,
		result.HeadingCounts.H4, result.HeadingCounts.H5, result.HeadingCounts.H6,
		result.InternalLinksCount, result.ExternalLinksCount, result.BrokenLinksCount,
		result.HasLoginForm, models.StatusCompleted, id)

	return err
}

func (s *CrawlerService) crawlURL(urlStr string) (*models.URLAnalysisResult, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(urlStr)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	result := &models.URLAnalysisResult{
		HTMLVersion: "HTML5", // Simplified
	}

	s.analyzeNode(doc, result)

	return result, nil
}

func (s *CrawlerService) analyzeNode(n *html.Node, result *models.URLAnalysisResult) {
	if n.Type == html.ElementNode {
		switch strings.ToLower(n.Data) {
		case "title":
			if n.FirstChild != nil {
				result.Title = n.FirstChild.Data
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
		case "input":
			for _, attr := range n.Attr {
				if attr.Key == "type" && attr.Val == "password" {
					result.HasLoginForm = true
					break
				}
			}
		case "a":
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					if strings.HasPrefix(attr.Val, "http") {
						result.ExternalLinksCount++
					} else {
						result.InternalLinksCount++
					}
					break
				}
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		s.analyzeNode(c, result)
	}
}

func (s *CrawlerService) DeleteURL(id int) error {
	_, err := s.db.Exec("DELETE FROM urls WHERE id = ?", id)
	return err
} 