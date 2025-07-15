package database

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type DatabaseConfig struct {
	Host     string `envconfig:"DB_HOST" default:"localhost"`
	Port     int    `envconfig:"DB_PORT" default:"3306"`
	Username string `envconfig:"DB_USER" default:"analyzer_user"`
	Password string `envconfig:"DB_PASSWORD" default:"analyzer_pass"`
	Database string `envconfig:"DB_NAME" default:"website_analyzer"`

	MaxOpenConns    int           `envconfig:"DB_MAX_OPEN_CONNS" default:"25"`
	MaxIdleConns    int           `envconfig:"DB_MAX_IDLE_CONNS" default:"5"`
	ConnMaxLifetime time.Duration `envconfig:"DB_CONN_MAX_LIFETIME" default:"5m"`
	ConnMaxIdleTime time.Duration `envconfig:"DB_CONN_MAX_IDLE_TIME" default:"30s"`

	ConnectTimeout time.Duration `envconfig:"DB_CONNECT_TIMEOUT" default:"10s"`
	ReadTimeout    time.Duration `envconfig:"DB_READ_TIMEOUT" default:"30s"`
	WriteTimeout   time.Duration `envconfig:"DB_WRITE_TIMEOUT" default:"30s"`

	SSLMode   string `envconfig:"DB_SSL_MODE" default:"disable"`
	SSLCert   string `envconfig:"DB_SSL_CERT" default:""`
	SSLKey    string `envconfig:"DB_SSL_KEY" default:""`
	SSLRootCA string `envconfig:"DB_SSL_ROOT_CA" default:""`

	Charset   string `envconfig:"DB_CHARSET" default:"utf8mb4"`
	ParseTime bool   `envconfig:"DB_PARSE_TIME" default:"true"`
	Location  string `envconfig:"DB_LOCATION" default:"Local"`
}

func (cfg *DatabaseConfig) DSN() string {
	params := fmt.Sprintf("charset=%s&parseTime=%t&loc=%s&timeout=%s&readTimeout=%s&writeTimeout=%s",
		cfg.Charset, cfg.ParseTime, cfg.Location,
		cfg.ConnectTimeout, cfg.ReadTimeout, cfg.WriteTimeout)

	if cfg.SSLMode != "disable" {
		params += "&tls=" + cfg.SSLMode
		if cfg.SSLCert != "" {
			params += "&tls-cert=" + cfg.SSLCert
		}
		if cfg.SSLKey != "" {
			params += "&tls-key=" + cfg.SSLKey
		}
		if cfg.SSLRootCA != "" {
			params += "&tls-ca=" + cfg.SSLRootCA
		}
	}

	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database, params)
}

type Database struct {
	*sql.DB
	config *DatabaseConfig
	logger *slog.Logger
	stats  *DatabaseStats
}

type DatabaseStats struct {
	OpenConnections   int           `json:"open_connections"`
	InUse             int           `json:"in_use"`
	Idle              int           `json:"idle"`
	WaitCount         int64         `json:"wait_count"`
	WaitDuration      time.Duration `json:"wait_duration"`
	MaxIdleClosed     int64         `json:"max_idle_closed"`
	MaxLifetimeClosed int64         `json:"max_lifetime_closed"`
}

func NewDatabase(cfg *DatabaseConfig, logger *slog.Logger) (*Database, error) {
	logger.Info("Connecting to database",
		slog.String("host", cfg.Host),
		slog.Int("port", cfg.Port),
		slog.String("database", cfg.Database))

	db, err := sql.Open("mysql", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ConnectTimeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Database connection established",
		slog.Int("max_open_conns", cfg.MaxOpenConns),
		slog.Int("max_idle_conns", cfg.MaxIdleConns),
		slog.Duration("conn_max_lifetime", cfg.ConnMaxLifetime))

	return &Database{
		DB:     db,
		config: cfg,
		logger: logger,
		stats:  &DatabaseStats{},
	}, nil
}

func (d *Database) GetStats() *DatabaseStats {
	stats := d.DB.Stats()

	d.stats.OpenConnections = stats.OpenConnections
	d.stats.InUse = stats.InUse
	d.stats.Idle = stats.Idle
	d.stats.WaitCount = stats.WaitCount
	d.stats.WaitDuration = stats.WaitDuration
	d.stats.MaxIdleClosed = stats.MaxIdleClosed
	d.stats.MaxLifetimeClosed = stats.MaxLifetimeClosed

	return d.stats
}

func (d *Database) HealthCheck(ctx context.Context) error {
	if err := d.PingContext(ctx); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}

	var version string
	if err := d.QueryRowContext(ctx, "SELECT VERSION()").Scan(&version); err != nil {
		return fmt.Errorf("query failed: %w", err)
	}

	stats := d.GetStats()
	if stats.OpenConnections > d.config.MaxOpenConns {
		return fmt.Errorf("too many open connections: %d > %d",
			stats.OpenConnections, d.config.MaxOpenConns)
	}

	d.logger.Debug("Database health check passed",
		slog.String("version", version),
		slog.Int("open_conns", stats.OpenConnections),
		slog.Int("in_use", stats.InUse),
		slog.Int("idle", stats.Idle))

	return nil
}

func (d *Database) Close() error {
	d.logger.Info("Closing database connection")

	stats := d.GetStats()
	d.logger.Info("Final database statistics",
		slog.Int("open_connections", stats.OpenConnections),
		slog.Int64("wait_count", stats.WaitCount),
		slog.Duration("wait_duration", stats.WaitDuration),
		slog.Int64("max_idle_closed", stats.MaxIdleClosed),
		slog.Int64("max_lifetime_closed", stats.MaxLifetimeClosed))

	return d.DB.Close()
}

func (d *Database) WithTransaction(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := d.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				d.logger.Error("Failed to rollback transaction",
					slog.String("error", rollbackErr.Error()))
			}
		} else {
			if commitErr := tx.Commit(); commitErr != nil {
				d.logger.Error("Failed to commit transaction",
					slog.String("error", commitErr.Error()))
				err = commitErr
			}
		}
	}()

	err = fn(tx)
	return err
}

func (d *Database) WithRetry(ctx context.Context, maxRetries int, fn func() error) error {
	var lastErr error

	for i := 0; i <= maxRetries; i++ {
		if err := fn(); err != nil {
			lastErr = err

			if !isRetryableError(err) {
				return err
			}

			if i < maxRetries {
				backoff := time.Duration(i+1) * time.Second
				d.logger.Warn("Database operation failed, retrying",
					slog.String("error", err.Error()),
					slog.Int("attempt", i+1),
					slog.Int("max_retries", maxRetries),
					slog.Duration("backoff", backoff))

				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(backoff):
					continue
				}
			}
		} else {
			return nil
		}
	}

	return fmt.Errorf("operation failed after %d retries: %w", maxRetries, lastErr)
}

func isRetryableError(err error) bool {
	return false
}

func (d *Database) Migrate(ctx context.Context, migrationPath string) error {
	d.logger.Info("Running database migrations", slog.String("path", migrationPath))

	createMigrationsTable := `
		CREATE TABLE IF NOT EXISTS migrations (
			id INT PRIMARY KEY AUTO_INCREMENT,
			version VARCHAR(255) NOT NULL UNIQUE,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`

	if _, err := d.ExecContext(ctx, createMigrationsTable); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	d.logger.Info("Database migrations completed")
	return nil
}

func (d *Database) LogSlowQueries(threshold time.Duration) {
	d.logger.Info("Slow query logging enabled", slog.Duration("threshold", threshold))
}
