package repositories

import (
	"context"
	"sort"
	"time"

	"github.com/todo-app/backend/database"
	"github.com/todo-app/backend/models"
)

type TodoRepository interface {
	Create(ctx context.Context, todo *models.Todo) error
	FindAllGroupsByUserID(ctx context.Context, userID uint, search string, status string, sortParam string) ([]models.Todo, error)
	FindGroupByID(ctx context.Context, id uint, userID uint) (*models.Todo, error)
	FindSubtasksByGroupID(ctx context.Context, groupID uint, userID uint) ([]models.Todo, error)
	FindByID(ctx context.Context, id uint) (*models.Todo, error)
	Update(ctx context.Context, todo *models.Todo) error
	Delete(ctx context.Context, id uint) error
}

type todoRepository struct {
	q *database.Queries
}

func NewTodoRepository(q *database.Queries) TodoRepository {
	return &todoRepository{q: q}
}

func (r *todoRepository) Create(ctx context.Context, todo *models.Todo) error {
	result, err := r.q.CreateTodo(ctx, database.CreateTodoParams{
		Title:        todo.Title,
		Description:  todo.Description,
		Completed:    todo.Completed,
		DueDate:      database.ToNullTime(todo.DueDate),
		UserID:       uint32(todo.UserID),
		ParentTodoID: database.ToNullInt32(todo.ParentTodoID),
	})
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	todo.ID = uint(id)
	todo.CreatedAt = time.Now()
	todo.UpdatedAt = time.Now()
	return nil
}

func (r *todoRepository) FindAllGroupsByUserID(ctx context.Context, userID uint, search, status, sortParam string) ([]models.Todo, error) {
	searchTerm := ""
	if search != "" {
		searchTerm = "%" + search + "%"
	}

	rows, err := r.q.GetGroupsByUserID(ctx, database.GetGroupsByUserIDParams{
		UserID: uint32(userID),
		Search: searchTerm,
	})
	if err != nil {
		return nil, err
	}

	// Fetch owner once — all owned groups share the same user.
	var owner *models.User
	if len(rows) > 0 {
		if u, err := r.q.GetUserByID(ctx, uint32(userID)); err == nil {
			owner = dbUserToModel(u)
		}
	}

	now := time.Now()
	groups := make([]models.Todo, 0, len(rows))
	for _, row := range rows {
		group := toModelTodo(row)

		subtaskRows, _ := r.q.GetSubtasksByParentID(ctx, row.ID)
		group.Subtasks = toModelTodos(subtaskRows)

		if owner != nil {
			ownerCopy := *owner
			group.Owner = &ownerCopy
		}

		if !applyStatusFilter(&group, status, now) {
			continue
		}
		groups = append(groups, group)
	}

	applySortParam(groups, sortParam)
	return groups, nil
}

func (r *todoRepository) FindGroupByID(ctx context.Context, id uint, userID uint) (*models.Todo, error) {
	row, err := r.q.GetGroupByID(ctx, uint32(id))
	if err != nil {
		return nil, err
	}
	group := toModelTodo(row)

	subtaskRows, _ := r.q.GetSubtasksByParentID(ctx, row.ID)
	group.Subtasks = toModelTodos(subtaskRows)

	if u, err := r.q.GetUserByID(ctx, row.UserID); err == nil {
		group.Owner = dbUserToModel(u)
	}
	return &group, nil
}

func (r *todoRepository) FindSubtasksByGroupID(ctx context.Context, groupID uint, userID uint) ([]models.Todo, error) {
	rows, err := r.q.GetSubtasksByParentAndUser(ctx, uint32(groupID), uint32(userID))
	if err != nil {
		return nil, err
	}
	return toModelTodos(rows), nil
}

func (r *todoRepository) FindByID(ctx context.Context, id uint) (*models.Todo, error) {
	row, err := r.q.GetTodoByID(ctx, uint32(id))
	if err != nil {
		return nil, err
	}
	t := toModelTodo(row)
	return &t, nil
}

func (r *todoRepository) Update(ctx context.Context, todo *models.Todo) error {
	return r.q.UpdateTodo(ctx, database.UpdateTodoParams{
		Title:       todo.Title,
		Description: todo.Description,
		Completed:   todo.Completed,
		DueDate:     database.ToNullTime(todo.DueDate),
		ID:          uint32(todo.ID),
	})
}

func (r *todoRepository) Delete(ctx context.Context, id uint) error {
	return r.q.DeleteTodo(ctx, uint32(id))
}

// ── Filter/sort helpers ───────────────────────────────────────────────────────

func applyStatusFilter(g *models.Todo, status string, now time.Time) bool {
	switch status {
	case "overdue":
		return g.DueDate != nil && g.DueDate.Before(now) && !g.Completed
	case "due-today":
		if g.DueDate == nil {
			return false
		}
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		return !g.DueDate.Before(start) && g.DueDate.Before(start.Add(24*time.Hour))
	case "due-this-week":
		if g.DueDate == nil {
			return false
		}
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		return !g.DueDate.Before(start) && g.DueDate.Before(start.Add(7*24*time.Hour))
	case "completed":
		return g.Completed
	case "active":
		return !g.Completed
	default:
		return true
	}
}

func applySortParam(groups []models.Todo, sortParam string) {
	switch sortParam {
	case "deadline":
		sort.SliceStable(groups, func(i, j int) bool {
			if groups[i].DueDate == nil {
				return false
			}
			if groups[j].DueDate == nil {
				return true
			}
			return groups[i].DueDate.Before(*groups[j].DueDate)
		})
	case "deadline-desc":
		sort.SliceStable(groups, func(i, j int) bool {
			if groups[i].DueDate == nil {
				return false
			}
			if groups[j].DueDate == nil {
				return true
			}
			return groups[i].DueDate.After(*groups[j].DueDate)
		})
	case "updated":
		sort.SliceStable(groups, func(i, j int) bool {
			return groups[i].UpdatedAt.After(groups[j].UpdatedAt)
		})
	}
}
