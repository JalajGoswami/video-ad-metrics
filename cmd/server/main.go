package main

import (
	"cmp"
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	apihelpers "github.com/JalajGoswami/video-ad-metrics/internal/api-helpers"
	"github.com/JalajGoswami/video-ad-metrics/internal/database"
	"github.com/JalajGoswami/video-ad-metrics/internal/handlers"
	"github.com/JalajGoswami/video-ad-metrics/internal/logger"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	port := cmp.Or(os.Getenv("PORT"), "5000")
	dbUrl := cmp.Or(
		os.Getenv("DATABASE_URL"),
		"postgres://postgres:postgres@postgres:5432/video_ad_metrics?sslmode=disable",
	)
	// Initialize database connection
	db, err := database.NewPostgresDB(dbUrl)
	if err != nil {
		logger.FatalLog("Failed to connect to database: %v", err)
	}
	defer db.Close()

	logger.SetupRequestLogger()

	// Setup database tables
	if err := db.Setup(); err != nil {
		logger.FatalLog("Failed to setup database tables: %v", err)
	}

	// might be done using a cron job in production
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := db.ArchiveOldClicks(); err != nil {
					logger.ErrorLog("Failed to archive old clicks: %v", err)
				}
			}
		}
	}()

	ctx := context.Background()
	mux := http.NewServeMux()
	h := handlers.NewHandler(db)

	// Register routes
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		if db.Ping() != nil {
			logger.RequestLogger.Error(r, "Database connection failed: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Database connection failed"))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		}
	})

	// Ad management routes
	mux.HandleFunc("GET /ads", h.ListAds)
	mux.HandleFunc("POST /ads", h.CreateAd)
	mux.HandleFunc("GET /ads/{id}", h.GetAd)

	// Tracking routes
	mux.HandleFunc("POST /ads/clicks", h.LogClick)

	// Analytics routes
	mux.HandleFunc("GET /ads/analytics", h.GetAdsAnalytics)
	mux.HandleFunc("GET /ads/analytics/{id}", h.GetAdAnalytics)

	handler := logger.RequestLogger.LoggingMiddleware(mux)
	handler = apihelpers.TraceMiddleware(handler)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}

	// Start server in a separate goroutine
	go func() {
		logger.LogColored(logger.ColorGreen, "⚡ Server listening on http://localhost:%s\n", port)
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.FatalLog("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.LogColored(logger.ColorYellow, "Server shutting down...")
	if err := server.Shutdown(ctx); err != nil {
		logger.FatalLog("Server failed in graceful shutdown: %v", err)
	}
}
