package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/JalajGoswami/video-ad-metrics/internal/database"
	"github.com/JalajGoswami/video-ad-metrics/internal/logger"
	"github.com/JalajGoswami/video-ad-metrics/internal/models"
)

// MockData structure to load data from JSON
type MockData struct {
	Ads                 []models.Ad                  `json:"ads"`
	Clicks              []models.Click               `json:"clicks"`
	ArchivedClicks      []models.ArchivedClick       `json:"archived_clicks"`
	AggregatedAnalytics []models.AggregatedAnalytics `json:"aggregated_analytics"`
	MonthlyAnalytics    []models.MonthlyAnalytics    `json:"monthly_analytics"`
}

func main() {
	godotenv.Load()
	connString := os.Getenv("DATABASE_URL")

	dbUrl, err := url.Parse(connString)
	if err != nil {
		logger.FatalLog("failed to parse database URL: %v", err)
	}
	dbName := dbUrl.Path[1:]
	dbUrl.Path = "/postgres"
	connString = dbUrl.String()
	db, err := sqlx.Connect("postgres", connString)
	if err != nil {
		logger.FatalLog("Error connecting to PostgreSQL: %v", err)
	}

	logger.InfoLog("Successfully connected to PostgreSQL")

	_, err = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName))
	if err != nil {
		logger.FatalLog("Error dropping database: %v", err)
	}
	logger.InfoLog("Dropped database %s if it existed", dbName)

	// Create database
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName))
	if err != nil {
		logger.FatalLog("Error creating database: %v", err)
	}
	logger.InfoLog("Created database %s", dbName)

	db.Close()

	// Connect to the new database
	dbUrl.Path = dbName
	connString = dbUrl.String()
	db, err = sqlx.Connect("postgres", connString)
	if err != nil {
		logger.FatalLog("Error connecting to %s database: %v", dbName, err)
	}
	defer db.Close()

	logger.InfoLog("Successfully connected to %s database", dbName)

	// Create tables
	createTables(connString)

	// Load mock data
	mockData, err := loadMockData()
	if err != nil {
		logger.FatalLog("Error loading mock data: %v", err)
	}

	// Insert mock data
	insertMockData(db, mockData)

	logger.InfoLog("Database setup completed successfully!")
}

func createTables(dbUrl string) {
	pgDb, err := database.NewPostgresDB(dbUrl)
	if err != nil {
		logger.FatalLog("Error connecting to new DB: %v", err)
	}
	err = pgDb.Setup()
	if err != nil {
		logger.FatalLog("Error creating tables: %v", err)
	}
}

func loadMockData() (MockData, error) {
	var mockData MockData

	// Get the executable path
	ex, err := os.Executable()
	if err != nil {
		return mockData, err
	}
	exPath := filepath.Dir(ex)

	// Read the mock data file
	jsonFile, err := os.ReadFile(filepath.Join(exPath, "mock-data.json"))
	if err != nil {
		// Try to read from current directory
		jsonFile, err = os.ReadFile("cmd/create-db/mock-data.json")
		if err != nil {
			return mockData, err
		}
	}

	err = json.Unmarshal(jsonFile, &mockData)
	if err != nil {
		return mockData, err
	}

	return mockData, nil
}

func insertMockData(db *sqlx.DB, mockData MockData) {
	// Insert ads
	for _, ad := range mockData.Ads {
		_, err := db.NamedExec(`
			INSERT INTO ads (id, name, description, image_url, target_url, created_at)
			VALUES (:id, :name, :description, :image_url, :target_url, :created_at)
		`, ad)

		if err != nil {
			logger.FatalLog("Error inserting ad %s: %v", ad.ID, err)
		}
	}
	logger.InfoLog("Inserted %d ads", len(mockData.Ads))

	// Insert clicks
	for _, click := range mockData.Clicks {
		_, err := db.NamedExec(`
			INSERT INTO clicks (id, ad_id, timestamp, ip_address, playback_time, created_at)
			VALUES (:id, :ad_id, :timestamp, :ip_address, :playback_time, :created_at)
		`, click)

		if err != nil {
			logger.FatalLog("Error inserting click %s: %v", click.ID, err)
		}
	}
	logger.InfoLog("Inserted %d clicks", len(mockData.Clicks))

	// Insert archived clicks
	for _, click := range mockData.ArchivedClicks {
		_, err := db.NamedExec(`
			INSERT INTO archived_clicks (id, ad_id, timestamp, ip_address, playback_time, created_at)
			VALUES (:id, :ad_id, :timestamp, :ip_address, :playback_time, :created_at)
		`, click)

		if err != nil {
			logger.FatalLog("Error inserting archived click %s: %v", click.ID, err)
		}
	}
	logger.InfoLog("Inserted %d archived clicks", len(mockData.ArchivedClicks))

	// Insert aggregated analytics
	for _, analytics := range mockData.AggregatedAnalytics {
		_, err := db.NamedExec(`
			INSERT INTO aggregated_analytics (id, ad_id, total_clicks, total_playback_time, updated_at, created_at)
			VALUES (:id, :ad_id, :total_clicks, :total_playback_time, :updated_at, :created_at)
		`, analytics)

		if err != nil {
			logger.FatalLog("Error inserting aggregated analytics %s: %v", analytics.ID, err)
		}
	}
	logger.InfoLog("Inserted %d aggregated analytics", len(mockData.AggregatedAnalytics))

	// Insert monthly analytics
	for _, analytics := range mockData.MonthlyAnalytics {
		_, err := db.NamedExec(`
			INSERT INTO monthly_analytics (id, ad_id, month, year, total_clicks, total_playback_time, created_at)
			VALUES (:id, :ad_id, :month, :year, :total_clicks, :total_playback_time, :created_at)
		`, analytics)

		if err != nil {
			logger.FatalLog("Error inserting monthly analytics %s: %v", analytics.ID, err)
		}
	}
	logger.InfoLog("Inserted %d monthly analytics", len(mockData.MonthlyAnalytics))
}
