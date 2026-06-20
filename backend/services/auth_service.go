package services

import (
	"context"
	"database/sql"
	"regexp"
	"strings"

	"github.com/todo-app/backend/apperrors"
	"github.com/todo-app/backend/controllers/dto"
	"github.com/todo-app/backend/models"
	"github.com/todo-app/backend/repositories"
	"github.com/todo-app/backend/utils"
)

// Blueprint
type AuthService interface {
	Register(ctx context.Context, req dto.RegisterRequest) (*models.User, error)
	Login(ctx context.Context, req dto.LoginRequest) (string, *models.User, error)
	UpdatePassword(ctx context.Context, userID uint, req dto.UpdatePasswordRequest) error
}

// implementation
type authService struct {
	userRepo  repositories.UserRepository
	jwtSecret string // injected — no global config dependency
}

// NewAuthService constructs an AuthService.
// jwtSecret is passed explicitly so this package has no import of config.
func NewAuthService(repo repositories.UserRepository, jwtSecret string) AuthService {
	return &authService{userRepo: repo, jwtSecret: jwtSecret}
}

// Register handles new user registration
// Register handles new user registration
func (s *authService) Register(ctx context.Context, req dto.RegisterRequest) (*models.User, error) {
    // 1. Normalize input parameters strings
    email := strings.ToLower(strings.TrimSpace(req.Email))
    name := strings.TrimSpace(req.Name)

    // 2. Call our newly separated validation helper block
    if err := s.validateRegisterInput(ctx, email, name, req.Password); err != nil {
        return nil, err
    }

    // 3. Hash password
    hashedPassword, err := utils.HashPassword(req.Password)
    if err != nil {
        return nil, apperrors.NewInternal(err, "failed to process password security")
    }

    user := &models.User{
        Name:     name,
        Email:    email,
        Password: hashedPassword,
    }

    // 4. Save to Database
    err = s.userRepo.Create(ctx, user)
    if err != nil {
        return nil, apperrors.NewInternal(err, "failed to create user")
    }

    // Clear password hash before returning so it never leaks via logging or non-JSON serialization
    user.Password = ""
    return user, nil
}

// validateRegisterInput is a private helper that encapsulates format checks - helper
func (s *authService) validateRegisterInput(ctx context.Context, email, name, password string) error {
    // 1. Validate email format (case-insensitive, supporting modern TLDs)
    emailRegex := regexp.MustCompile(`(?i)^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)
    if !emailRegex.MatchString(email) {
        return apperrors.NewBadRequest("invalid email format")
    }

    // 2. Validate password length constraint
    if len(password) < 8 {
        return apperrors.NewBadRequest("password must be at least 8 characters long")
    }

    // 3. Validate password complexity criteria
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
        return apperrors.NewBadRequest("password must contain at least one uppercase letter, one lowercase letter, one number, and one special character")
    }

    // 4. Validate user uniqueness across database records
    existingUser, err := s.userRepo.FindByEmail(ctx, email)
    if err != nil && err != sql.ErrNoRows {
        return apperrors.NewInternal(err, "database connectivity error")
    }
    if existingUser != nil {
        return apperrors.NewBadRequest("user with this email already exists")
    }

    return nil
}

// Login authenticates a user and returns a JWT
func (s *authService) Login(ctx context.Context, req dto.LoginRequest) (string, *models.User, error) {
	// Normalize email
	email := strings.ToLower(strings.TrimSpace(req.Email))

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil, apperrors.NewUnauthorized("invalid email or password")
		}
		return "", nil, apperrors.NewInternal(err, "database connectivity error")
	}

	// Check password
	if !utils.CheckPasswordHash(req.Password, user.Password) {
		return "", nil, apperrors.NewUnauthorized("invalid email or password")
	}

	// Generate JWT — secret injected, no global access
	token, err := utils.GenerateJWT(user.ID, s.jwtSecret)
	if err != nil {
		return "", nil, apperrors.NewInternal(err, "failed to generate authentication token")
	}

	// Clear password hash before returning so it never leaks via logging or non-JSON serialization
	user.Password = ""
	return token, user, nil
}

func (s *authService) UpdatePassword(ctx context.Context, userID uint, req dto.UpdatePasswordRequest) error {
	// Fetch user to verify current password
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return apperrors.NewNotFound("user not found")
	}

	// Verify current password
	if !utils.CheckPasswordHash(req.CurrentPassword, user.Password) {
		return apperrors.NewBadRequest("current password is incorrect")
	}

	// Validate new password strength
	if len(req.NewPassword) < 8 {
		return apperrors.NewBadRequest("new password must be at least 8 characters long")
	}
	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, ch := range req.NewPassword {
		switch {
		case ch >= 'a' && ch <= 'z':
			hasLower = true
		case ch >= 'A' && ch <= 'Z':
			hasUpper = true
		case ch >= '0' && ch <= '9':
			hasDigit = true
		default:
			hasSpecial = true
		}
	}
	if !hasUpper || !hasLower || !hasDigit || !hasSpecial {
		return apperrors.NewBadRequest("new password must contain at least one uppercase letter, one lowercase letter, one number, and one special character")
	}

	// Hash and store
	hashed, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return apperrors.NewInternal(err, "failed to process new password")
	}

	if err := s.userRepo.UpdatePassword(ctx, userID, hashed); err != nil {
		return apperrors.NewInternal(err, "failed to save new password")
	}

	return nil
}
