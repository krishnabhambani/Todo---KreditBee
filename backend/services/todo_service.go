package services

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/todo-app/backend/models"
	"github.com/todo-app/backend/repositories"
)

type TodoService interface {
	// Group Services
	CreateGroup(ctx context.Context, title, description string, dueDate *time.Time, userID uint) (*models.Todo, error)
	GetGroups(ctx context.Context, userID uint, search string, status string, sortParam string) ([]models.Todo, error)
	GetGroupByID(ctx context.Context, id uint, userID uint) (*models.Todo, error)
	UpdateGroup(ctx context.Context, id uint, title, description string, dueDate *time.Time, userID uint) (*models.Todo, error)
	DeleteGroup(ctx context.Context, id uint, userID uint) error

	// Subtask Services
	GetSubtasks(ctx context.Context, groupID uint, userID uint) ([]models.Todo, error)
	CreateSubtask(ctx context.Context, title, description string, dueDate *time.Time, groupID uint, userID uint) (*models.Todo, error)
	UpdateSubtask(ctx context.Context, id uint, title, description string, dueDate *time.Time, userID uint) (*models.Todo, error)
	DeleteSubtask(ctx context.Context, id uint, userID uint) error
	ToggleCompleteSubtask(ctx context.Context, id uint, userID uint) (*models.Todo, error)

	// Sharing Services
	ShareGroup(ctx context.Context, groupID, ownerID uint, email, permission string) (*models.GroupShare, error)
	RemoveShare(ctx context.Context, groupID, ownerID, targetUserID uint) error
	GetGroupMembers(ctx context.Context, groupID, userID uint) ([]models.GroupShare, error)
	GetSharedGroups(ctx context.Context, userID uint, status string, sortParam string) ([]models.Todo, error)
	SearchUsers(ctx context.Context, query string, excludeUserID uint) ([]models.User, error)
	GetPermission(ctx context.Context, entityID, userID uint) (string, error)
}

type todoService struct {
	todoRepo       repositories.TodoRepository
	groupShareRepo repositories.GroupShareRepository
	userRepo       repositories.UserRepository
}

func NewTodoService(
	todoRepo repositories.TodoRepository,
	groupShareRepo repositories.GroupShareRepository,
	userRepo repositories.UserRepository,
) TodoService {
	return &todoService{
		todoRepo:       todoRepo,
		groupShareRepo: groupShareRepo,
		userRepo:       userRepo,
	}
}

// Helper: Calculate group progress parameters dynamically
func CalculateGroupProgress(todo *models.Todo) {
	total := len(todo.Subtasks)
	completed := 0
	for _, sub := range todo.Subtasks {
		if sub.Completed {
			completed++
		}
	}
	todo.TotalSubtasks = total
	todo.CompletedSubtasks = completed
	if total == 0 {
		todo.Progress = 0.0
		todo.Completed = false
	} else {
		todo.Progress = float64(completed) / float64(total) * 100.0
		todo.Completed = (completed == total)
	}
}

// Helper: Calculate group health parameters dynamically
func CalculateGroupHealth(todo *models.Todo) {
	CalculateGroupProgress(todo)

	now := time.Now()

	// Days remaining calculations
	if todo.DueDate != nil {
		diff := todo.DueDate.Sub(now)
		days := int(diff.Hours() / 24)
		if diff.Hours() > 0 && diff.Hours() < 24 {
			days = 1
		} else if diff.Hours() < 0 {
			days = int(diff.Hours() / 24)
		}
		todo.DaysRemaining = days
	} else {
		todo.DaysRemaining = 9999
	}

	// Health status calculations
	if todo.Completed {
		todo.HealthStatus = "COMPLETED"
	} else if todo.DueDate != nil && todo.DueDate.Before(now) {
		todo.HealthStatus = "OVERDUE"
	} else if todo.DueDate != nil {
		diff := todo.DueDate.Sub(now)
		days := diff.Hours() / 24

		if days <= 3 && todo.Progress < 75 {
			todo.HealthStatus = "AT_RISK"
		} else if days <= 7 && todo.Progress < 40 {
			todo.HealthStatus = "AT_RISK"
		} else {
			todo.HealthStatus = "ON_TRACK"
		}
	} else {
		todo.HealthStatus = "ON_TRACK"
	}
}

