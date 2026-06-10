package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/todo-app/backend/controllers"
	"github.com/todo-app/backend/database"
	"github.com/todo-app/backend/middleware"
	"github.com/todo-app/backend/repositories"
	"github.com/todo-app/backend/services"
)

func SetupRouter() *gin.Engine {
	router := gin.New()

	// Global Middlewares
	router.Use(middleware.LoggerMiddleware())
	router.Use(middleware.ErrorRecoveryMiddleware())

	// CORS Middleware (Default)
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

	// Dependency Injection
	userRepo := repositories.NewUserRepository(database.DB)
	todoRepo := repositories.NewTodoRepository(database.DB)
	groupShareRepo := repositories.NewGroupShareRepository(database.DB)

	authService := services.NewAuthService(userRepo)
	todoService := services.NewTodoService(todoRepo, groupShareRepo, userRepo)

	authController := controllers.NewAuthController(authService)
	groupController := controllers.NewGroupController(todoService)
	subtaskController := controllers.NewSubtaskController(todoService)
	shareController := controllers.NewShareController(todoService)

	// API Group
	api := router.Group("/api")
	{
		// Auth Routes (Public)
		auth := api.Group("/auth")
		{
			auth.POST("/register", authController.Register)
			auth.POST("/login", authController.Login)
		}

		// Group Routes (Protected)
		groups := api.Group("/groups")
		groups.Use(middleware.AuthMiddleware())
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
			groups.GET("/:id/members", shareController.GetGroupMembers)
		}

		// Individual Subtask actions (Protected)
		tasks := api.Group("/tasks")
		tasks.Use(middleware.AuthMiddleware())
		{
			tasks.PUT("/:id", subtaskController.UpdateSubtask)
			tasks.DELETE("/:id", subtaskController.DeleteSubtask)
			tasks.PATCH("/:id/complete", subtaskController.ToggleCompleteSubtask)
		}

		// Shared Groups (Protected)
		shared := api.Group("/shared-groups")
		shared.Use(middleware.AuthMiddleware())
		{
			shared.GET("", shareController.GetSharedGroups)
		}

		// Users Search (Protected)
		users := api.Group("/users")
		users.Use(middleware.AuthMiddleware())
		{
			users.GET("", shareController.SearchUsers)
		}
	}

	return router
}
