package dao

import (
	"context"
	"errors"
	"shortvideo/internal/social/model"

	"gorm.io/gorm"
)

type FollowRepository interface {
	Create(ctx context.Context, follow *model.Follow) error
	Delete(ctx context.Context, userID, targetUserID int64) error
	Find(ctx context.Context, userID, targetUserID int64) (*model.Follow, error)
	Exists(ctx context.Context, userID, targetUserID int64) (bool, error)
	FindFollowing(ctx context.Context, userID int64, page, pageSize int) ([]*model.Follow, int64, error)
	FindFollowers(ctx context.Context, userID int64, page, pageSize int) ([]*model.Follow, int64, error)
	FindFriends(ctx context.Context, userID int64, page, pageSize int) ([]*model.Follow, error)
	CountFollowing(ctx context.Context, userID int64) (int64, error)
	CountFollowers(ctx context.Context, userID int64) (int64, error)
	CountFriends(ctx context.Context, userID int64) (int64, error)
	BatchCheckFollow(ctx context.Context, userID int64, targetUserIDs []int64) (map[int64]bool, error)
	WithTransaction(ctx context.Context, fn func(txRepo FollowRepository) error) error
}

type followRepositoryImpl struct {
	db *gorm.DB
}

func NewFollowRepository(db *gorm.DB) FollowRepository {
	return &followRepositoryImpl{db: db}
}

func (r *followRepositoryImpl) Create(ctx context.Context, follow *model.Follow) error {
	return r.db.WithContext(ctx).Create(follow).Error
}

func (r *followRepositoryImpl) Delete(ctx context.Context, userID, targetUserID int64) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND target_user_id = ?", userID, targetUserID).
		Delete(&model.Follow{}).Error
}

func (r *followRepositoryImpl) Find(ctx context.Context, userID, targetUserID int64) (*model.Follow, error) {
	var follow model.Follow
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND target_user_id = ?", userID, targetUserID).
		First(&follow).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &follow, err
}

func (r *followRepositoryImpl) Exists(ctx context.Context, userID, targetUserID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Follow{}).
		Where("user_id = ? AND target_user_id = ?", userID, targetUserID).
		Count(&count).Error
	return count > 0, err
}

func (r *followRepositoryImpl) FindFollowing(ctx context.Context, userID int64, page, pageSize int) ([]*model.Follow, int64, error) {
	var follows []*model.Follow
	var total int64
	offset := (page - 1) * pageSize

	if err := r.db.WithContext(ctx).Model(&model.Follow{}).
		Where("user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.WithContext(ctx).Where("user_id = ?", userID).
		Offset(offset).Limit(pageSize).
		Order("created_at DESC").
		Find(&follows).Error

	return follows, total, err
}

func (r *followRepositoryImpl) FindFollowers(ctx context.Context, userID int64, page, pageSize int) ([]*model.Follow, int64, error) {
	var follows []*model.Follow
	var total int64
	offset := (page - 1) * pageSize

	if err := r.db.WithContext(ctx).Model(&model.Follow{}).
		Where("target_user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.WithContext(ctx).Where("target_user_id = ?", userID).
		Offset(offset).Limit(pageSize).
		Order("created_at DESC").
		Find(&follows).Error

	return follows, total, err
}

func (r *followRepositoryImpl) FindFriends(ctx context.Context, userID int64, page, pageSize int) ([]*model.Follow, error) {
	var follows []*model.Follow
	offset := (page - 1) * pageSize

	subQuery := r.db.WithContext(ctx).Model(&model.Follow{}).
		Select("target_user_id").
		Where("user_id = ?", userID).
		Where("target_user_id IN (?)",
			r.db.Model(&model.Follow{}).
				Select("user_id").
				Where("target_user_id = ?", userID))

	err := r.db.WithContext(ctx).
		Where("user_id = ? AND target_user_id IN (?)", userID, subQuery).
		Offset(offset).Limit(pageSize).
		Order("created_at DESC").
		Find(&follows).Error

	return follows, err
}

func (r *followRepositoryImpl) CountFollowing(ctx context.Context, userID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Follow{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return count, err
}

func (r *followRepositoryImpl) CountFollowers(ctx context.Context, userID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Follow{}).
		Where("target_user_id = ?", userID).
		Count(&count).Error
	return count, err
}

func (r *followRepositoryImpl) CountFriends(ctx context.Context, userID int64) (int64, error) {
	var count int64

	subQuery := r.db.WithContext(ctx).Model(&model.Follow{}).
		Select("target_user_id").
		Where("user_id = ?", userID).
		Where("target_user_id IN (?)",
			r.db.Model(&model.Follow{}).
				Select("user_id").
				Where("target_user_id = ?", userID))

	err := r.db.WithContext(ctx).Model(&model.Follow{}).
		Where("user_id = ? AND target_user_id IN (?)", userID, subQuery).
		Count(&count).Error

	return count, err
}

func (r *followRepositoryImpl) BatchCheckFollow(ctx context.Context, userID int64, targetUserIDs []int64) (map[int64]bool, error) {
	var follows []model.Follow
	result := make(map[int64]bool)

	for _, targetID := range targetUserIDs {
		result[targetID] = false
	}

	err := r.db.WithContext(ctx).Select("target_user_id").
		Where("user_id = ? AND target_user_id IN ?", userID, targetUserIDs).
		Find(&follows).Error
	if err != nil {
		return nil, err
	}

	for _, follow := range follows {
		result[follow.TargetUserID] = true
	}

	return result, nil
}

func (r *followRepositoryImpl) WithTransaction(ctx context.Context, fn func(txRepo FollowRepository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := &followRepositoryImpl{db: tx}
		return fn(txRepo)
	})
}
