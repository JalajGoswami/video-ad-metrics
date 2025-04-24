package database

import (
	"errors"
	"time"

	apihelpers "github.com/JalajGoswami/video-ad-metrics/internal/api-helpers"
	"github.com/JalajGoswami/video-ad-metrics/internal/models"
)

var (
	ErrNotFound  = errors.New("record not found")
	ErrInvalidID = errors.New("invalid id")
)

// interface for database operations which can have different database implementations
// like mock database for testing
type Repository interface {
	// Config operations
	Setup() error
	Close() error

	// Ad operations
	CreateAd(ad *models.Ad) error
	GetAd(id string) (*models.Ad, error)
	ListAds(opts ListAdOptions) (*[]models.Ad, error)
	CountAds(opts ListAdOptions) (int, error)

	// Click operations
	LogClick(click *models.Click) error
	ArchiveOldClicks() error

	// Analytics operations
	GetAdAnalytics(adID string, rangeDate time.Time) (*models.AdAnalyticsData, error)
	GetAdsAnalytics(rangeDate time.Time) (*models.AnalyticsData, error)
}

type ListAdOptions struct {
	apihelpers.PaginationOptions
	apihelpers.SortOrderOptions
	Search string
}

func (o *ListAdOptions) Default() {
	o.PaginationOptions.Default()
	o.SortOrderOptions.Default()
}
