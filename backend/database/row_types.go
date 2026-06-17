package database

import (
	"database/sql"
	"time"
)

// User is the flat DB row type for the `users` table.
type User struct {
	ID        uint32
	Name      string
	Email     string
	Password  string
	CreatedAt time.Time
}

// Todo is the flat DB row type for the `todos` table.
// Virtual/computed fields live on models.Todo and are set by the service layer.
type Todo struct {
	ID           uint32
	Title        string
	Description  string
	Completed    bool
	DueDate      sql.NullTime
	UserID       uint32
	ParentTodoID sql.NullInt32
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// GroupShare is the flat DB row type for the `group_shares` table.
type GroupShare struct {
	ID               uint32
	GroupID          uint32
	OwnerID          uint32
	SharedWithUserID uint32
	Permission       string
	CreatedAt        time.Time
}

// GetGroupSharesByGroupIDRow is the JOIN result from GetGroupSharesByGroupID —
// includes the shared_with user's id, name, and email.
type GetGroupSharesByGroupIDRow struct {
	ID               uint32
	GroupID          uint32
	OwnerID          uint32
	SharedWithUserID uint32
	Permission       string
	CreatedAt        time.Time
	SwID             uint32
	SwName           string
	SwEmail          string
}