// GetPermission evaluates the permissions of a user relative to a group or subtask ID
func (s *todoService) GetPermission(ctx context.Context, entityID, userID uint) (string, error) {
	todo, err := s.todoRepo.FindByID(ctx, entityID)
	if err != nil {
		return "", errors.New("item not found")
	}

	// Cascade up to parent group ID if this is a subtask
	if todo.ParentTodoID != nil {
		return s.GetPermission(ctx, *todo.ParentTodoID, userID)
	}

	// Owner check
	if todo.UserID == userID {
		return "OWNER", nil
	}

	// Collaborator share check
	share, err := s.groupShareRepo.FindShare(ctx, todo.ID, userID)
	if err == nil && share != nil {
		return share.Permission, nil
	}

	return "", errors.New("unauthorized to access this group")
}

// CreateGroup creates a new parent group task with deadline validations
func (s *todoService) CreateGroup(ctx context.Context, title, description string, dueDate *time.Time, userID uint) (*models.Todo, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, errors.New("group title is required")
	}

	if dueDate != nil {
		now := time.Now()
		// Tiny buffer (1 minute)
		if dueDate.Before(now.Add(-1 * time.Minute)) {
			return nil, errors.New("group deadline cannot be in the past")
		}
	}

	group := &models.Todo{
		Title:       title,
		Description: description,
		DueDate:     dueDate,
		UserID:      userID,
		Completed:   false,
	}

	err := s.todoRepo.Create(ctx, group)
	if err != nil {
		return nil, err
	}

	CalculateGroupHealth(group)
	group.UserPermission = "OWNER"
	group.MemberCount = 0
	return group, nil
}

// GetGroups fetches all groups owned by a user with status/sort filters
func (s *todoService) GetGroups(ctx context.Context, userID uint, search string, status string, sortParam string) ([]models.Todo, error) {
	groups, err := s.todoRepo.FindAllGroupsByUserID(ctx, userID, search, status, sortParam)
	if err != nil {
		return nil, err
	}

	for i := range groups {
		CalculateGroupHealth(&groups[i])
		groups[i].UserPermission = "OWNER"
		count, _ := s.groupShareRepo.CountMembersByGroupID(ctx, groups[i].ID)
		groups[i].MemberCount = count
	}
	return groups, nil
}

// GetGroupByID fetches details for a specific group (if authorized)
func (s *todoService) GetGroupByID(ctx context.Context, id uint, userID uint) (*models.Todo, error) {
	// Verify permission first
	permission, err := s.GetPermission(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	group, err := s.todoRepo.FindGroupByID(ctx, id, userID)
	if err != nil {
		return nil, errors.New("group not found")
	}

	CalculateGroupHealth(group)
	group.UserPermission = permission
	count, _ := s.groupShareRepo.CountMembersByGroupID(ctx, group.ID)
	group.MemberCount = count

	return group, nil
}

// UpdateGroup updates parent task group metadata (Owner only)
func (s *todoService) UpdateGroup(ctx context.Context, id uint, title, description string, dueDate *time.Time, userID uint) (*models.Todo, error) {
	permission, err := s.GetPermission(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	if permission != "OWNER" {
		return nil, errors.New("only group owners can edit group details")
	}

	title = strings.TrimSpace(title)
	if title == "" {
		return nil, errors.New("title is required")
	}

	group, err := s.todoRepo.FindGroupByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	group.Title = title
	group.Description = description
	group.DueDate = dueDate

	err = s.todoRepo.Update(ctx, group)
	if err != nil {
		return nil, err
	}

	CalculateGroupHealth(group)
	group.UserPermission = permission
	count, _ := s.groupShareRepo.CountMembersByGroupID(ctx, group.ID)
	group.MemberCount = count

	return group, nil
}

// DeleteGroup deletes group task (Owner only)
func (s *todoService) DeleteGroup(ctx context.Context, id uint, userID uint) error {
	permission, err := s.GetPermission(ctx, id, userID)
	if err != nil {
		return err
	}

	if permission != "OWNER" {
		return errors.New("only group owners can delete groups")
	}

	return s.todoRepo.Delete(ctx, id)
}

// GetSubtasks fetches subtasks
func (s *todoService) GetSubtasks(ctx context.Context, groupID uint, userID uint) ([]models.Todo, error) {
	_, err := s.GetPermission(ctx, groupID, userID)
	if err != nil {
		return nil, err
	}
	return s.todoRepo.FindSubtasksByGroupID(ctx, groupID, userID)
}

// CreateSubtask inserts a child task under a parent group (Owner or Editor) with validations
func (s *todoService) CreateSubtask(ctx context.Context, title, description string, dueDate *time.Time, groupID uint, userID uint) (*models.Todo, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, errors.New("subtask title is required")
	}

	permission, err := s.GetPermission(ctx, groupID, userID)
	if err != nil {
		return nil, err
	}

	if permission != "OWNER" && permission != "EDIT" {
		return nil, errors.New("unauthorized to create subtasks in this group")
	}

	if dueDate != nil {
		now := time.Now()
		if dueDate.Before(now.Add(-1 * time.Minute)) {
			return nil, errors.New("subtask deadline cannot be in the past")
		}

		// Check parent group due date
		parent, err := s.todoRepo.FindByID(ctx, groupID)
		if err == nil && parent != nil && parent.DueDate != nil {
			if dueDate.After(*parent.DueDate) {
				return nil, errors.New("subtask deadline cannot exceed group deadline")
			}
		}
	}

	subtask := &models.Todo{
		Title:        title,
		Description:  description,
		DueDate:      dueDate,
		UserID:       userID,
		ParentTodoID: &groupID,
		Completed:    false,
	}

	err = s.todoRepo.Create(ctx, subtask)
	if err != nil {
		return nil, err
	}

	// Sync parent group state in GORM
	parent, err := s.todoRepo.FindGroupByID(ctx, groupID, userID)
	if err == nil {
		CalculateGroupHealth(parent)
		_ = s.todoRepo.Update(ctx, parent)
	}

	return subtask, nil
}

// UpdateSubtask edits child task parameters (Owner or Editor) with validations
func (s *todoService) UpdateSubtask(ctx context.Context, id uint, title, description string, dueDate *time.Time, userID uint) (*models.Todo, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, errors.New("title is required")
	}

	permission, err := s.GetPermission(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	if permission != "OWNER" && permission != "EDIT" {
		return nil, errors.New("unauthorized to edit subtasks in this group")
	}

	subtask, err := s.todoRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("subtask not found")
	}

	if dueDate != nil {
		now := time.Now()
		if dueDate.Before(now.Add(-1 * time.Minute)) {
			return nil, errors.New("subtask deadline cannot be in the past")
		}

		if subtask.ParentTodoID != nil {
			parent, err := s.todoRepo.FindByID(ctx, *subtask.ParentTodoID)
			if err == nil && parent != nil && parent.DueDate != nil {
				if dueDate.After(*parent.DueDate) {
					return nil, errors.New("subtask deadline cannot exceed group deadline")
				}
			}
		}
	}

	subtask.Title = title
	subtask.Description = description
	subtask.DueDate = dueDate

	err = s.todoRepo.Update(ctx, subtask)
	if err != nil {
		return nil, err
	}

	// Sync parent group state
	if subtask.ParentTodoID != nil {
		parent, err := s.todoRepo.FindGroupByID(ctx, *subtask.ParentTodoID, userID)
		if err == nil {
			CalculateGroupHealth(parent)
			_ = s.todoRepo.Update(ctx, parent)
		}
	}

	return subtask, nil
}

