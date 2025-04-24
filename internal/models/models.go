package models

import (
	"time"
)

// Ad represents a video advertisement
type Ad struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	ImageURL    string    `json:"image_url" db:"image_url"`
	TargetURL   string    `json:"target_url" db:"target_url"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// Click represents a user interaction with an ad
type Click struct {
	ID           string    `json:"id" db:"id"`
	AdID         string    `json:"ad_id" db:"ad_id"`
	Timestamp    time.Time `json:"timestamp" db:"timestamp"`
	IPAddress    string    `json:"ip_address" db:"ip_address"`
	PlaybackTime int       `json:"playback_time" db:"playback_time"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// ArchivedClick has the same structure as Click but is stored in a separate table
type ArchivedClick struct {
	ID           string    `json:"id" db:"id"`
	AdID         string    `json:"ad_id" db:"ad_id"`
	Timestamp    time.Time `json:"timestamp" db:"timestamp"`
	IPAddress    string    `json:"ip_address" db:"ip_address"`
	PlaybackTime int       `json:"playback_time" db:"playback_time"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// AggregatedAnalytics represents precomputed analytics for ads
type AggregatedAnalytics struct {
	ID                string    `json:"id" db:"id"`
	AdID              string    `json:"adId" db:"ad_id"`
	TotalClicks       int       `json:"total_clicks" db:"total_clicks"`
	TotalPlaybackTime int       `json:"total_playback_time" db:"total_playback_time"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
}

// MonthlyAnalytics represents analytics for ads broken down by month
type MonthlyAnalytics struct {
	ID                string    `json:"id" db:"id"`
	AdID              string    `json:"adId" db:"ad_id"`
	Month             int       `json:"month" db:"month"`
	Year              int       `json:"year" db:"year"`
	TotalClicks       int       `json:"total_clicks" db:"total_clicks"`
	TotalPlaybackTime int       `json:"total_playback_time" db:"total_playback_time"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
}

// AnalyticsData represents the response format for analytics API
type AnalyticsData struct {
	AdID                       string  `json:"ad_id"`
	TotalClicks                int     `json:"total_clicks"`
	AverageClicksPerAd         float64 `json:"average_clicks_per_ad"`
	TotalPlaybackTime          int     `json:"total_playback_time"`
	AveragePlaybackTime        float64 `json:"average_playback_time"`
	Period                     string  `json:"period"` // minute, hour, day, week, month
	TotalClicksInRange         int     `json:"total_clicks_in_range"`
	AverageClicksPerAdInRange  float64 `json:"average_clicks_per_ad_in_range"`
	TotalPlaybackTimeInRange   int     `json:"total_playback_time_in_range"`
	AveragePlaybackTimeInRange float64 `json:"average_playback_time_in_range"`
}

// AdAnalyticsData represents the response format for ad analytics API
type AdAnalyticsData struct {
	AdID                       string  `json:"ad_id"`
	TotalClicks                int     `json:"total_clicks"`
	TotalPlaybackTime          int     `json:"total_playback_time"`
	AveragePlaybackTime        float64 `json:"average_playback_time"`
	Period                     string  `json:"period"` // minute, hour, day, week, month
	TotalClicksInRange         int     `json:"total_clicks_in_range"`
	TotalPlaybackTimeInRange   int     `json:"total_playback_time_in_range"`
	AveragePlaybackTimeInRange float64 `json:"average_playback_time_in_range"`
}
