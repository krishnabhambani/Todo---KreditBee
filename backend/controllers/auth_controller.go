package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/todo-app/backend/apperrors"
	"github.com/todo-app/backend/controllers/dto"
	"github.com/todo-app/backend/response"
	"github.com/todo-app/backend/services"
)

type AuthController interface {
	Register(c *gin.Context)
	Login(c *gin.Context)
	UpdatePassword(c *gin.Context)
}

type authController struct {
	authService services.AuthService
}

// NewAuthController injects the auth service.
func NewAuthController(service services.AuthService) AuthController {
	return &authController{authService: service}
}

// Register handles user registration
func (ctrl *authController) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, err)
		return
	}

	user, err := ctrl.authService.Register(c.Request.Context(), req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "User registered successfully", dto.MapUser(user))
}

// Login handles user authentication
func (ctrl *authController) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, err)
		return
	}

	token, user, err := ctrl.authService.Login(c.Request.Context(), req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Login successful", dto.LoginResponse{
		Token: token,
		User:  *dto.MapUser(user),
	})
}

func (ctrl *authController) UpdatePassword(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		response.HandleError(c, apperrors.NewUnauthorized("unauthorized"))
		return
	}

	var req dto.UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, err)
		return
	}

	if err := ctrl.authService.UpdatePassword(c.Request.Context(), userID.(uint), req); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Password updated successfully", nil)
}
