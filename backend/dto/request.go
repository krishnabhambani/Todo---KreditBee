package dto

import "time"

// -----------------------------------------------------------------------------
// Auth Requests
// -----------------------------------------------------------------------------

type RegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UpdatePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

// -----------------------------------------------------------------------------
// Group Requests
// -----------------------------------------------------------------------------

type CreateGroupRequest struct {
	Title       string     `json:"title" binding:"required"`
	Description string     `json:"description"`
	DueDate     *time.Time `json:"due_date"`
}

type UpdateGroupRequest struct {
	Title       string     `json:"title" binding:"required"`
	Description string     `json:"description"`
	DueDate     *time.Time `json:"due_date"`
}

// -----------------------------------------------------------------------------
// Subtask Requests
// -----------------------------------------------------------------------------

type CreateSubtaskRequest struct {
	Title       string     `json:"title" binding:"required"`
	Description string     `json:"description"`
	DueDate     *time.Time `json:"due_date"`
	GroupID     uint       `json:"group_id" binding:"required"`
}

type UpdateSubtaskRequest struct {
	Title       string     `json:"title" binding:"required"`
	Description string     `json:"description"`
	DueDate     *time.Time `json:"due_date"`
}

// -----------------------------------------------------------------------------
// Share Requests
// -----------------------------------------------------------------------------

type ShareGroupRequest struct {
	Email      string `json:"email" binding:"required,email"`
	Permission string `json:"permission" binding:"required,oneof=VIEW EDIT"`
}

type UpdateShareRoleRequest struct {
	Permission string `json:"permission" binding:"required,oneof=VIEW EDIT"`
}
