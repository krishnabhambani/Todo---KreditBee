package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/todo-app/backend/apperrors"
	"github.com/todo-app/backend/controllers/dto"
	"github.com/todo-app/backend/logger"
	"github.com/todo-app/backend/services"
)

type AuthController interface {
	Register(c *gin.Context)
	Login(c *gin.Context)
	UpdatePassword(c *gin.Context)
}

type authController struct {
	authService services.AuthService
	log         logger.Logger
}

// NewAuthController injects both the auth service and the structured logger.
func NewAuthController(service services.AuthService, log logger.Logger) AuthController {
	return &authController{authService: service, log: log}
}

// Register handles user registration
func (ctrl *authController) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.NewBadRequest("invalid request payload"))
		return
	}

	ctrl.log.Info(c.Request.Context(), "register attempt", logger.F("email", req.Email))

	user, err := ctrl.authService.Register(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}

	ctrl.log.Info(c.Request.Context(), "register success", logger.F("email", req.Email), logger.F("userID", user.ID))
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "User registered successfully",
		"data":    dto.MapUser(user),
	})
}

// Login handles user authentication
func (ctrl *authController) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.NewBadRequest("email and password are required"))
		return
	}

	ctrl.log.Info(c.Request.Context(), "login attempt", logger.F("email", req.Email))

	token, user, err := ctrl.authService.Login(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}

	ctrl.log.Info(c.Request.Context(), "login success", logger.F("email", req.Email))
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Login successful",
		"data": dto.LoginResponse{
			Token: token,
			User:  *dto.MapUser(user),
		},
	})
}

func (ctrl *authController) UpdatePassword(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.Error(apperrors.NewUnauthorized("unauthorized"))
		return
	}

	var req dto.UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.NewBadRequest("current_password and new_password are required"))
		return
	}

	ctrl.log.Info(c.Request.Context(), "update password attempt", logger.F("userID", userID))

	if err := ctrl.authService.UpdatePassword(c.Request.Context(), userID.(uint), req); err != nil {
		c.Error(err)
		return
	}

	ctrl.log.Info(c.Request.Context(), "update password success", logger.F("userID", userID))
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Password updated successfully",
		"data":    nil,
	})
}
