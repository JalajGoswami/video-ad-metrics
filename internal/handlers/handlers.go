package handlers

import (
	"encoding/json"
	"net/http"
	"slices"
	"time"

	apihelpers "github.com/JalajGoswami/video-ad-metrics/internal/api-helpers"
	"github.com/JalajGoswami/video-ad-metrics/internal/database"
	"github.com/JalajGoswami/video-ad-metrics/internal/logger"
	"github.com/JalajGoswami/video-ad-metrics/internal/models"
	"github.com/google/uuid"
)

// Handler contains the dependencies needed for the HTTP handlers
type Handler struct {
	DB database.Repository
}

// NewHandler creates a new Handler
func NewHandler(db database.Repository) *Handler {
	return &Handler{
		DB: db,
	}
}

func (h *Handler) CreateAd(w http.ResponseWriter, r *http.Request) {
	var ad models.Ad
	if err := json.NewDecoder(r.Body).Decode(&ad); err != nil {
		logger.RequestLogger.Error(r, "Error decoding request body: %v", err)
		apihelpers.ErrorResponse(r, w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	ad.ID = uuid.New().String()
	ad.CreatedAt = time.Now()

	if err := h.DB.CreateAd(&ad); err != nil {
		logger.RequestLogger.Error(r, "Error creating ad: %v", err)
		apihelpers.ErrorResponse(r, w, http.StatusInternalServerError, "Error creating ad")
		return
	}

	apihelpers.SuccessResponse(r, w, http.StatusCreated, ad, "Ad created successfully")
}

// GetAd retrieves an ad by ID
func (h *Handler) GetAd(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if uuid.Validate(id) != nil {
		logger.RequestLogger.Error(r, "Invalid ad ID: %v", id)
		apihelpers.ErrorResponse(r, w, http.StatusBadRequest, "Invalid ad ID")
		return
	}

	ad, err := h.DB.GetAd(id)
	if err != nil {
		if err == database.ErrNotFound {
			logger.RequestLogger.Error(r, "Ad not found")
			apihelpers.ErrorResponse(r, w, http.StatusNotFound, "Ad not found")
		} else {
			logger.RequestLogger.Error(r, "Error retrieving ad: %v", err)
			apihelpers.ErrorResponse(r, w, http.StatusInternalServerError, "Error retrieving ad")
		}
		return
	}

	apihelpers.SuccessResponse(r, w, http.StatusOK, ad, "")
}

// ListAds returns all ads
func (h *Handler) ListAds(w http.ResponseWriter, r *http.Request) {
	opts := database.ListAdOptions{}
	query := r.URL.Query()
	opts.Search = query.Get("search")
	opts.Order = query.Get("order")
	pageOpts, getPaginationObject, err := apihelpers.Pagination(r)
	if err != nil {
		logger.RequestLogger.Error(r, "Error in pagination parameters: %v", err)
		apihelpers.ErrorResponse(r, w, http.StatusBadRequest, err.Error())
		return
	}
	opts.PaginationOptions = pageOpts
	opts.Default()

	ads, err := h.DB.ListAds(opts)
	if err != nil {
		logger.RequestLogger.Error(r, "Error retrieving ads: %v", err)
		apihelpers.ErrorResponse(r, w, http.StatusInternalServerError, "Error retrieving ads")
		return
	}
	totalCount, err := h.DB.CountAds(opts)
	if err != nil {
		logger.RequestLogger.Error(r, "Error retrieving ads count: %v", err)
		apihelpers.ErrorResponse(r, w, http.StatusInternalServerError, "Error retrieving ads count")
		return
	}

	result := map[string]any{
		"values": ads,
		"pages":  getPaginationObject(len(*ads), totalCount),
	}
	apihelpers.SuccessResponse(r, w, http.StatusOK, result, "")
}

// LogClick records a click on an ad
func (h *Handler) LogClick(w http.ResponseWriter, r *http.Request) {
	var click models.Click
	if err := json.NewDecoder(r.Body).Decode(&click); err != nil {
		logger.RequestLogger.Error(r, "Error decoding request body: %v", err)
		apihelpers.ErrorResponse(r, w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if uuid.Validate(click.AdID) != nil {
		logger.RequestLogger.Error(r, "Invalid ad ID: %v", click.AdID)
		apihelpers.ErrorResponse(r, w, http.StatusBadRequest, "Invalid ad ID")
		return
	}

	click.ID = uuid.New().String()

	if click.Timestamp.IsZero() {
		click.Timestamp = time.Now().Add(-time.Duration(click.PlaybackTime) * time.Second)
	}
	click.CreatedAt = time.Now()

	if click.IPAddress == "" {
		click.IPAddress = r.RemoteAddr
	}

	if err := h.DB.LogClick(&click); err != nil {
		if err == database.ErrNotFound {
			logger.RequestLogger.Error(r, "Ad not found")
			apihelpers.ErrorResponse(r, w, http.StatusNotFound, "Ad not found")
		} else {
			logger.RequestLogger.Error(r, "Error logging click: %v", err)
			apihelpers.ErrorResponse(r, w, http.StatusInternalServerError, "Error logging click")
		}
		return
	}

	apihelpers.SuccessResponse(r, w, http.StatusCreated, click, "Click logged successfully")
}

var durationMap = map[string]time.Duration{
	"minute": time.Minute,
	"hour":   time.Hour,
	"day":    time.Hour * 24,
	"week":   time.Hour * 24 * 7,
}

// GetAdAnalytics retrieves analytics for an ad
func (h *Handler) GetAdAnalytics(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if uuid.Validate(id) != nil {
		logger.RequestLogger.Error(r, "Invalid ad ID: %v", id)
		apihelpers.ErrorResponse(r, w, http.StatusBadRequest, "Invalid ad ID")
		return
	}

	query := r.URL.Query()
	period := query.Get("period")
	if period == "" {
		period = "hour"
	} else if !slices.Contains([]string{"minute", "hour", "day", "week", "month"}, period) {
		logger.RequestLogger.Error(r, "Invalid period: %v", period)
		apihelpers.ErrorResponse(r, w, http.StatusBadRequest, "Invalid period")
		return
	}

	startDate := time.Now().AddDate(0, -1, 0) // for month period
	if period != "month" {
		startDate = time.Now().Add(-durationMap[period])
	}

	analytics, err := h.DB.GetAdAnalytics(id, startDate)
	if err != nil {
		if err == database.ErrNotFound {
			logger.RequestLogger.Error(r, "Ad not found")
			apihelpers.ErrorResponse(r, w, http.StatusNotFound, "Ad not found")
		} else {
			logger.RequestLogger.Error(r, "Error retrieving analytics: %v", err)
			apihelpers.ErrorResponse(r, w, http.StatusInternalServerError, "Error retrieving analytics")
		}
		return
	}

	analytics.Period = period
	if analytics.TotalClicks > 0 {
		analytics.AveragePlaybackTime = float64(analytics.TotalPlaybackTime) / float64(analytics.TotalClicks)
	}
	if analytics.TotalClicksInRange > 0 {
		analytics.AveragePlaybackTimeInRange = float64(analytics.TotalPlaybackTimeInRange) / float64(analytics.TotalClicksInRange)
	}

	apihelpers.SuccessResponse(r, w, http.StatusOK, analytics, "")
}

// GetAdsAnalytics retrieves analytics for all ads
func (h *Handler) GetAdsAnalytics(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	period := query.Get("period")
	if period == "" {
		period = "hour"
	} else if !slices.Contains([]string{"minute", "hour", "day", "week", "month"}, period) {
		logger.RequestLogger.Error(r, "Invalid period: %v", period)
		apihelpers.ErrorResponse(r, w, http.StatusBadRequest, "Invalid period")
		return
	}

	startDate := time.Now().AddDate(0, -1, 0) // for month period
	if period != "month" {
		startDate = time.Now().Add(-durationMap[period])
	}

	analytics, err := h.DB.GetAdsAnalytics(startDate)
	if err != nil {
		logger.RequestLogger.Error(r, "Error retrieving analytics: %v", err)
		apihelpers.ErrorResponse(r, w, http.StatusInternalServerError, "Error retrieving analytics")
		return
	}

	analytics.Period = period
	if analytics.TotalClicks > 0 {
		analytics.AveragePlaybackTime = float64(analytics.TotalPlaybackTime) / float64(analytics.TotalClicks)
	}
	if analytics.TotalClicksInRange > 0 {
		analytics.AveragePlaybackTimeInRange = float64(analytics.TotalPlaybackTimeInRange) / float64(analytics.TotalClicksInRange)
	}

	apihelpers.SuccessResponse(r, w, http.StatusOK, analytics, "")
}
