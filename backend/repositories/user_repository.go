package repositories

import (
	"context"
	"database/sql"
	"time"

	"github.com/todo-app/backend/database"
	"github.com/todo-app/backend/models"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByID(ctx context.Context, id uint) (*models.User, error)
	SearchUsers(ctx context.Context, query string, excludeUserID uint) ([]models.User, error)
	UpdatePassword(ctx context.Context, userID uint, hashedPassword string) error
}

type userRepository struct {
	q database.Querier
}

func NewUserRepository(q database.Querier) UserRepository {
	return &userRepository{q: q}
}

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	result, err := r.q.CreateUser(ctx, database.CreateUserParams{
		Name:     user.Name,
		Email:    user.Email,
		Password: user.Password,
	})
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	row, err := r.q.GetUserByID(ctx, uint32(id))
	if err != nil {
		return err
	}
	user.ID = uint(row.ID)
	user.CreatedAt = row.CreatedAt
	return nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	row, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return dbUserToModel(row), nil
}

func (r *userRepository) FindByID(ctx context.Context, id uint) (*models.User, error) {
	row, err := r.q.GetUserByID(ctx, uint32(id))
	if err != nil {
		return nil, err
	}
	return dbUserToModel(row), nil
}

func (r *userRepository) SearchUsers(ctx context.Context, query string, excludeUserID uint) ([]models.User, error) {
	term := "%" + query + "%"
	rows, err := r.q.SearchUsers(ctx, database.SearchUsersParams{
		ID:    uint32(excludeUserID),
		Name:  term,
		Email: term,
	})
	if err != nil {
		return nil, err
	}
	users := make([]models.User, 0, len(rows))
	for _, row := range rows {
		users = append(users, *dbUserToModel(row))
	}
	return users, nil
}

func (r *userRepository) UpdatePassword(ctx context.Context, userID uint, hashedPassword string) error {
	return r.q.UpdateUserPassword(ctx, database.UpdateUserPasswordParams{
		Password: hashedPassword,
		ID:       uint32(userID),
	})
}

// ── Conversion helpers (shared across all three repository files) ─────────────

// ErrNotFound re-exports sql.ErrNoRows for service-layer callers.
var ErrNotFound = sql.ErrNoRows

func dbUserToModel(u database.User) *models.User {
	return &models.User{
		ID:        uint(u.ID),
		Name:      u.Name,
		Email:     u.Email,
		Password:  u.Password,
		CreatedAt: u.CreatedAt,
	}
}

func toModelTodo(t database.Todo) models.Todo {
	m := models.Todo{
		ID:          uint(t.ID),
		Title:       t.Title,
		Description: t.Description,
		Completed:   t.Completed,
		DueDate:     NullTimeTo(t.DueDate),
		UserID:      uint(t.UserID),
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
	if t.ParentTodoID.Valid {
		p := uint(t.ParentTodoID.Int32)
		m.ParentTodoID = &p
	}
	return m
}

func toModelTodos(rows []database.Todo) []models.Todo {
	todos := make([]models.Todo, 0, len(rows))
	for _, r := range rows {
		todos = append(todos, toModelTodo(r))
	}
	return todos
}

// suppress unused warning for nowPtr (used implicitly via time.Now() calls)
var _ = time.Now
