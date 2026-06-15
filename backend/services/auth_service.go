package services

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/todo-app/backend/models"
	"github.com/todo-app/backend/repositories"
	"github.com/todo-app/backend/utils"
	"gorm.io/gorm"
)
//Blueprint
type AuthService interface {
	Register(ctx context.Context, name, email, password string) (*models.User, error)
	Login(ctx context.Context, email, password string) (string, *models.User, error)
}

//implementation
type authService struct {
	userRepo repositories.UserRepository
}

// Used for testing
func NewAuthService(repo repositories.UserRepository) AuthService {
	return &authService{userRepo: repo}
}

func (s *authService) Register(ctx context.Context, name, email, password string) (*models.User, error) {
	// Normalize email: trim spaces and lowercase
	email = strings.ToLower(strings.TrimSpace(email))
	name = strings.TrimSpace(name)

	// Validate email format (case-insensitive, supporting modern TLDs)
	emailRegex := regexp.MustCompile(`(?i)^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return nil, errors.New("invalid email format")
	}

	// Validate password strength: min 8 chars, 1 upper, 1 lower, 1 digit, 1 special
	if len(password) < 8 {
		return nil, errors.New("password must be at least 8 characters long")
	}
	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, char := range password {
		switch {
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= '0' && char <= '9':
			hasDigit = true
		default:
			hasSpecial = true
		}
	}
	if !hasUpper || !hasLower || !hasDigit || !hasSpecial {
		return nil, errors.New("password must contain at least one uppercase letter, one lowercase letter, one number, and one special character")
	}

	// Check if user already exists, handling database errors safely
	existingUser, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("failed to query database for existing user")
	}
	if existingUser != nil {
		return nil, errors.New("email is already registered")
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, errors.New("failed to process password security")
	}

	user := &models.User{
		Name:     name,
		Email:    email,
		Password: hashedPassword,
	}

	err = s.userRepo.Create(ctx, user)
	if err != nil {
		// Mask raw database schema errors (e.g. duplicate key, database locks)
		if strings.Contains(err.Error(), "Duplicate entry") {
			return nil, errors.New("email is already registered")
		}
		return nil, errors.New("failed to create user account due to a database issue")
	}

	// Clear password hash before returning so it never leaks via logging or non-JSON serialization
	user.Password = ""
	return user, nil
}

func (s *authService) Login(ctx context.Context, email, password string) (string, *models.User, error) {
	// Normalize email
	email = strings.ToLower(strings.TrimSpace(email))

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil, errors.New("invalid email or password")
		}
		return "", nil, errors.New("database connectivity error")
	}

	// Check password
	if !utils.CheckPasswordHash(password, user.Password) {
		return "", nil, errors.New("invalid email or password")
	}

	// Generate JWT
	token, err := utils.GenerateJWT(user.ID)
	if err != nil {
		return "", nil, errors.New("failed to generate access session")
	}

	// Clear password hash before returning so it never leaks via logging or non-JSON serialization
	user.Password = ""
	return token, user, nil
}
