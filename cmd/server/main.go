package main

import (
	"cmp"
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JalajGoswami/video-ad-metrics/internal/database"
	"github.com/JalajGoswami/video-ad-metrics/internal/handlers"
)

func main() {
	port := cmp.Or(os.Getenv("PORT"), "5000")
	dbUrl := cmp.Or(
		os.Getenv("DATABASE_URL"),
		"postgres://postgres:postgres@postgres:5432/video-ad-metrics?sslmode=disable",
	)

	// Initialize database connection
	db, err := database.NewPostgresDB(dbUrl)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Setup database tables
	if err := db.Setup(); err != nil {
		log.Fatalf("Failed to setup database tables: %v", err)
	}

	// might be done using a cron job in production
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := db.ArchiveOldClicks(); err != nil {
					log.Printf("Failed to archive old clicks: %v", err)
				}
			}
		}
	}()

	ctx := context.Background()
	mux := http.NewServeMux()
	h := handlers.NewHandler(db)

	// Register routes
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, r.URL)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Ad management routes
	mux.HandleFunc("GET /api/ads", h.ListAds)
	mux.HandleFunc("POST /api/ads", h.CreateAd)
	mux.HandleFunc("GET /api/ads/{id}", h.GetAd)

	// Tracking routes
	mux.HandleFunc("POST /api/clicks", h.LogClick)

	// Analytics routes
	mux.HandleFunc("GET /api/analytics", h.GetAdsAnalytics)
	mux.HandleFunc("GET /api/analytics/{id}", h.GetAdAnalytics)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	// Start server in a separate goroutine
	go func() {
		log.Printf("⚡ Server listening on http://localhost:%s\n", port)
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Server shutting down...")
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server failed in graceful shutdown: %v", err)
	}
}
