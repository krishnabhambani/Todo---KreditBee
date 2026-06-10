package models

import "time"

type GroupShare struct {
	ID               uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	GroupID          uint      `gorm:"not null;index;uniqueIndex:idx_group_share_unique" json:"group_id"`
	OwnerID          uint      `gorm:"not null" json:"owner_id"`
	SharedWithUserID uint      `gorm:"not null;index;uniqueIndex:idx_group_share_unique" json:"shared_with_user_id"`
	Permission       string    `gorm:"type:varchar(50);not null" json:"permission"` // "VIEW" or "EDIT"
	CreatedAt        time.Time `json:"created_at"`

	// Relations
	Group      *Todo `gorm:"foreignKey:GroupID;constraint:OnDelete:CASCADE" json:"group,omitempty"`
	Owner      *User `gorm:"foreignKey:OwnerID;constraint:OnDelete:CASCADE" json:"owner,omitempty"`
	SharedWith *User `gorm:"foreignKey:SharedWithUserID;constraint:OnDelete:CASCADE" json:"shared_with,omitempty"`
}
