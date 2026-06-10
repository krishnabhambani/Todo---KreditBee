package repositories

import (
	"context"
	"time"

	"github.com/todo-app/backend/models"
	"gorm.io/gorm"
)

type TodoRepository interface {
	Create(ctx context.Context, todo *models.Todo) error
	FindAllGroupsByUserID(ctx context.Context, userID uint, search string, status string, sort string) ([]models.Todo, error)
	FindGroupByID(ctx context.Context, id uint, userID uint) (*models.Todo, error)
	FindSubtasksByGroupID(ctx context.Context, groupID uint, userID uint) ([]models.Todo, error)
	FindByID(ctx context.Context, id uint) (*models.Todo, error)
	Update(ctx context.Context, todo *models.Todo) error
	Delete(ctx context.Context, id uint) error
}

type todoRepository struct {
	db *gorm.DB
}

func NewTodoRepository(db *gorm.DB) TodoRepository {
	return &todoRepository{db: db}
}

func (r *todoRepository) Create(ctx context.Context, todo *models.Todo) error {
	return r.db.WithContext(ctx).Create(todo).Error
}

func (r *todoRepository) FindAllGroupsByUserID(ctx context.Context, userID uint, search string, status string, sort string) ([]models.Todo, error) {
	var groups []models.Todo
	query := r.db.WithContext(ctx).Model(&models.Todo{}).
		Preload("Subtasks").
		Preload("Owner").
		Where("user_id = ? AND parent_todo_id IS NULL", userID)

	if search != "" {
		searchTerm := "%" + search + "%"
		query = query.Where("title LIKE ? OR description LIKE ?", searchTerm, searchTerm)
	}

	// Status filters
	now := time.Now()
	switch status {
	case "overdue":
		query = query.Where("due_date IS NOT NULL AND due_date < ? AND completed = ?", now, false)
	case "due-today":
		todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		todayEnd := todayStart.Add(24 * time.Hour)
		query = query.Where("due_date >= ? AND due_date < ?", todayStart, todayEnd)
	case "due-this-week":
		todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		weekEnd := todayStart.Add(7 * 24 * time.Hour)
		query = query.Where("due_date >= ? AND due_date < ?", todayStart, weekEnd)
	case "completed":
		query = query.Where("completed = ?", true)
	case "active":
		query = query.Where("completed = ?", false)
	}

	// Sorting
	switch sort {
	case "deadline":
		query = query.Order("due_date IS NULL, due_date ASC")
	case "deadline-desc":
		query = query.Order("due_date IS NULL, due_date DESC")
	case "updated":
		query = query.Order("updated_at DESC")
	default:
		query = query.Order("created_at DESC")
	}

	err := query.Find(&groups).Error
	return groups, err
}

func (r *todoRepository) FindGroupByID(ctx context.Context, id uint, userID uint) (*models.Todo, error) {
	var group models.Todo
	err := r.db.WithContext(ctx).
		Preload("Subtasks").
		Preload("Owner").
		Where("id = ? AND parent_todo_id IS NULL", id).
		First(&group).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *todoRepository) FindSubtasksByGroupID(ctx context.Context, groupID uint, userID uint) ([]models.Todo, error) {
	var subtasks []models.Todo
	err := r.db.WithContext(ctx).
		Where("parent_todo_id = ? AND user_id = ?", groupID, userID).
		Order("created_at asc").
		Find(&subtasks).Error
	return subtasks, err
}

func (r *todoRepository) FindByID(ctx context.Context, id uint) (*models.Todo, error) {
	var todo models.Todo
	err := r.db.WithContext(ctx).First(&todo, id).Error
	if err != nil {
		return nil, err
	}
	return &todo, nil
}

func (r *todoRepository) Update(ctx context.Context, todo *models.Todo) error {
	return r.db.WithContext(ctx).Save(todo).Error
}

func (r *todoRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Todo{}, id).Error
}