// DeleteSubtask deletes subtask (Owner only)
func (s *todoService) DeleteSubtask(ctx context.Context, id uint, userID uint) error {
	permission, err := s.GetPermission(ctx, id, userID)
	if err != nil {
		return err
	}

	if permission != "OWNER" {
		return errors.New("only group owners can delete subtasks")
	}

	subtask, err := s.todoRepo.FindByID(ctx, id)
	if err != nil {
		return errors.New("subtask not found")
	}

	err = s.todoRepo.Delete(ctx, id)
	if err != nil {
		return err
	}

	// Sync parent group state
	if subtask.ParentTodoID != nil {
		parent, err := s.todoRepo.FindGroupByID(ctx, *subtask.ParentTodoID, userID)
		if err == nil {
			CalculateGroupHealth(parent)
			_ = s.todoRepo.Update(ctx, parent)
		}
	}

	return nil
}

// ToggleCompleteSubtask completes subtask and updates parent progress (Owner or Editor)
func (s *todoService) ToggleCompleteSubtask(ctx context.Context, id uint, userID uint) (*models.Todo, error) {
	permission, err := s.GetPermission(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	if permission != "OWNER" && permission != "EDIT" {
		return nil, errors.New("unauthorized to toggle subtask status")
	}

	subtask, err := s.todoRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("subtask not found")
	}

	subtask.Completed = !subtask.Completed
	err = s.todoRepo.Update(ctx, subtask)
	if err != nil {
		return nil, err
	}

	// Sync parent group state
	if subtask.ParentTodoID != nil {
		parent, err := s.todoRepo.FindGroupByID(ctx, *subtask.ParentTodoID, userID)
		if err == nil {
			CalculateGroupHealth(parent)
			_ = s.todoRepo.Update(ctx, parent)
		}
	}

	return subtask, nil
}

