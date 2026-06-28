package app

import (
	"context"
	"database/sql"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/todo-app/backend/config"
	"github.com/todo-app/backend/database"
	"github.com/todo-app/backend/logger"
)

// Container defines the dependencies exposed by the application bootstrapper.
type Container interface {
	Config() config.Config
	Logger() logger.Logger
	DB() *sql.DB
}

// App is the unified application bootstrapper that owns configuration,
// infrastructure dependencies, and the HTTP router.
type App struct {
	cfg    config.Config
	log    logger.Logger
	db     *sql.DB
	router http.Handler
	srv    *http.Server
}

// Config returns the application configuration.
func (a *App) Config() config.Config { return a.cfg }

// Logger returns the application logger.
func (a *App) Logger() logger.Logger { return a.log }

// DB returns the application database connection.
func (a *App) DB() *sql.DB { return a.db }

// NewApp constructs the application with injected config, logger, and DB.
func NewApp(cfg config.Config, log logger.Logger, db *sql.DB) *App {
	a := &App{
		cfg: cfg,
		log: log,
		db:  db,
	}

	a.router = NewRouter(a)
	port := cfg.Server().Port
	a.srv = &http.Server{
		Addr:    ":" + port,
		Handler: a.router,
	}

	return a
}

// Start starts the application services in the background.
func (a *App) Start(ctx context.Context) {
	a.srv.BaseContext = func(l net.Listener) context.Context {
		return ctx
	}

	go func() {
		a.log.Info(ctx, "server listening", logger.F("port", a.cfg.Server().Port))
		if err := a.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.log.Fatal(ctx, "server failed to start", logger.F("error", err))
		}
	}()
}

// Stop gracefully shuts down the application services.
func (a *App) Stop(ctx context.Context) {
	a.log.Info(ctx, "shutting down server...")

	if err := a.srv.Shutdown(ctx); err != nil {
		a.log.Fatal(ctx, "server forced to shutdown", logger.F("error", err))
	}

	if err := database.CloseDatabase(a.db); err != nil {
		a.log.Error(ctx, "error closing database connection", logger.F("error", err))
	} else {
		a.log.Info(ctx, "database connection closed cleanly")
	}

	a.log.Info(ctx, "server exiting")
}

// Run orchestrates the application lifecycle rather than owning shutdown logic.
func (a *App) Run(ctx context.Context) {
	a.Start(ctx)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		// OS signal received
	case <-ctx.Done():
		// Context cancelled externally
	}

	// Use context.Background() as the base for the shutdown timeout
	// because the original ctx might already be cancelled.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	a.Stop(shutdownCtx)
}
