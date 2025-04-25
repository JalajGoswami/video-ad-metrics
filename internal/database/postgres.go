package database

import (
	"database/sql"
	"fmt"
	"net/url"
	"time"

	"github.com/JalajGoswami/video-ad-metrics/internal/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// PostgresDB implements the Repository interface using PostgreSQL
type PostgresDB struct {
	db *sqlx.DB
}

// NewPostgresDB creates a new PostgreSQL repository
func NewPostgresDB(connString string) (*PostgresDB, error) {
	db, err := sqlx.Connect("postgres", connString)
	if err != nil {
		pqErr, ok := err.(*pq.Error)
		if ok && pqErr.Code == "3D000" {
			err = createDatabase(connString)
			if err != nil {
				return nil, fmt.Errorf("failed to create database: %w", err)
			}
			db, err = sqlx.Connect("postgres", connString)
			if err != nil {
				return nil, fmt.Errorf("failed to connect to postgres: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to connect to postgres: %w", err)
		}
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return &PostgresDB{
		db: db,
	}, nil
}

func createDatabase(connString string) error {
	dbUrl, err := url.Parse(connString)
	if err != nil {
		return fmt.Errorf("failed to parse database URL: %w", err)
	}
	dbName := dbUrl.Path[1:]
	dbUrl.Path = "/postgres"
	connString = dbUrl.String()
	db, err := sqlx.Connect("postgres", connString)
	if err != nil {
		return fmt.Errorf("failed to connect to postgres: %w", err)
	}
	_, err = db.Exec(fmt.Sprintf(`CREATE DATABASE %s`, pq.QuoteIdentifier(dbName)))
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}
	err = db.Close()
	if err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}
	return nil
}

func (p *PostgresDB) Ping() error {
	return p.db.Ping()
}

func (p *PostgresDB) Close() error {
	return p.db.Close()
}

// Setup creates the necessary database tables if they don't exist
func (p *PostgresDB) Setup() error {
	// Create ads table
	_, err := p.db.Exec(`
		CREATE TABLE IF NOT EXISTS ads (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			description TEXT,
			image_url TEXT NOT NULL,
			target_url TEXT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create ads table: %w", err)
	}

	// Create clicks table
	_, err = p.db.Exec(`
		CREATE TABLE IF NOT EXISTS clicks (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			ad_id UUID NOT NULL REFERENCES ads(id) ON DELETE CASCADE,
			timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
			ip_address VARCHAR(45) NOT NULL,
			playback_time INTEGER NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create clicks table: %w", err)
	}

	// Create archived_clicks table
	_, err = p.db.Exec(`
		CREATE TABLE IF NOT EXISTS archived_clicks (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			ad_id UUID NOT NULL REFERENCES ads(id) ON DELETE CASCADE,
			timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
			ip_address VARCHAR(45) NOT NULL,
			playback_time INTEGER NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create archived_clicks table: %w", err)
	}

	// Create aggregated_analytics table
	_, err = p.db.Exec(`
		CREATE TABLE IF NOT EXISTS aggregated_analytics (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			ad_id UUID NOT NULL REFERENCES ads(id) ON DELETE CASCADE,
			total_clicks INTEGER NOT NULL DEFAULT 0,
			total_playback_time INTEGER NOT NULL DEFAULT 0,
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create aggregated_analytics table: %w", err)
	}

	// Create monthly_analytics table
	_, err = p.db.Exec(`
		CREATE TABLE IF NOT EXISTS monthly_analytics (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			ad_id UUID NOT NULL REFERENCES ads(id) ON DELETE CASCADE,
			month INTEGER NOT NULL,
			year INTEGER NOT NULL,
			total_clicks INTEGER NOT NULL DEFAULT 0,
			total_playback_time INTEGER NOT NULL DEFAULT 0,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			UNIQUE (ad_id, month, year)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create monthly_analytics table: %w", err)
	}

	// Create an index on ad_id in the clicks table
	_, err = p.db.Exec(`CREATE INDEX IF NOT EXISTS clicks_ad_id_idx ON clicks (ad_id)`)
	if err != nil {
		return fmt.Errorf("failed to create index on clicks: %w", err)
	}

	return nil
}

// ArchiveOldClicks moves clicks older than one month to the archived_clicks table
func (p *PostgresDB) ArchiveOldClicks() error {
	// Begin transaction
	tx, err := p.db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	oneMonthAgo := time.Now().AddDate(0, 0, -30)

	// Insert old clicks into archived_clicks table
	_, err = tx.Exec(`
		INSERT INTO archived_clicks
		SELECT * FROM clicks
		WHERE timestamp < $1
	`, oneMonthAgo)
	if err != nil {
		return fmt.Errorf("failed to insert into archived_clicks: %w", err)
	}

	// Delete the archived clicks from the original table
	_, err = tx.Exec(`
		DELETE FROM clicks
		WHERE timestamp < $1
	`, oneMonthAgo)
	if err != nil {
		return fmt.Errorf("failed to delete from clicks: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// CreateAd stores a new ad
func (p *PostgresDB) CreateAd(ad *models.Ad) error {
	_, err := p.db.NamedExec(`
		INSERT INTO ads (id, name, description, image_url, target_url, created_at)
		VALUES (:id, :name, :description, :image_url, :target_url, :created_at)
	`, ad)
	if err != nil {
		return fmt.Errorf("failed to insert ad: %w", err)
	}

	// Create initial aggregated analytics entry for this ad
	_, err = p.db.Exec(`
		INSERT INTO aggregated_analytics (id, ad_id, total_clicks, total_playback_time)
		VALUES ($1, $2, 0, 0)
	`, uuid.New().String(), ad.ID)
	if err != nil {
		return fmt.Errorf("failed to insert initial analytics: %w", err)
	}

	return nil
}

// GetAd retrieves an ad by ID
func (p *PostgresDB) GetAd(id string) (*models.Ad, error) {
	var ad models.Ad
	err := p.db.Get(&ad, `SELECT * FROM ads WHERE id = $1`, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get ad: %w", err)
	}
	return &ad, nil
}

// ListAds returns all ads
func (p *PostgresDB) ListAds(opts ListAdOptions) (*[]models.Ad, error) {
	ads := []models.Ad{}
	query := `SELECT * FROM ads`
	if opts.Search != "" {
		query += ` WHERE name ILIKE '%' || :search || '%'`
	}
	if opts.Order == "asc" {
		query += ` ORDER BY "created_at" ASC`
	} else {
		query += ` ORDER BY "created_at" DESC`
	}
	query += ` LIMIT :limit OFFSET :offset;`
	stmt, err := p.db.PrepareNamed(query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}
	err = stmt.Select(&ads, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list ads: %w", err)
	}
	return &ads, nil
}

// used for pagination
func (p *PostgresDB) CountAds(opts ListAdOptions) (int, error) {
	var count int
	var err error
	query := `SELECT COUNT(*) FROM ads`
	if opts.Search != "" {
		query += ` WHERE name ILIKE '%' || $1 || '%'`
		err = p.db.Get(&count, query, opts.Search)
	} else {
		err = p.db.Get(&count, query)
	}
	if err != nil {
		return 0, fmt.Errorf("failed to count ads: %w", err)
	}
	return count, nil
}

// LogClick records a click and updates analytics
func (p *PostgresDB) LogClick(click *models.Click) error {
	tx, err := p.db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var exists bool
	err = tx.Get(&exists, `SELECT EXISTS(SELECT 1 FROM ads WHERE id = $1)`, click.AdID)
	if err != nil {
		return fmt.Errorf("failed to check if ad exists: %w", err)
	}

	if !exists {
		return ErrNotFound
	}

	_, err = tx.NamedExec(`
		INSERT INTO clicks (id, ad_id, timestamp, ip_address, playback_time, created_at)
		VALUES (:id, :ad_id, :timestamp, :ip_address, :playback_time, :created_at)
	`, click)
	if err != nil {
		return fmt.Errorf("failed to insert click: %w", err)
	}

	// Update aggregated analytics
	_, err = tx.Exec(`
		UPDATE aggregated_analytics
		SET total_clicks = total_clicks + 1,
			total_playback_time = total_playback_time + $1,
			updated_at = NOW()
		WHERE ad_id = $2
	`, click.PlaybackTime, click.AdID)
	if err != nil {
		return fmt.Errorf("failed to update aggregated analytics: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetAdAnalytics retrieves analytics for a specific ad
func (p *PostgresDB) GetAdAnalytics(adID string, rangeDate time.Time) (*models.AdAnalyticsData, error) {
	// Check if ad exists
	var exists bool
	err := p.db.Get(&exists, `SELECT EXISTS(SELECT 1 FROM ads WHERE id = $1)`, adID)
	if err != nil {
		return nil, fmt.Errorf("failed to check if ad exists: %w", err)
	}
	if !exists {
		return nil, ErrNotFound
	}

	var result models.AdAnalyticsData
	err = p.db.Get(&result, `SELECT ad_id, total_clicks, total_playback_time FROM aggregated_analytics WHERE ad_id = $1`, adID)
	if err != nil {
		return nil, fmt.Errorf("failed to get aggregated analytics: %w", err)
	}

	// Query combines current and archived clicks
	err = p.db.Get(&result, `
		SELECT 
			COUNT(*) as total_clicks_in_range,
			COALESCE(SUM(playback_time), 0) as total_playback_time_in_range
		FROM (
			SELECT playback_time FROM clicks
			WHERE ad_id = $1 AND timestamp >= $2
		)
	`, adID, rangeDate)

	if err != nil {
		return nil, fmt.Errorf("failed to get range analytics: %w", err)
	}

	return &result, nil
}

// GetAdsAnalytics retrieves aggregate analytics for all ads
func (p *PostgresDB) GetAdsAnalytics(rangeDate time.Time) (*models.AnalyticsData, error) {
	var result struct {
		models.AnalyticsData
		AdCount int `db:"ad_count"`
	}

	err := p.db.Get(&result, `
		SELECT 
			SUM(total_clicks) AS total_clicks,
			SUM(total_playback_time) AS total_playback_time,
			COUNT(*) AS ad_count
		FROM aggregated_analytics
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get aggregated analytics: %w", err)
	}

	if result.AdCount > 0 {
		result.AverageClicksPerAd = float64(result.TotalClicks) / float64(result.AdCount)
	}

	err = p.db.Get(&result, `
		SELECT 
			COUNT(*) AS total_clicks_in_range,
			COALESCE(SUM(playback_time), 0) AS total_playback_time_in_range,
			COUNT(DISTINCT ad_id) AS ad_count
		FROM clicks
		WHERE timestamp >= $1
	`, rangeDate)

	if err != nil {
		return nil, fmt.Errorf("failed to get today's analytics: %w", err)
	}

	if result.AdCount > 0 {
		result.AverageClicksPerAdInRange = float64(result.TotalClicksInRange) / float64(result.AdCount)
	}

	return &result.AnalyticsData, nil
}
