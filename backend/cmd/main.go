package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/todo-app/backend/config"
	"github.com/todo-app/backend/database"
	"github.com/todo-app/backend/logger"
	"github.com/todo-app/backend/routes"
)

func main() {
	// 1. Bootstrap logger first — everything else logs through it
	log := logger.NewLogger()

	log.Info("starting Todo Application backend")

	// 2. Load configuration — returns interface, no global state
	cfg := config.LoadConfig()

	// Warn if running with the insecure default JWT secret
	if cfg.GetJWTSecret() == config.DefaultJWTSecret() {
		log.Warn("using default JWT secret — set JWT_SECRET env variable before deploying to production")
	}

	// 3. Connect and migrate database
	db := database.ConnectDatabase(cfg, log)

	// 4. Setup router with all injected dependencies (pass DB explicitly)
	router := routes.SetupRouter(cfg, log, db)

	port := cfg.GetPort()
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// 5. Start server in a goroutine
	go func() {
		log.Info("server listening", logger.F("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("server failed to start", logger.F("error", err))
		}
	}()

	// 6. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("shutting down server...")

	// 10 seconds timeout for in-flight requests
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("server forced to shutdown", logger.F("error", err))
	}

	if err := database.CloseDatabase(db); err != nil {
		log.Error("error closing database connection", logger.F("error", err))
	} else {
		log.Info("database connection closed cleanly")
	}

	log.Info("server exiting")
}
