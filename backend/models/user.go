package models

import (
	"time"
)

type User struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"type:varchar(100);not null" json:"name"`
	Email     string    `gorm:"type:varchar(191);uniqueIndex;not null" json:"email"`
	Password  string    `gorm:"type:varchar(255);not null" json:"-"` // hide password from JSON responses
	CreatedAt time.Time `json:"created_at"`
	Todos     []Todo    `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"todos,omitempty"`
}
