package models

import "time"

type User struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // never serialised in responses
	CreatedAt time.Time `json:"created_at"`
}
