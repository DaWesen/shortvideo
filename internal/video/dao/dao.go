package dao

import (
	"context"
	"errors"
	"shortvideo/internal/video/model"

	"gorm.io/gorm"
)

type VideoRepository interface {
	Create(ctx context.Context, video *model.Video) error
	FindByID(ctx context.Context, id int64) (*model.Video, error)
	Update(ctx context.Context, video *model.Video) error
	Delete(ctx context.Context, id int64, userID int64) error
	ListByAuthorID(ctx context.Context, authorID int64, page, pageSize int) ([]*model.Video, int64, error)
	ListByIDs(ctx context.Context, ids []int64) ([]*model.Video, error)
	BatchGetByIDs(ctx context.Context, ids []int64) (map[int64]*model.Video, error)
	ListFeedVideos(ctx context.Context, latestTime int64, pageSize int) ([]*model.Video, error)
	Search(ctx context.Context, keyword string, page, pageSize int) ([]*model.Video, int64, error)
	CountByAuthorID(ctx context.Context, authorID int64) (int64, error)
	GetStats(ctx context.Context, videoID int64) (*model.VideoStats, error)
	UpdateLikeCount(ctx context.Context, videoID int64, delta int64) error
	UpdateCommentCount(ctx context.Context, videoID int64, delta int64) error
	UpdateShareCount(ctx context.Context, videoID int64, delta int64) error
	IncrementViewCount(ctx context.Context, videoID int64) error
	WithTransaction(ctx context.Context, fn func(txRepo VideoRepository) error) error
}

type videoRepositoryImpl struct {
	db *gorm.DB
}

func NewVideoRepository(db *gorm.DB) VideoRepository {
	return &videoRepositoryImpl{db: db}
}

func (r *videoRepositoryImpl) Create(ctx context.Context, video *model.Video) error {
	return r.db.WithContext(ctx).Create(video).Error
}

func (r *videoRepositoryImpl) FindByID(ctx context.Context, id int64) (*model.Video, error) {
	var video model.Video
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&video).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &video, err
}

func (r *videoRepositoryImpl) Update(ctx context.Context, video *model.Video) error {
	return r.db.WithContext(ctx).Save(video).Error
}

func (r *videoRepositoryImpl) Delete(ctx context.Context, id int64, userID int64) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND author_id = ?", id, userID).
		Delete(&model.Video{}).Error
}

func (r *videoRepositoryImpl) ListByAuthorID(ctx context.Context, authorID int64, page, pageSize int) ([]*model.Video, int64, error) {
	var videos []*model.Video
	var total int64
	offset := (page - 1) * pageSize

	if err := r.db.WithContext(ctx).Model(&model.Video{}).
		Where("author_id = ?", authorID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.WithContext(ctx).Where("author_id = ?", authorID).
		Offset(offset).Limit(pageSize).
		Order("publish_time DESC").
		Find(&videos).Error

	return videos, total, err
}

func (r *videoRepositoryImpl) ListByIDs(ctx context.Context, ids []int64) ([]*model.Video, error) {
	var videos []*model.Video
	err := r.db.WithContext(ctx).Where("id IN ?", ids).
		Order("publish_time DESC").
		Find(&videos).Error
	return videos, err
}

func (r *videoRepositoryImpl) BatchGetByIDs(ctx context.Context, ids []int64) (map[int64]*model.Video, error) {
	var videos []*model.Video
	result := make(map[int64]*model.Video)

	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&videos).Error
	if err != nil {
		return nil, err
	}

	for _, video := range videos {
		result[video.ID] = video
	}

	return result, nil
}

func (r *videoRepositoryImpl) ListFeedVideos(ctx context.Context, latestTime int64, pageSize int) ([]*model.Video, error) {
	var videos []*model.Video

	query := r.db.WithContext(ctx)
	if latestTime > 0 {
		query = query.Where("publish_time < ?", latestTime)
	}

	err := query.Order("publish_time DESC").
		Limit(pageSize).
		Find(&videos).Error

	return videos, err
}

func (r *videoRepositoryImpl) Search(ctx context.Context, keyword string, page, pageSize int) ([]*model.Video, int64, error) {
	var videos []*model.Video
	var total int64
	offset := (page - 1) * pageSize
	db := r.db.WithContext(ctx)

	if keyword != "" {
		db = db.Where("title LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	if err := db.Model(&model.Video{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := db.Offset(offset).Limit(pageSize).
		Order("publish_time DESC").
		Find(&videos).Error

	return videos, total, err
}

func (r *videoRepositoryImpl) CountByAuthorID(ctx context.Context, authorID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Video{}).
		Where("author_id = ?", authorID).
		Count(&count).Error
	return count, err
}

func (r *videoRepositoryImpl) GetStats(ctx context.Context, videoID int64) (*model.VideoStats, error) {
	var video model.Video
	err := r.db.WithContext(ctx).Select("like_count", "comment_count").
		Where("id = ?", videoID).First(&video).Error
	if err != nil {
		return nil, err
	}

	return &model.VideoStats{
		VideoID:      videoID,
		LikeCount:    video.LikeCount,
		CommentCount: video.CommentCount,
	}, nil
}

func (r *videoRepositoryImpl) UpdateLikeCount(ctx context.Context, videoID int64, delta int64) error {
	return r.db.WithContext(ctx).Model(&model.Video{}).
		Where("id = ?", videoID).
		UpdateColumn("like_count", gorm.Expr("like_count + ?", delta)).Error
}

func (r *videoRepositoryImpl) UpdateCommentCount(ctx context.Context, videoID int64, delta int64) error {
	return r.db.WithContext(ctx).Model(&model.Video{}).
		Where("id = ?", videoID).
		UpdateColumn("comment_count", gorm.Expr("comment_count + ?", delta)).Error
}

func (r *videoRepositoryImpl) UpdateShareCount(ctx context.Context, videoID int64, delta int64) error {
	return r.db.WithContext(ctx).Model(&model.Video{}).
		Where("id = ?", videoID).
		UpdateColumn("share_count", gorm.Expr("share_count + ?", delta)).Error
}

func (r *videoRepositoryImpl) IncrementViewCount(ctx context.Context, videoID int64) error {
	return r.db.WithContext(ctx).Model(&model.Video{}).
		Where("id = ?", videoID).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

func (r *videoRepositoryImpl) WithTransaction(ctx context.Context, fn func(txRepo VideoRepository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := &videoRepositoryImpl{db: tx}
		return fn(txRepo)
	})
}
