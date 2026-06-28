package services

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/todo-app/backend/apperrors"
	"github.com/todo-app/backend/controllers/dto"
	"github.com/todo-app/backend/mocks/github.com/todo-app/backend/repositories"
	"github.com/todo-app/backend/models"
	"github.com/todo-app/backend/utils"
)

func TestAuthService_Register_Success(t *testing.T) {
	mockRepo := repositories.NewMockUserRepository(t)
	authService := NewAuthService(mockRepo, "test_secret")

	req := dto.RegisterRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "Password123!",
	}

	// Mock FindByEmail to return nothing (user doesn't exist)
	mockRepo.EXPECT().
		FindByEmail(mock.Anything, "test@example.com").
		Return(nil, sql.ErrNoRows)

	// Mock Create to succeed
	mockRepo.EXPECT().
		Create(mock.Anything, mock.MatchedBy(func(u *models.User) bool {
			return u.Email == "test@example.com" && u.Name == "Test User"
		})).
		Return(nil)

	user, err := authService.Register(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Empty(t, user.Password, "Password hash should be cleared from response")
}

func TestAuthService_Register_EmailExists(t *testing.T) {
	mockRepo := repositories.NewMockUserRepository(t)
	authService := NewAuthService(mockRepo, "test_secret")

	req := dto.RegisterRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "Password123!",
	}

	// Mock FindByEmail to return an existing user
	mockRepo.EXPECT().
		FindByEmail(mock.Anything, "test@example.com").
		Return(&models.User{ID: 1, Email: "test@example.com"}, nil)

	user, err := authService.Register(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, user)
	appErr, ok := err.(*apperrors.AppError)
	assert.True(t, ok)
	assert.Equal(t, apperrors.ErrBadRequest, appErr.Code)
	assert.Equal(t, "user with this email already exists", appErr.Message)
}

func TestAuthService_Login_Success(t *testing.T) {
	mockRepo := repositories.NewMockUserRepository(t)
	authService := NewAuthService(mockRepo, "test_secret")

	req := dto.LoginRequest{
		Email:    "test@example.com",
		Password: "Password123!",
	}

	hashedPassword, _ := utils.HashPassword("Password123!")

	// Mock FindByEmail to return an existing user
	mockRepo.EXPECT().
		FindByEmail(mock.Anything, "test@example.com").
		Return(&models.User{ID: 1, Email: "test@example.com", Password: hashedPassword}, nil)

	token, user, err := authService.Login(context.Background(), req)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.NotNil(t, user)
	assert.Equal(t, uint(1), user.ID)
	assert.Empty(t, user.Password, "Password hash should be cleared from response")
}
