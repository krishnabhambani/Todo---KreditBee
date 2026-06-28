package main

import (
	"context"

	"github.com/todo-app/backend/app"
	"github.com/todo-app/backend/config"
	"github.com/todo-app/backend/database"
	"github.com/todo-app/backend/logger"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := logger.NewLogger()
	log.Info(ctx, "starting Todo Application backend")

	cfg := config.LoadConfig()
	if cfg.JWT().Secret == config.JWTDefaultSecret() {
		log.Warn(ctx, "using default JWT secret — set JWT_SECRET env variable before deploying to production")
	}

	db := database.ConnectDatabase(ctx, cfg, log)
	application := app.NewApp(cfg, log, db)
	application.Run(ctx)
}
