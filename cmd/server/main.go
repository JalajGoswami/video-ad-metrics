package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/JalajGoswami/video-ad-metrics/internal/database"
	"github.com/JalajGoswami/video-ad-metrics/internal/handlers"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	// Get postgres connection string from environment variable or use default
	pgConnString := os.Getenv("DATABASE_URL")
	if pgConnString == "" {
		pgConnString = "postgres://postgres:postgres@localhost:5432/video_ad_metrics?sslmode=disable"
	}

	// Initialize database connection
	db, err := database.NewPostgresDB(pgConnString)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Setup database tables
	if err := db.Setup(); err != nil {
		log.Fatalf("Failed to setup database tables: %v", err)
	}

	// Setup periodic archiving of old clicks (to be done in a real production environment)
	// For this example, we'll just run it once at startup
	if err := db.ArchiveOldClicks(); err != nil {
		log.Printf("Warning: Failed to archive old clicks: %v", err)
	}

	// In production, you would set up a background job for this
	// go func() {
	//     ticker := time.NewTicker(24 * time.Hour)
	//     defer ticker.Stop()
	//     for {
	//         select {
	//         case <-ticker.C:
	//             if err := db.ArchiveOldClicks(); err != nil {
	//                 log.Printf("Failed to archive old clicks: %v", err)
	//             }
	//         }
	//     }
	// }()

	ctx := context.Background()
	mux := http.NewServeMux()
	h := handlers.NewHandler(db)

	// Register routes
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, r.URL)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Ad management routes
	mux.HandleFunc("GET /api/ads", h.ListAds)
	mux.HandleFunc("POST /api/ads", h.CreateAd)
	mux.HandleFunc("GET /api/ads/{id}", h.GetAd)

	// Tracking routes
	mux.HandleFunc("/api/clicks", h.LogClick)

	// Analytics routes
	mux.HandleFunc("/api/analytics", h.GetAdAnalytics)

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
