package app

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/todo-app/backend/config"
	"github.com/todo-app/backend/database"
	"github.com/todo-app/backend/logger"
)

// App is the unified application bootstrapper that owns configuration,
// infrastructure dependencies, and the HTTP router.
type App struct {
	cfg    *config.Config
	log    logger.Logger
	db     *sql.DB
	router http.Handler
}

// NewApp constructs the application with injected config, logger, and DB.
func NewApp(cfg *config.Config, log logger.Logger, db *sql.DB) *App {
	return &App{
		cfg:    cfg,
		log:    log,
		db:     db,
		router: NewRouter(cfg, log, db),
	}
}

// Run starts the HTTP server and blocks until shutdown.
func (a *App) Run() {
	port := a.cfg.GetPort()
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: a.router,
	}

	go func() {
		a.log.Info(context.Background(), "server listening", logger.F("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.log.Fatal(context.Background(), "server failed to start", logger.F("error", err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	a.log.Info(context.Background(), "shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		a.log.Fatal(context.Background(), "server forced to shutdown", logger.F("error", err))
	}

	if err := database.CloseDatabase(a.db); err != nil {
		a.log.Error(context.Background(), "error closing database connection", logger.F("error", err))
	} else {
		a.log.Info(context.Background(), "database connection closed cleanly")
	}

	a.log.Info(context.Background(), "server exiting")
}
