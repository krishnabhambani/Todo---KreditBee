package models

import "time"

type GroupShare struct {
	ID               uint      `json:"id"`
	GroupID          uint      `json:"group_id"`
	OwnerID          uint      `json:"owner_id"`
	SharedWithUserID uint      `json:"shared_with_user_id"`
	Permission       string    `json:"permission"` // "VIEW" or "EDIT"
	CreatedAt        time.Time `json:"created_at"`

	// Associations — populated by repositories, omitted when nil.
	Group      *Todo `json:"group,omitempty"`
	Owner      *User `json:"owner,omitempty"`
	SharedWith *User `json:"shared_with,omitempty"`
}
