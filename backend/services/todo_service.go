package services

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/todo-app/backend/apperrors"
	"github.com/todo-app/backend/dto"
	"github.com/todo-app/backend/models"
	"github.com/todo-app/backend/repositories"
)

type TodoService interface {
	// Group Services
	CreateGroup(ctx context.Context, req dto.CreateGroupRequest, userID uint) (*models.Todo, error)
	GetGroups(ctx context.Context, userID uint, search string, status string, sortParam string, page, limit int) ([]models.Todo, *dto.PaginationMeta, error)
	GetGroupByID(ctx context.Context, id uint, userID uint) (*models.Todo, error)
	UpdateGroup(ctx context.Context, id uint, req dto.UpdateGroupRequest, userID uint) (*models.Todo, error)
	DeleteGroup(ctx context.Context, id uint, userID uint) error

	// Subtask Services
	GetSubtasks(ctx context.Context, groupID uint, userID uint, page, limit int) ([]models.Todo, *dto.PaginationMeta, error)
	CreateSubtask(ctx context.Context, req dto.CreateSubtaskRequest, userID uint) (*models.Todo, error)
	UpdateSubtask(ctx context.Context, id uint, req dto.UpdateSubtaskRequest, userID uint) (*models.Todo, error)
	DeleteSubtask(ctx context.Context, id uint, userID uint) error
	ToggleCompleteSubtask(ctx context.Context, id uint, userID uint) (*models.Todo, error)

	// Sharing Services
	ShareGroup(ctx context.Context, groupID, ownerID uint, req dto.ShareGroupRequest) (*models.GroupShare, error)
	GetSharedGroups(ctx context.Context, userID uint, search string, status string, sortParam string, page, limit int) ([]models.GroupShare, *dto.PaginationMeta, error)
	GetGroupMembers(ctx context.Context, groupID uint, requesterID uint) ([]models.GroupShare, error)
	RemoveShare(ctx context.Context, groupID uint, ownerID uint, sharedWithUserID uint) error
	UpdateSharePermission(ctx context.Context, groupID uint, ownerID uint, sharedWithUserID uint, req dto.UpdateShareRoleRequest) error
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

// Helper: Pagination
func paginate[T any](items []T, page, limit int) ([]T, *dto.PaginationMeta) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	totalItems := len(items)
	totalPages := (totalItems + limit - 1) / limit

	start := (page - 1) * limit
	if start > totalItems {
		start = totalItems
	}
	end := start + limit
	if end > totalItems {
		end = totalItems
	}

	meta := &dto.PaginationMeta{
		TotalItems:  totalItems,
		TotalPages:  totalPages,
		CurrentPage: page,
		Limit:       limit,
	}

	if start >= totalItems {
		return []T{}, meta
	}

	return items[start:end], meta
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
	// First check if entity is a group owned by the user
	group, err := s.todoRepo.FindByID(ctx, entityID)
	if err != nil {
		return "", apperrors.NewNotFound("todo not found")
	}
	
	// If it's a subtask, look up the parent's permission
	if group.ParentTodoID != nil {
		return s.GetPermission(ctx, *group.ParentTodoID, userID)
	}

	if group.UserID == userID {
		return "OWNER", nil
	}

	// Check if group is shared with user
	share, err := s.groupShareRepo.FindShare(ctx, entityID, userID)
	if err == nil && share != nil {
		return share.Permission, nil
	}

	return "", apperrors.NewForbidden("user not authorized to access this resource")
}

// CreateGroup creates a new parent group task with deadline validations
func (s *todoService) CreateGroup(ctx context.Context, req dto.CreateGroupRequest, userID uint) (*models.Todo, error) {
	group := &models.Todo{
		Title:       strings.TrimSpace(req.Title),
		Description: strings.TrimSpace(req.Description),
		DueDate:     req.DueDate,
		UserID:      userID,
		Completed:   false,
	}

	if req.DueDate != nil {
		// Compare dates only (truncate to start of day) to allow today's date
		now := time.Now()
		todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		dueDateStart := time.Date(req.DueDate.Year(), req.DueDate.Month(), req.DueDate.Day(), 0, 0, 0, 0, now.Location())
		if dueDateStart.Before(todayStart) {
			return nil, apperrors.NewBadRequest("group deadline cannot be in the past")
		}
	}

	err := s.todoRepo.Create(ctx, group)
	if err != nil {
		return nil, apperrors.NewInternal(err, "failed to create group")
	}

	CalculateGroupHealth(group)
	group.UserPermission = "OWNER"
	group.MemberCount = 0
	return group, nil
}

