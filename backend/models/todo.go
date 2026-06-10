package models

import "time"

type Todo struct {
	ID           uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	Title        string     `gorm:"type:varchar(255);not null" json:"title"`
	Description  string     `gorm:"type:text" json:"description"`
	Completed    bool       `gorm:"default:false" json:"completed"`
	DueDate      *time.Time `gorm:"default:null" json:"due_date"`
	UserID       uint       `gorm:"not null" json:"user_id"`
	ParentTodoID *uint      `gorm:"default:null" json:"parent_todo_id"`
	Subtasks     []Todo     `gorm:"foreignKey:ParentTodoID;constraint:OnDelete:CASCADE" json:"subtasks,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`

	// Relationships
	Owner        *User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"owner,omitempty"`

	// Virtual fields for dynamic progress calculation (ignored by GORM)
	TotalSubtasks     int     `gorm:"-" json:"total_subtasks"`
	CompletedSubtasks int     `gorm:"-" json:"completed_subtasks"`
	Progress          float64 `gorm:"-" json:"progress"`
	MemberCount       int     `gorm:"-" json:"member_count"`
	UserPermission    string  `gorm:"-" json:"user_permission"` // "OWNER", "EDIT", "VIEW"
	HealthStatus      string  `gorm:"-" json:"health_status"`   // "COMPLETED", "ON_TRACK", "AT_RISK", "OVERDUE"
	DaysRemaining     int     `gorm:"-" json:"days_remaining"`
}
