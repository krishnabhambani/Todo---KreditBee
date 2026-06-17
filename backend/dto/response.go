package dto

import (
	"time"

	"github.com/todo-app/backend/models"
)

type UserResponse struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type LoginResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

type TodoResponse struct {
	ID           uint       `json:"id"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	Completed    bool       `json:"completed"`
	DueDate      *time.Time `json:"due_date"`
	UserID       uint       `json:"user_id"`
	ParentTodoID *uint      `json:"parent_todo_id"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`

	Subtasks []TodoResponse `json:"subtasks,omitempty"`
	Owner    *UserResponse  `json:"owner,omitempty"`

	TotalSubtasks     int     `json:"total_subtasks"`
	CompletedSubtasks int     `json:"completed_subtasks"`
	Progress          float64 `json:"progress"`
	MemberCount       int     `json:"member_count"`
	UserPermission    string  `json:"user_permission"`
	HealthStatus      string  `json:"health_status"`
	DaysRemaining     int     `json:"days_remaining"`
}

type GroupShareResponse struct {
	ID               uint      `json:"id"`
	GroupID          uint      `json:"group_id"`
	OwnerID          uint      `json:"owner_id"`
	SharedWithUserID uint      `json:"shared_with_user_id"`
	Permission       string    `json:"permission"`
	CreatedAt        time.Time `json:"created_at"`

	Group      *TodoResponse `json:"group,omitempty"`
	Owner      *UserResponse `json:"owner,omitempty"`
	SharedWith *UserResponse `json:"shared_with,omitempty"`
}

// -----------------------------------------------------------------------------
// Mapping Helpers (Domain Model -> DTO)
// -----------------------------------------------------------------------------

func MapUser(user *models.User) *UserResponse {
	if user == nil {
		return nil
	}
	return &UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}
}

func MapTodo(todo *models.Todo) *TodoResponse {
	if todo == nil {
		return nil
	}

	var subtasks []TodoResponse
	if todo.Subtasks != nil {
		subtasks = make([]TodoResponse, len(todo.Subtasks))
		for i, subtask := range todo.Subtasks {
			subtasks[i] = *MapTodo(&subtask)
		}
	}

	return &TodoResponse{
		ID:                todo.ID,
		Title:             todo.Title,
		Description:       todo.Description,
		Completed:         todo.Completed,
		DueDate:           todo.DueDate,
		UserID:            todo.UserID,
		ParentTodoID:      todo.ParentTodoID,
		CreatedAt:         todo.CreatedAt,
		UpdatedAt:         todo.UpdatedAt,
		Subtasks:          subtasks,
		Owner:             MapUser(todo.Owner),
		TotalSubtasks:     todo.TotalSubtasks,
		CompletedSubtasks: todo.CompletedSubtasks,
		Progress:          todo.Progress,
		MemberCount:       todo.MemberCount,
		UserPermission:    todo.UserPermission,
		HealthStatus:      todo.HealthStatus,
		DaysRemaining:     todo.DaysRemaining,
	}
}

func MapGroupShare(share *models.GroupShare) *GroupShareResponse {
	if share == nil {
		return nil
	}
	return &GroupShareResponse{
		ID:               share.ID,
		GroupID:          share.GroupID,
		OwnerID:          share.OwnerID,
		SharedWithUserID: share.SharedWithUserID,
		Permission:       share.Permission,
		CreatedAt:        share.CreatedAt,
		Group:            MapTodo(share.Group),
		Owner:            MapUser(share.Owner),
		SharedWith:       MapUser(share.SharedWith),
	}
}

func MapTodos(todos []models.Todo) []TodoResponse {
	if todos == nil {
		return []TodoResponse{}
	}
	res := make([]TodoResponse, len(todos))
	for i, t := range todos {
		res[i] = *MapTodo(&t)
	}
	return res
}

func MapGroupShareList(shares []models.GroupShare) []GroupShareResponse {
	if shares == nil {
		return []GroupShareResponse{}
	}
	res := make([]GroupShareResponse, len(shares))
	for i, s := range shares {
		res[i] = *MapGroupShare(&s)
	}
	return res
}
