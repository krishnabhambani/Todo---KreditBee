package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/todo-app/backend/controllers/dto"
	"github.com/todo-app/backend/mocks/github.com/todo-app/backend/services"
	"github.com/todo-app/backend/models"
)

func TestAuthController_Register_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockAuthService := services.NewMockAuthService(t)
	authController := NewAuthController(mockAuthService)

	router := gin.New()
	router.POST("/register", authController.Register)

	reqDto := dto.RegisterRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "Password123!",
	}

	mockUser := &models.User{
		ID:    1,
		Name:  "Test User",
		Email: "test@example.com",
	}

	mockAuthService.EXPECT().
		Register(mock.Anything, reqDto).
		Return(mockUser, nil)

	body, _ := json.Marshal(reqDto)
	req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.True(t, response["success"].(bool))
	assert.Equal(t, "User registered successfully", response["message"])
	
	data := response["data"].(map[string]interface{})
	assert.Equal(t, "test@example.com", data["email"])
}
