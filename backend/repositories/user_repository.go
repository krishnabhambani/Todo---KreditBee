package repositories

import (
	"context"

	"github.com/todo-app/backend/models"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByID(ctx context.Context, id uint) (*models.User, error)
	SearchUsers(ctx context.Context, query string, excludeUserID uint) ([]models.User, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByID(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) SearchUsers(ctx context.Context, query string, excludeUserID uint) ([]models.User, error) {
	var users []models.User
	searchTerm := "%" + query + "%"
	err := r.db.WithContext(ctx).
		Where("id != ? AND (name LIKE ? OR email LIKE ?)", excludeUserID, searchTerm, searchTerm).
		Limit(20).
		Find(&users).Error
	return users, err
}
