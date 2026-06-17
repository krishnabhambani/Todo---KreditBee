package repositories

import (
	"context"
	"database/sql"

	"github.com/todo-app/backend/database"
	"github.com/todo-app/backend/models"
)

type GroupShareRepository interface {
	Create(ctx context.Context, share *models.GroupShare) error
	Delete(ctx context.Context, groupID uint, userID uint) error
	FindShare(ctx context.Context, groupID uint, sharedWithUserID uint) (*models.GroupShare, error)
	FindMembersByGroupID(ctx context.Context, groupID uint) ([]models.GroupShare, error)
	FindSharedGroupsByUserID(ctx context.Context, userID uint) ([]models.GroupShare, error)
	CountMembersByGroupID(ctx context.Context, groupID uint) (int, error)
	UpdatePermission(ctx context.Context, groupID uint, sharedWithUserID uint, permission string) error
}

type groupShareRepository struct {
	q *database.Queries
}

func NewGroupShareRepository(q *database.Queries) GroupShareRepository {
	return &groupShareRepository{q: q}
}

func (r *groupShareRepository) Create(ctx context.Context, share *models.GroupShare) error {
	result, err := r.q.CreateGroupShare(ctx, database.CreateGroupShareParams{
		GroupID:          uint32(share.GroupID),
		OwnerID:          uint32(share.OwnerID),
		SharedWithUserID: uint32(share.SharedWithUserID),
		Permission:       share.Permission,
	})
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	share.ID = uint(id)
	return nil
}

func (r *groupShareRepository) Delete(ctx context.Context, groupID uint, userID uint) error {
	return r.q.DeleteGroupShare(ctx, uint32(groupID), uint32(userID))
}

func (r *groupShareRepository) FindShare(ctx context.Context, groupID uint, sharedWithUserID uint) (*models.GroupShare, error) {
	row, err := r.q.GetGroupShare(ctx, uint32(groupID), uint32(sharedWithUserID))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &models.GroupShare{
		ID:               uint(row.ID),
		GroupID:          uint(row.GroupID),
		OwnerID:          uint(row.OwnerID),
		SharedWithUserID: uint(row.SharedWithUserID),
		Permission:       row.Permission,
		CreatedAt:        row.CreatedAt,
	}, nil
}

func (r *groupShareRepository) FindMembersByGroupID(ctx context.Context, groupID uint) ([]models.GroupShare, error) {
	rows, err := r.q.GetGroupSharesByGroupID(ctx, uint32(groupID))
	if err != nil {
		return nil, err
	}
	shares := make([]models.GroupShare, 0, len(rows))
	for _, row := range rows {
		shares = append(shares, models.GroupShare{
			ID:               uint(row.ID),
			GroupID:          uint(row.GroupID),
			OwnerID:          uint(row.OwnerID),
			SharedWithUserID: uint(row.SharedWithUserID),
			Permission:       row.Permission,
			CreatedAt:        row.CreatedAt,
			SharedWith: &models.User{
				ID:    uint(row.SwID),
				Name:  row.SwName,
				Email: row.SwEmail,
			},
		})
	}
	return shares, nil
}

func (r *groupShareRepository) FindSharedGroupsByUserID(ctx context.Context, userID uint) ([]models.GroupShare, error) {
	rawShares, err := r.q.GetSharesByUserID(ctx, uint32(userID))
	if err != nil {
		return nil, err
	}
	shares := make([]models.GroupShare, 0, len(rawShares))
	for _, s := range rawShares {
		groupRow, err := r.q.GetGroupByID(ctx, s.GroupID)
		if err != nil {
			continue
		}
		group := toModelTodo(groupRow)

		subtaskRows, _ := r.q.GetSubtasksByParentID(ctx, groupRow.ID)
		group.Subtasks = toModelTodos(subtaskRows)

		if ownerRow, err := r.q.GetUserByID(ctx, groupRow.UserID); err == nil {
			group.Owner = dbUserToModel(ownerRow)
		}

		shares = append(shares, models.GroupShare{
			ID:               uint(s.ID),
			GroupID:          uint(s.GroupID),
			OwnerID:          uint(s.OwnerID),
			SharedWithUserID: uint(s.SharedWithUserID),
			Permission:       s.Permission,
			CreatedAt:        s.CreatedAt,
			Group:            &group,
		})
	}
	return shares, nil
}

func (r *groupShareRepository) CountMembersByGroupID(ctx context.Context, groupID uint) (int, error) {
	count, err := r.q.CountGroupMembers(ctx, uint32(groupID))
	return int(count), err
}

func (r *groupShareRepository) UpdatePermission(ctx context.Context, groupID uint, sharedWithUserID uint, permission string) error {
	return r.q.UpdateGroupSharePermission(ctx, database.UpdateGroupSharePermissionParams{
		Permission:       permission,
		GroupID:          uint32(groupID),
		SharedWithUserID: uint32(sharedWithUserID),
	})
}
