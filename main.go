package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eraasoft/ai-camp-server/internal/config"
	"github.com/eraasoft/ai-camp-server/internal/database"
	"github.com/eraasoft/ai-camp-server/internal/handlers"
	"github.com/eraasoft/ai-camp-server/internal/middleware"
)

func main() {
	log.Println("==================================================")
	log.Println(" EraaSoft AI Camp Registration Server (Go + Mongo)")
	log.Println("==================================================")

	// 1. Load runtime configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("[FATAL] Configuration error: %v", err)
	}

	// 2. Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db, err := database.Connect(ctx, cfg.MongoURI, cfg.DBName)
	if err != nil {
		log.Fatalf("[FATAL] Could not connect to MongoDB: %v", err)
	}
	defer db.Disconnect(context.Background())

	// 3. Ensure database indexes
	idxCtx, idxCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer idxCancel()
	if err := db.EnsureIndexes(idxCtx); err != nil {
		log.Printf("[WARN] Failed to ensure database indexes: %v", err)
	}

	// 4. Initialize handlers
	regHandler := handlers.NewRegistrationHandler(db)
	adminHandler := handlers.NewAdminHandler(db)

	// 5. Setup HTTP routes
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"database":  "connected",
		})
	})

	// Public registration endpoint
	mux.HandleFunc("/api/register", regHandler.Register)

	// Admin protected routes
	adminAuth := middleware.AdminAuth(cfg.AdminSecret)
	mux.Handle("/api/admin/export", adminAuth(http.HandlerFunc(adminHandler.ExportCSV)))
	mux.Handle("/api/admin/stats", adminAuth(http.HandlerFunc(adminHandler.GetStats)))
	mux.Handle("/api/admin/list", adminAuth(http.HandlerFunc(adminHandler.ListRegistrations)))

	// 6. Apply global CORS middleware
	handlerWithCORS := middleware.CORS(cfg.AllowedOrigins)(mux)

	// 7. Configure HTTP server
	addr := fmt.Sprintf(":%s", cfg.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      handlerWithCORS,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 45 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 8. Run server in background goroutine
	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("[HTTP] Server is listening on http://localhost%s", addr)
		log.Printf("[HTTP] Allowed CORS origins: %v", cfg.AllowedOrigins)
		log.Printf("[HTTP] Admin endpoints protected with secret key")
		serverErrors <- srv.ListenAndServe()
	}()

	// 9. Listen for OS interrupt signals for graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("[FATAL] Server failed: %v", err)
		}

	case sig := <-shutdown:
		log.Printf("[SHUTDOWN] Received signal %v. Gracefully shutting down...", sig)

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("[WARN] Forced server shutdown: %v", err)
			_ = srv.Close()
		}
		log.Println("[SHUTDOWN] Server stopped cleanly")
	}
}
