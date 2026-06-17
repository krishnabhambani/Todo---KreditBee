package models

import "time"

type Todo struct {
	ID          uint       `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Completed   bool       `json:"completed"`
	DueDate     *time.Time `json:"due_date"`
	UserID      uint       `json:"user_id"`
	ParentTodoID *uint     `json:"parent_todo_id"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	// Associations — populated by repositories, omitted when nil/empty.
	Subtasks []Todo `json:"subtasks,omitempty"`
	Owner    *User  `json:"owner,omitempty"`

	// Virtual fields — computed by the service layer, never stored in DB.
	TotalSubtasks     int     `json:"total_subtasks"`
	CompletedSubtasks int     `json:"completed_subtasks"`
	Progress          float64 `json:"progress"`
	MemberCount       int     `json:"member_count"`
	UserPermission    string  `json:"user_permission"` // "OWNER", "EDIT", "VIEW"
	HealthStatus      string  `json:"health_status"`   // "COMPLETED", "ON_TRACK", "AT_RISK", "OVERDUE"
	DaysRemaining     int     `json:"days_remaining"`
}
