package app

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/todo-app/backend/config"
	"github.com/todo-app/backend/controllers"
	"github.com/todo-app/backend/database"
	"github.com/todo-app/backend/logger"
	"github.com/todo-app/backend/middleware"
	"github.com/todo-app/backend/repositories"
	"github.com/todo-app/backend/services"
)

// NewRouter wires all dependencies and registers all routes.
func NewRouter(cfg *config.Config, log logger.Logger, db *sql.DB) *gin.Engine {
	router := gin.New()

	// 404 and 405 handlers
	router.NoRoute(middleware.NotFoundHandler())

	// 405 Middleware — enable Gin to match routes but report method mismatch
	router.HandleMethodNotAllowed = true
	router.NoMethod(middleware.MethodNotAllowedHandler())

	// Global Middlewares — inject logger and jwt secret
	router.Use(middleware.LoggerMiddleware(log))
	router.Use(middleware.ErrorHandler(log))

	// CORS Middleware
	router.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		}
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Dependency Injection — build one Queries instance, share across all repos.
	queries := database.New(db)

	userRepo := repositories.NewUserRepository(queries)
	todoRepo := repositories.NewTodoRepository(queries)
	groupShareRepo := repositories.NewGroupShareRepository(queries)

	authService := services.NewAuthService(userRepo, cfg.GetJWTSecret())
	todoService := services.NewTodoService(todoRepo, groupShareRepo, userRepo)

	authController := controllers.NewAuthController(authService, log)
	groupController := controllers.NewGroupController(todoService)
	subtaskController := controllers.NewSubtaskController(todoService)
	shareController := controllers.NewShareController(todoService)

	// Auth middleware is constructed with the JWT secret from config
	authMW := middleware.AuthMiddleware(cfg.GetJWTSecret())

	// API Group
	api := router.Group("/api")
	{
		// Auth Routes (Public)
		auth := api.Group("/auth")
		{
			auth.POST("/register", authController.Register)
			auth.POST("/login", authController.Login)
		}

		// Auth Routes (Protected)
		authProtected := api.Group("/auth")
		authProtected.Use(authMW)
		{
			authProtected.PATCH("/password", authController.UpdatePassword)
		}

		// Group Routes (Protected)
		groups := api.Group("/groups")
		groups.Use(authMW)
		{
			groups.GET("", groupController.GetGroups)
			groups.GET("/:id", groupController.GetGroupByID)
			groups.POST("", groupController.CreateGroup)
			groups.PUT("/:id", groupController.UpdateGroup)
			groups.DELETE("/:id", groupController.DeleteGroup)

			// Subtasks under a group
			groups.GET("/:id/tasks", subtaskController.GetSubtasks)
			groups.POST("/:id/tasks", subtaskController.CreateSubtask)

			// Group sharing
			groups.POST("/:id/share", shareController.ShareGroup)
			groups.DELETE("/:id/share/:userId", shareController.RemoveShare)
			groups.PATCH("/:id/share/:userId/role", shareController.UpdateSharePermission)
			groups.GET("/:id/members", shareController.GetGroupMembers)
		}

		// Individual Subtask actions (Protected)
		tasks := api.Group("/tasks")
		tasks.Use(authMW)
		{
			tasks.PUT("/:id", subtaskController.UpdateSubtask)
			tasks.DELETE("/:id", subtaskController.DeleteSubtask)
			tasks.PATCH("/:id/complete", subtaskController.ToggleCompleteSubtask)
		}

		// Shared Groups (Protected)
		shared := api.Group("/shared-groups")
		shared.Use(authMW)
		{
			shared.GET("", shareController.GetSharedGroups)
		}

		// Users Search (Protected)
		users := api.Group("/users")
		users.Use(authMW)
		{
			users.GET("", shareController.SearchUsers)
		}
	}

	return router
}
