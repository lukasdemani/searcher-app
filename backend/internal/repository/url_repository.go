package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"website-analyzer/internal/models"
)

type URLRepository interface {
	Save(ctx context.Context, url *models.URL) error
	FindByID(ctx context.Context, id int) (*models.URL, error)
	FindByHash(ctx context.Context, hash string) (*models.URL, error)
	FindAll(ctx context.Context, filter URLFilter) ([]models.URL, int, error)
	Update(ctx context.Context, url *models.URL) error
	Delete(ctx context.Context, id int) error
	DeleteBatch(ctx context.Context, ids []int) error
}

type URLFilter struct {
	Search string
	Status models.URLStatus
	Page   int
	Limit  int
}

type MySQLURLRepository struct {
	db *sql.DB
}

func NewMySQLURLRepository(db *sql.DB) URLRepository {
	return &MySQLURLRepository{db: db}
}

func (r *MySQLURLRepository) Save(ctx context.Context, url *models.URL) error {
	query := `
		INSERT INTO urls (url, url_hash, title, html_version, h1_count, h2_count, h3_count, h4_count, h5_count, h6_count,
		                 internal_links_count, external_links_count, broken_links_count, has_login_form, status, error_message)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := r.db.ExecContext(ctx, query,
		url.URL, url.URLHash, url.Title, url.HTMLVersion,
		url.H1Count, url.H2Count, url.H3Count, url.H4Count, url.H5Count, url.H6Count,
		url.InternalLinksCount, url.ExternalLinksCount, url.BrokenLinksCount,
		url.HasLoginForm, url.Status, url.ErrorMessage)

	if err != nil {
		return fmt.Errorf("failed to save URL: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert ID: %w", err)
	}

	url.ID = int(id)
	return nil
}

func (r *MySQLURLRepository) FindByID(ctx context.Context, id int) (*models.URL, error) {
	query := `
		SELECT id, url, url_hash, title, html_version, h1_count, h2_count, h3_count, h4_count, h5_count, h6_count,
		       internal_links_count, external_links_count, broken_links_count, has_login_form, status,
		       error_message, created_at, updated_at
		FROM urls WHERE id = ?`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var url models.URL
	var title, htmlVersion, errorMessage sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&url.ID, &url.URL, &url.URLHash, &title, &htmlVersion,
		&url.H1Count, &url.H2Count, &url.H3Count, &url.H4Count, &url.H5Count, &url.H6Count,
		&url.InternalLinksCount, &url.ExternalLinksCount, &url.BrokenLinksCount,
		&url.HasLoginForm, &url.Status, &errorMessage, &url.CreatedAt, &url.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("URL not found with ID %d", id)
		}
		return nil, fmt.Errorf("failed to find URL by ID: %w", err)
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

func (r *MySQLURLRepository) FindByHash(ctx context.Context, hash string) (*models.URL, error) {
	query := `
		SELECT id, url, url_hash, title, html_version, h1_count, h2_count, h3_count, h4_count, h5_count, h6_count,
		       internal_links_count, external_links_count, broken_links_count, has_login_form, status,
		       error_message, created_at, updated_at
		FROM urls WHERE url_hash = ?`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var url models.URL
	var title, htmlVersion, errorMessage sql.NullString

	err := r.db.QueryRowContext(ctx, query, hash).Scan(
		&url.ID, &url.URL, &url.URLHash, &title, &htmlVersion,
		&url.H1Count, &url.H2Count, &url.H3Count, &url.H4Count, &url.H5Count, &url.H6Count,
		&url.InternalLinksCount, &url.ExternalLinksCount, &url.BrokenLinksCount,
		&url.HasLoginForm, &url.Status, &errorMessage, &url.CreatedAt, &url.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find URL by hash: %w", err)
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

func (r *MySQLURLRepository) FindAll(ctx context.Context, filter URLFilter) ([]models.URL, int, error) {
	whereClause := "WHERE 1=1"
	args := []interface{}{}

	if filter.Search != "" {
		whereClause += " AND url LIKE ?"
		args = append(args, "%"+filter.Search+"%")
	}

	if filter.Status != "" {
		whereClause += " AND status = ?"
		args = append(args, filter.Status)
	}

	countQuery := "SELECT COUNT(*) FROM urls " + whereClause

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count URLs: %w", err)
	}

	offset := (filter.Page - 1) * filter.Limit
	query := `
		SELECT id, url, url_hash, title, html_version, h1_count, h2_count, h3_count, h4_count, h5_count, h6_count,
		       internal_links_count, external_links_count, broken_links_count, has_login_form, status,
		       error_message, created_at, updated_at
		FROM urls ` + whereClause + `
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`

	args = append(args, filter.Limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query URLs: %w", err)
	}
	defer rows.Close()

	var urls []models.URL
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
			return nil, 0, fmt.Errorf("failed to scan URL: %w", err)
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

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}

	return urls, total, nil
}

func (r *MySQLURLRepository) Update(ctx context.Context, url *models.URL) error {
	query := `
		UPDATE urls SET 
			title = ?, html_version = ?, h1_count = ?, h2_count = ?, h3_count = ?, h4_count = ?, h5_count = ?, h6_count = ?,
			internal_links_count = ?, external_links_count = ?, broken_links_count = ?, has_login_form = ?,
			status = ?, error_message = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := r.db.ExecContext(ctx, query,
		url.Title, url.HTMLVersion,
		url.H1Count, url.H2Count, url.H3Count, url.H4Count, url.H5Count, url.H6Count,
		url.InternalLinksCount, url.ExternalLinksCount, url.BrokenLinksCount,
		url.HasLoginForm, url.Status, url.ErrorMessage, url.ID)

	if err != nil {
		return fmt.Errorf("failed to update URL: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if affected == 0 {
		return fmt.Errorf("no URL found with ID %d", url.ID)
	}

	return nil
}

func (r *MySQLURLRepository) Delete(ctx context.Context, id int) error {
	query := "DELETE FROM urls WHERE id = ?"

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete URL: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if affected == 0 {
		return fmt.Errorf("no URL found with ID %d", id)
	}

	return nil
}

func (r *MySQLURLRepository) DeleteBatch(ctx context.Context, ids []int) error {
	if len(ids) == 0 {
		return nil
	}

	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf("DELETE FROM urls WHERE id IN (%s)",
		fmt.Sprintf("%s", placeholders[0]))
	for i := 1; i < len(placeholders); i++ {
		query = fmt.Sprintf("%s,%s", query, placeholders[i])
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete URLs: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if affected == 0 {
		return fmt.Errorf("no URLs found with provided IDs")
	}

	return nil
}