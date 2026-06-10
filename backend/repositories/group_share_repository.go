package repositories

import (
	"context"

	"github.com/todo-app/backend/models"
	"gorm.io/gorm"
)

type GroupShareRepository interface {
	Create(ctx context.Context, share *models.GroupShare) error
	Delete(ctx context.Context, groupID uint, userID uint) error
	FindShare(ctx context.Context, groupID uint, sharedWithUserID uint) (*models.GroupShare, error)
	FindMembersByGroupID(ctx context.Context, groupID uint) ([]models.GroupShare, error)
	FindSharedGroupsByUserID(ctx context.Context, userID uint) ([]models.GroupShare, error)
	CountMembersByGroupID(ctx context.Context, groupID uint) (int, error)
}

type groupShareRepository struct {
	db *gorm.DB
}

func NewGroupShareRepository(db *gorm.DB) GroupShareRepository {
	return &groupShareRepository{db: db}
}

func (r *groupShareRepository) Create(ctx context.Context, share *models.GroupShare) error {
	return r.db.WithContext(ctx).Create(share).Error
}

func (r *groupShareRepository) Delete(ctx context.Context, groupID uint, userID uint) error {
	return r.db.WithContext(ctx).
		Where("group_id = ? AND shared_with_user_id = ?", groupID, userID).
		Delete(&models.GroupShare{}).Error
}

func (r *groupShareRepository) FindShare(ctx context.Context, groupID uint, sharedWithUserID uint) (*models.GroupShare, error) {
	var share models.GroupShare
	err := r.db.WithContext(ctx).
		Where("group_id = ? AND shared_with_user_id = ?", groupID, sharedWithUserID).
		First(&share).Error
	if err != nil {
		return nil, err
	}
	return &share, nil
}

func (r *groupShareRepository) FindMembersByGroupID(ctx context.Context, groupID uint) ([]models.GroupShare, error) {
	var shares []models.GroupShare
	err := r.db.WithContext(ctx).
		Preload("SharedWith").
		Where("group_id = ?", groupID).
		Find(&shares).Error
	return shares, err
}

func (r *groupShareRepository) FindSharedGroupsByUserID(ctx context.Context, userID uint) ([]models.GroupShare, error) {
	var shares []models.GroupShare
	err := r.db.WithContext(ctx).
		Preload("Group").
		Preload("Group.Subtasks").
		Preload("Group.Owner").
		Where("shared_with_user_id = ?", userID).
		Find(&shares).Error
	return shares, err
}

func (r *groupShareRepository) CountMembersByGroupID(ctx context.Context, groupID uint) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.GroupShare{}).
		Where("group_id = ?", groupID).
		Count(&count).Error
	return int(count), err
}
