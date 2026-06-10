package services

import (
	"testing"
	"time"

	"github.com/todo-app/backend/models"
)

func TestCalculateGroupHealth(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name           string
		todo           models.Todo
		expectedHealth string
		expectedDays   int
	}{
		{
			name: "Completed Group",
			todo: models.Todo{
				Completed: true,
				Subtasks: []models.Todo{
					{Completed: true},
				},
			},
			expectedHealth: "COMPLETED",
			expectedDays:   9999,
		},
		{
			name: "Overdue Group",
			todo: models.Todo{
				Completed: false,
				DueDate:   func() *time.Time { t := now.Add(-24 * time.Hour); return &t }(),
				Subtasks: []models.Todo{
					{Completed: false},
				},
			},
			expectedHealth: "OVERDUE",
			expectedDays:   -1,
		},
		{
			name: "At Risk Group - Deadline Close (2 days remaining, low progress)",
			todo: models.Todo{
				Completed: false,
				DueDate:   func() *time.Time { t := now.Add(48 * time.Hour); return &t }(),
				Subtasks: []models.Todo{
					{Completed: false},
					{Completed: false},
					{Completed: false},
					{Completed: true}, // 25% progress
				},
			},
			expectedHealth: "AT_RISK",
			expectedDays:   2,
		},
		{
			name: "On Track Group - Deadline Far (10 days remaining)",
			todo: models.Todo{
				Completed: false,
				DueDate:   func() *time.Time { t := now.Add(240 * time.Hour); return &t }(),
				Subtasks: []models.Todo{
					{Completed: false},
				},
			},
			expectedHealth: "ON_TRACK",
			expectedDays:   10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			CalculateGroupHealth(&tt.todo)
			if tt.todo.HealthStatus != tt.expectedHealth {
				t.Errorf("Expected health status %s, got %s", tt.expectedHealth, tt.todo.HealthStatus)
			}
			if tt.todo.DaysRemaining != tt.expectedDays {
				t.Errorf("Expected days remaining %d, got %d", tt.expectedDays, tt.todo.DaysRemaining)
			}
		})
	}
}