// GetGroups fetches all groups owned by a user with status/sort filters and pagination
func (s *todoService) GetGroups(ctx context.Context, userID uint, search string, status string, sortParam string, page, limit int) ([]models.Todo, *dto.PaginationMeta, error) {
	groups, err := s.todoRepo.FindAllGroupsByUserID(ctx, userID, search, status, sortParam)
	if err != nil {
		return nil, nil, err
	}

	for i := range groups {
		CalculateGroupHealth(&groups[i])
		groups[i].UserPermission = "OWNER"
		count, _ := s.groupShareRepo.CountMembersByGroupID(ctx, groups[i].ID)
		groups[i].MemberCount = count
	}

	paginatedGroups, meta := paginate(groups, page, limit)
	return paginatedGroups, meta, nil
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
		return nil, apperrors.NewNotFound("group not found")
	}

	CalculateGroupHealth(group)
	group.UserPermission = permission
	count, _ := s.groupShareRepo.CountMembersByGroupID(ctx, group.ID)
	group.MemberCount = count

	return group, nil
}

// UpdateGroup updates parent task group metadata (Owner only)
func (s *todoService) UpdateGroup(ctx context.Context, id uint, req dto.UpdateGroupRequest, userID uint) (*models.Todo, error) {
	permission, err := s.GetPermission(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	if permission != "OWNER" {
		return nil, apperrors.NewForbidden("only group owners can edit group details")
	}

	group, err := s.todoRepo.FindGroupByID(ctx, id, userID)
	if err != nil {
		return nil, apperrors.NewNotFound("group not found")
	}

	group.Title = strings.TrimSpace(req.Title)
	group.Description = strings.TrimSpace(req.Description)
	group.DueDate = req.DueDate

	err = s.todoRepo.Update(ctx, group)
	if err != nil {
		return nil, apperrors.NewInternal(err, "failed to update group")
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
		return apperrors.NewForbidden("only group owners can delete a group")
	}

	return s.todoRepo.Delete(ctx, id)
}

// GetSubtasks fetches subtasks
func (s *todoService) GetSubtasks(ctx context.Context, groupID uint, userID uint, page, limit int) ([]models.Todo, *dto.PaginationMeta, error) {
	_, err := s.GetPermission(ctx, groupID, userID)
	if err != nil {
		return nil, nil, err
	}
	subtasks, err := s.todoRepo.FindSubtasksByGroupID(ctx, groupID, userID)
	if err != nil {
		return nil, nil, err
	}
	
	paginatedSubtasks, meta := paginate(subtasks, page, limit)
	return paginatedSubtasks, meta, nil
}

// CreateSubtask inserts a child task under a parent group (Owner or Editor) with validations
func (s *todoService) CreateSubtask(ctx context.Context, req dto.CreateSubtaskRequest, userID uint) (*models.Todo, error) {
	permission, err := s.GetPermission(ctx, req.GroupID, userID)
	if err != nil {
		return nil, err
	}

	if permission == "VIEW" {
		return nil, apperrors.NewForbidden("unauthorized to create subtasks in this group")
	}

	if req.DueDate != nil {
		// Compare dates only (truncate to start of day) to allow today's date
		now := time.Now()
		todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		dueDateStart := time.Date(req.DueDate.Year(), req.DueDate.Month(), req.DueDate.Day(), 0, 0, 0, 0, now.Location())
		if dueDateStart.Before(todayStart) {
			return nil, apperrors.NewBadRequest("subtask deadline cannot be in the past")
		}

		// Check parent group due date
		parent, err := s.todoRepo.FindByID(ctx, req.GroupID)
		if err == nil && parent != nil && parent.DueDate != nil {
			if req.DueDate.After(*parent.DueDate) {
				return nil, apperrors.NewBadRequest("subtask deadline cannot exceed group deadline")
			}
		}
	}

	subtask := &models.Todo{
		Title:        strings.TrimSpace(req.Title),
		Description:  strings.TrimSpace(req.Description),
		DueDate:      req.DueDate,
		UserID:       userID,
		ParentTodoID: &req.GroupID,
		Completed:    false,
	}

	err = s.todoRepo.Create(ctx, subtask)
	if err != nil {
		return nil, apperrors.NewInternal(err, "failed to create subtask")
	}

	// Sync parent group state in GORM
	parent, err := s.todoRepo.FindGroupByID(ctx, req.GroupID, userID)
	if err == nil {
		CalculateGroupHealth(parent)
		_ = s.todoRepo.Update(ctx, parent)
	}

	return subtask, nil
}

// UpdateSubtask edits child task parameters (Owner or Editor) with validations
func (s *todoService) UpdateSubtask(ctx context.Context, id uint, req dto.UpdateSubtaskRequest, userID uint) (*models.Todo, error) {
	permission, err := s.GetPermission(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	if permission == "VIEW" {
		return nil, apperrors.NewForbidden("unauthorized to update subtasks in this group")
	}

	subtask, err := s.todoRepo.FindByID(ctx, id)
	if err != nil {
		return nil, apperrors.NewNotFound("subtask not found")
	}

	if req.DueDate != nil {
		// Compare dates only (truncate to start of day) to allow today's date
		now := time.Now()
		todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		dueDateStart := time.Date(req.DueDate.Year(), req.DueDate.Month(), req.DueDate.Day(), 0, 0, 0, 0, now.Location())
		if dueDateStart.Before(todayStart) {
			return nil, apperrors.NewBadRequest("subtask deadline cannot be in the past")
		}

		if subtask.ParentTodoID != nil {
			parent, err := s.todoRepo.FindByID(ctx, *subtask.ParentTodoID)
			if err == nil && parent != nil && parent.DueDate != nil {
				if req.DueDate.After(*parent.DueDate) {
					return nil, apperrors.NewBadRequest("subtask deadline cannot exceed group deadline")
				}
			}
		}
	}

	subtask.Title = strings.TrimSpace(req.Title)
	subtask.Description = strings.TrimSpace(req.Description)
	subtask.DueDate = req.DueDate

	err = s.todoRepo.Update(ctx, subtask)
	if err != nil {
		return nil, apperrors.NewInternal(err, "failed to update subtask")
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

	if permission == "VIEW" {
		return apperrors.NewForbidden("unauthorized to delete subtasks in this group")
	}

	subtask, err := s.todoRepo.FindByID(ctx, id)
	if err != nil {
		return apperrors.NewNotFound("subtask not found")
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

	if permission == "VIEW" {
		return nil, apperrors.NewForbidden("unauthorized to complete subtasks in this group")
	}

	subtask, err := s.todoRepo.FindByID(ctx, id)
	if err != nil {
		return nil, apperrors.NewNotFound("subtask not found")
	}

	subtask.Completed = !subtask.Completed
	err = s.todoRepo.Update(ctx, subtask)
	if err != nil {
		return nil, apperrors.NewInternal(err, "failed to update subtask status")
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
func (s *todoService) ShareGroup(ctx context.Context, groupID, ownerID uint, req dto.ShareGroupRequest) (*models.GroupShare, error) {
	groupPermission, err := s.GetPermission(ctx, groupID, ownerID)
	if err != nil || groupPermission != "OWNER" {
		return nil, apperrors.NewForbidden("only group owners can manage sharing options")
	}

	permission := strings.ToUpper(strings.TrimSpace(req.Permission))
	if permission != "VIEW" && permission != "EDIT" {
		return nil, apperrors.NewBadRequest("invalid permission parameter: must be VIEW or EDIT")
	}

	targetUser, err := s.userRepo.FindByEmail(ctx, strings.TrimSpace(req.Email))
	if err != nil {
		return nil, apperrors.NewBadRequest("sharing target user not found")
	}

	if targetUser.ID == ownerID {
		return nil, apperrors.NewBadRequest("cannot share group with yourself")
	}

	existing, err := s.groupShareRepo.FindShare(ctx, groupID, targetUser.ID)
	if err == nil && existing != nil {
		return nil, apperrors.NewBadRequest("group is already shared with this user")
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
func (s *todoService) RemoveShare(ctx context.Context, groupID uint, ownerID uint, sharedWithUserID uint) error {
	groupPermission, err := s.GetPermission(ctx, groupID, ownerID)
	if err != nil || groupPermission != "OWNER" {
		return apperrors.NewForbidden("only group owners can manage sharing options")
	}

	return s.groupShareRepo.Delete(ctx, groupID, sharedWithUserID)
}

// UpdateSharePermission updates a collaborator's role/permission (Owner only)
func (s *todoService) UpdateSharePermission(ctx context.Context, groupID uint, ownerID uint, sharedWithUserID uint, req dto.UpdateShareRoleRequest) error {
	groupPermission, err := s.GetPermission(ctx, groupID, ownerID)
	if err != nil || groupPermission != "OWNER" {
		return apperrors.NewForbidden("only group owners can manage sharing options")
	}

	permission := strings.ToUpper(strings.TrimSpace(req.Permission))
	if permission != "VIEW" && permission != "EDIT" {
		return apperrors.NewBadRequest("invalid permission: must be VIEW or EDIT")
	}

	// Verify share exists
	share, err := s.groupShareRepo.FindShare(ctx, groupID, sharedWithUserID)
	if err != nil || share == nil {
		return apperrors.NewNotFound("share not found for this user")
	}

	return s.groupShareRepo.UpdatePermission(ctx, groupID, sharedWithUserID, permission)
}

// GetGroupMembers retrieves current collaborators list (For group members)
func (s *todoService) GetGroupMembers(ctx context.Context, groupID uint, requesterID uint) ([]models.GroupShare, error) {
	_, err := s.GetPermission(ctx, groupID, requesterID)
	if err != nil {
		return nil, err
	}

	return s.groupShareRepo.FindMembersByGroupID(ctx, groupID)
}

// GetSharedGroups lists parent groups shared with a target user ID (sorted/filtered in memory)
func (s *todoService) GetSharedGroups(ctx context.Context, userID uint, search string, status string, sortParam string, page, limit int) ([]models.GroupShare, *dto.PaginationMeta, error) {
	shares, err := s.groupShareRepo.FindSharedGroupsByUserID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	var result []models.GroupShare
	now := time.Now()
	searchTerm := strings.ToLower(search)

	for _, share := range shares {
		if share.Group != nil {
			group := *share.Group
			
			if search != "" {
				if !strings.Contains(strings.ToLower(group.Title), searchTerm) && !strings.Contains(strings.ToLower(group.Description), searchTerm) {
					continue
				}
			}

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
				shareCopy := share
				shareCopy.Group = &group
				result = append(result, shareCopy)
			}
		}
	}

	// Apply in-memory sorting
	if sortParam == "deadline" {
		sort.Slice(result, func(i, j int) bool {
			if result[i].Group.DueDate == nil {
				return false
			}
			if result[j].Group.DueDate == nil {
				return true
			}
			return result[i].Group.DueDate.Before(*result[j].Group.DueDate)
		})
	} else if sortParam == "deadline-desc" {
		sort.Slice(result, func(i, j int) bool {
			if result[i].Group.DueDate == nil {
				return false
			}
			if result[j].Group.DueDate == nil {
				return true
			}
			return result[i].Group.DueDate.After(*result[j].Group.DueDate)
		})
	} else if sortParam == "updated" {
		sort.Slice(result, func(i, j int) bool {
			return result[i].Group.UpdatedAt.After(result[j].Group.UpdatedAt)
		})
	}

	paginatedResult, meta := paginate(result, page, limit)
	return paginatedResult, meta, nil
}

// SearchUsers allows group owners to search for other members on email or name
func (s *todoService) SearchUsers(ctx context.Context, query string, excludeUserID uint) ([]models.User, error) {
	return s.userRepo.SearchUsers(ctx, query, excludeUserID)
}
