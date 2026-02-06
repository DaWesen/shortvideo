package dao

import (
	"context"
	"errors"
	"shortvideo/internal/danmu/model"

	"gorm.io/gorm"
)

type DanmuRepository interface {
	Create(ctx context.Context, danmu *model.Danmu) error
	FindByID(ctx context.Context, id int64) (*model.Danmu, error)
	Delete(ctx context.Context, id int64) error
	ListByLiveID(ctx context.Context, liveID int64, page, pageSize int) ([]*model.Danmu, int64, error)
	ListByLiveIDAndTime(ctx context.Context, liveID int64, startTime, endTime string, limit int) ([]*model.Danmu, error)
	CountByLiveID(ctx context.Context, liveID int64) (int64, error)
	GetDanmuStats(ctx context.Context, liveID int64) (*model.DanmuStats, error)
	GetActiveUsers(ctx context.Context, liveID int64, limit int) ([]int64, error)
	WithTransaction(ctx context.Context, fn func(txRepo DanmuRepository) error) error
}

type DanmuFilterRepository interface {
	CreateOrUpdate(ctx context.Context, filter *model.DanmuFilter) error
	FindByUserAndLive(ctx context.Context, userID, liveID int64) (*model.DanmuFilter, error)
	FindByLiveID(ctx context.Context, liveID int64) ([]*model.DanmuFilter, error)
	Delete(ctx context.Context, userID, liveID int64) error
	WithTransaction(ctx context.Context, fn func(txRepo DanmuFilterRepository) error) error
}

type danmuRepositoryImpl struct {
	db *gorm.DB
}

func NewDanmuRepository(db *gorm.DB) DanmuRepository {
	return &danmuRepositoryImpl{db: db}
}

func (r *danmuRepositoryImpl) Create(ctx context.Context, danmu *model.Danmu) error {
	return r.db.WithContext(ctx).Create(danmu).Error
}

func (r *danmuRepositoryImpl) FindByID(ctx context.Context, id int64) (*model.Danmu, error) {
	var danmu model.Danmu
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&danmu).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &danmu, err
}

func (r *danmuRepositoryImpl) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.Danmu{}, id).Error
}

func (r *danmuRepositoryImpl) ListByLiveID(ctx context.Context, liveID int64, page, pageSize int) ([]*model.Danmu, int64, error) {
	var danmus []*model.Danmu
	var total int64
	offset := (page - 1) * pageSize

	if err := r.db.WithContext(ctx).Model(&model.Danmu{}).
		Where("live_id = ?", liveID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.WithContext(ctx).Where("live_id = ?", liveID).
		Offset(offset).Limit(pageSize).
		Order("create_time DESC").
		Find(&danmus).Error

	return danmus, total, err
}

func (r *danmuRepositoryImpl) ListByLiveIDAndTime(ctx context.Context, liveID int64, startTime, endTime string, limit int) ([]*model.Danmu, error) {
	var danmus []*model.Danmu

	query := r.db.WithContext(ctx).Where("live_id = ?", liveID)

	if startTime != "" {
		query = query.Where("create_time >= ?", startTime)
	}
	if endTime != "" {
		query = query.Where("create_time <= ?", endTime)
	}

	err := query.Order("create_time ASC").
		Limit(limit).
		Find(&danmus).Error

	return danmus, err
}

func (r *danmuRepositoryImpl) CountByLiveID(ctx context.Context, liveID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Danmu{}).
		Where("live_id = ?", liveID).
		Count(&count).Error
	return count, err
}

func (r *danmuRepositoryImpl) GetDanmuStats(ctx context.Context, liveID int64) (*model.DanmuStats, error) {
	var stats model.DanmuStats
	stats.LiveID = liveID

	if totalCount, err := r.CountByLiveID(ctx, liveID); err == nil {
		stats.TotalDanmuCount = totalCount
	}

	var activeUserCount int64
	r.db.WithContext(ctx).Model(&model.Danmu{}).
		Select("COUNT(DISTINCT user_id)").
		Where("live_id = ?", liveID).
		Scan(&activeUserCount)
	stats.ActiveUserCount = activeUserCount

	return &stats, nil
}

func (r *danmuRepositoryImpl) GetActiveUsers(ctx context.Context, liveID int64, limit int) ([]int64, error) {
	var userIDs []int64

	err := r.db.WithContext(ctx).Model(&model.Danmu{}).
		Select("user_id").
		Where("live_id = ?", liveID).
		Group("user_id").
		Order("COUNT(*) DESC").
		Limit(limit).
		Pluck("user_id", &userIDs).Error

	return userIDs, err
}

func (r *danmuRepositoryImpl) WithTransaction(ctx context.Context, fn func(txRepo DanmuRepository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := &danmuRepositoryImpl{db: tx}
		return fn(txRepo)
	})
}

type danmuFilterRepositoryImpl struct {
	db *gorm.DB
}

func NewDanmuFilterRepository(db *gorm.DB) DanmuFilterRepository {
	return &danmuFilterRepositoryImpl{db: db}
}

func (r *danmuFilterRepositoryImpl) CreateOrUpdate(ctx context.Context, filter *model.DanmuFilter) error {
	return r.db.WithContext(ctx).Save(filter).Error
}

func (r *danmuFilterRepositoryImpl) FindByUserAndLive(ctx context.Context, userID, liveID int64) (*model.DanmuFilter, error) {
	var filter model.DanmuFilter
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND live_id = ?", userID, liveID).
		First(&filter).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &filter, err
}

func (r *danmuFilterRepositoryImpl) FindByLiveID(ctx context.Context, liveID int64) ([]*model.DanmuFilter, error) {
	var filters []*model.DanmuFilter
	err := r.db.WithContext(ctx).Where("live_id = ?", liveID).Find(&filters).Error
	return filters, err
}

func (r *danmuFilterRepositoryImpl) Delete(ctx context.Context, userID, liveID int64) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND live_id = ?", userID, liveID).
		Delete(&model.DanmuFilter{}).Error
}

func (r *danmuFilterRepositoryImpl) WithTransaction(ctx context.Context, fn func(txRepo DanmuFilterRepository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := &danmuFilterRepositoryImpl{db: tx}
		return fn(txRepo)
	})
}
