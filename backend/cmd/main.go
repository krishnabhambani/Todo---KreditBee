package main

import (
	"context"

	"github.com/todo-app/backend/app"
	"github.com/todo-app/backend/config"
	"github.com/todo-app/backend/database"
	"github.com/todo-app/backend/logger"
)

func main() {
	log := logger.NewLogger()
	log.Info(context.Background(), "starting Todo Application backend")

	cfg := config.LoadConfig()
	if cfg.GetJWTSecret() == config.JWTDefaultSecret() {
		log.Warn(context.Background(), "using default JWT secret — set JWT_SECRET env variable before deploying to production")
	}

	db := database.ConnectDatabase(cfg, log)
	application := app.NewApp(cfg, log, db)
	application.Run()
}