// ShareGroup adds a user permission to a parent task group (Owner only)
func (s *todoService) ShareGroup(ctx context.Context, groupID, ownerID uint, email, permission string) (*models.GroupShare, error) {
	groupPermission, err := s.GetPermission(ctx, groupID, ownerID)
	if err != nil || groupPermission != "OWNER" {
		return nil, errors.New("only group owners can manage sharing options")
	}

	permission = strings.ToUpper(strings.TrimSpace(permission))
	if permission != "VIEW" && permission != "EDIT" {
		return nil, errors.New("invalid permission parameter: must be VIEW or EDIT")
	}

	targetUser, err := s.userRepo.FindByEmail(ctx, strings.TrimSpace(email))
	if err != nil {
		return nil, errors.New("sharing target user not found")
	}

	if targetUser.ID == ownerID {
		return nil, errors.New("cannot share group with yourself")
	}

	existing, err := s.groupShareRepo.FindShare(ctx, groupID, targetUser.ID)
	if err == nil && existing != nil {
		return nil, errors.New("group is already shared with this user")
	}

	share := &models.GroupShare{
		GroupID:          groupID,
		OwnerID:          ownerID,
		SharedWithUserID: targetUser.ID,
		Permission:       permission,
	}

	err = s.groupShareRepo.Create(ctx, share)
	if err != nil {
		return nil, err
	}

	share.SharedWith = targetUser

	return share, nil
}

// RemoveShare removes access to a shared task group (Owner only)
func (s *todoService) RemoveShare(ctx context.Context, groupID, ownerID, targetUserID uint) error {
	groupPermission, err := s.GetPermission(ctx, groupID, ownerID)
	if err != nil || groupPermission != "OWNER" {
		return errors.New("only group owners can manage sharing options")
	}

	return s.groupShareRepo.Delete(ctx, groupID, targetUserID)
}

// GetGroupMembers retrieves current collaborators list (For group members)
func (s *todoService) GetGroupMembers(ctx context.Context, groupID, userID uint) ([]models.GroupShare, error) {
	_, err := s.GetPermission(ctx, groupID, userID)
	if err != nil {
		return nil, err
	}

	return s.groupShareRepo.FindMembersByGroupID(ctx, groupID)
}

// GetSharedGroups lists parent groups shared with a target user ID (sorted/filtered in memory)
func (s *todoService) GetSharedGroups(ctx context.Context, userID uint, status string, sortParam string) ([]models.Todo, error) {
	shares, err := s.groupShareRepo.FindSharedGroupsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var groups []models.Todo
	now := time.Now()
	for _, share := range shares {
		if share.Group != nil {
			group := *share.Group
			CalculateGroupHealth(&group)
			group.UserPermission = share.Permission
			count, _ := s.groupShareRepo.CountMembersByGroupID(ctx, group.ID)
			group.MemberCount = count

			// Apply in-memory status filter
			keep := true
			switch status {
			case "overdue":
				keep = group.DueDate != nil && group.DueDate.Before(now) && !group.Completed
			case "due-today":
				if group.DueDate != nil {
					todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
					todayEnd := todayStart.Add(24 * time.Hour)
					keep = !group.DueDate.Before(todayStart) && group.DueDate.Before(todayEnd)
				} else {
					keep = false
				}
			case "due-this-week":
				if group.DueDate != nil {
					todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
					weekEnd := todayStart.Add(7 * 24 * time.Hour)
					keep = !group.DueDate.Before(todayStart) && group.DueDate.Before(weekEnd)
				} else {
					keep = false
				}
			case "completed":
				keep = group.Completed
			case "active":
				keep = !group.Completed
			}

			if keep {
				groups = append(groups, group)
			}
		}
	}

	// Apply in-memory sorting
	if sortParam == "deadline" {
		sort.Slice(groups, func(i, j int) bool {
			if groups[i].DueDate == nil {
				return false
			}
			if groups[j].DueDate == nil {
				return true
			}
			return groups[i].DueDate.Before(*groups[j].DueDate)
		})
	} else if sortParam == "deadline-desc" {
		sort.Slice(groups, func(i, j int) bool {
			if groups[i].DueDate == nil {
				return false
			}
			if groups[j].DueDate == nil {
				return true
			}
			return groups[i].DueDate.After(*groups[j].DueDate)
		})
	} else if sortParam == "updated" {
		sort.Slice(groups, func(i, j int) bool {
			return groups[i].UpdatedAt.After(groups[j].UpdatedAt)
		})
	}

	return groups, nil
}

// SearchUsers allows group owners to search for other members on email or name
func (s *todoService) SearchUsers(ctx context.Context, query string, excludeUserID uint) ([]models.User, error) {
	return s.userRepo.SearchUsers(ctx, query, excludeUserID)
}
