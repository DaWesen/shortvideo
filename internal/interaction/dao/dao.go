package dao

import (
	"context"
	"errors"
	"shortvideo/internal/interaction/model"

	"gorm.io/gorm"
)

type CommentRepository interface {
	Create(ctx context.Context, comment *model.Comment) error
	Delete(ctx context.Context, commentID, userID, videoID int64) error
	FindByID(ctx context.Context, id int64) (*model.Comment, error)
	ListByVideoID(ctx context.Context, videoID int64, page, pageSize int) ([]*model.Comment, int64, error)
	CountByVideoID(ctx context.Context, videoID int64) (int64, error)
	ListReplies(ctx context.Context, commentID int64, page, pageSize int) ([]*model.Comment, int64, error)
	WithTransaction(ctx context.Context, fn func(txRepo CommentRepository) error) error
}

type LikeRepository interface {
	Create(ctx context.Context, like *model.Like) error
	Delete(ctx context.Context, userID, videoID int64) error
	Find(ctx context.Context, userID, videoID int64) (*model.Like, error)
	Exists(ctx context.Context, userID, videoID int64) (bool, error)
	CountByVideoID(ctx context.Context, videoID int64) (int64, error)
	ListByUserID(ctx context.Context, userID int64, page, pageSize int) ([]*model.Like, int64, error)
	ListByVideoIDs(ctx context.Context, videoIDs []int64) ([]*model.Like, error)
	WithTransaction(ctx context.Context, fn func(txRepo LikeRepository) error) error
}

type StarRepository interface {
	Create(ctx context.Context, star *model.Star) error
	Delete(ctx context.Context, userID, videoID int64) error
	Find(ctx context.Context, userID, videoID int64) (*model.Star, error)
	Exists(ctx context.Context, userID, videoID int64) (bool, error)
	CountByVideoID(ctx context.Context, videoID int64) (int64, error)
	ListByUserID(ctx context.Context, userID int64, page, pageSize int) ([]*model.Star, int64, error)
	WithTransaction(ctx context.Context, fn func(txRepo StarRepository) error) error
}

type ShareRepository interface {
	Create(ctx context.Context, share *model.Share) error
	CountByVideoID(ctx context.Context, videoID int64) (int64, error)
	ListByUserID(ctx context.Context, userID int64, page, pageSize int) ([]*model.Share, int64, error)
	WithTransaction(ctx context.Context, fn func(txRepo ShareRepository) error) error
}

type VideoInteractionStatsRepository interface {
	CreateOrUpdate(ctx context.Context, stats *model.VideoInteractionStats) error
	FindByVideoID(ctx context.Context, videoID int64) (*model.VideoInteractionStats, error)
	IncrementViewCount(ctx context.Context, videoID int64) error
	IncrementLikeCount(ctx context.Context, videoID int64) error
	DecrementLikeCount(ctx context.Context, videoID int64) error
	IncrementCommentCount(ctx context.Context, videoID int64) error
	IncrementStarCount(ctx context.Context, videoID int64) error
	DecrementStarCount(ctx context.Context, videoID int64) error
	IncrementShareCount(ctx context.Context, videoID int64) error
	WithTransaction(ctx context.Context, fn func(txRepo VideoInteractionStatsRepository) error) error
}

type commentRepositoryImpl struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) CommentRepository {
	return &commentRepositoryImpl{db: db}
}

func (r *commentRepositoryImpl) Create(ctx context.Context, comment *model.Comment) error {
	return r.db.WithContext(ctx).Create(comment).Error
}

func (r *commentRepositoryImpl) Delete(ctx context.Context, commentID, userID, videoID int64) error {
	query := r.db.WithContext(ctx).Where("id = ?", commentID)

	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if videoID > 0 {
		query = query.Where("video_id = ?", videoID)
	}

	return query.Delete(&model.Comment{}).Error
}

func (r *commentRepositoryImpl) FindByID(ctx context.Context, id int64) (*model.Comment, error) {
	var comment model.Comment
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&comment).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &comment, err
}

func (r *commentRepositoryImpl) ListByVideoID(ctx context.Context, videoID int64, page, pageSize int) ([]*model.Comment, int64, error) {
	var comments []*model.Comment
	var total int64
	offset := (page - 1) * pageSize

	if err := r.db.WithContext(ctx).Model(&model.Comment{}).
		Where("video_id = ? AND reply_to_id = 0", videoID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.WithContext(ctx).Where("video_id = ? AND reply_to_id = 0", videoID).
		Offset(offset).Limit(pageSize).
		Order("create_time DESC").
		Find(&comments).Error

	return comments, total, err
}

func (r *commentRepositoryImpl) CountByVideoID(ctx context.Context, videoID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Comment{}).
		Where("video_id = ?", videoID).
		Count(&count).Error
	return count, err
}

func (r *commentRepositoryImpl) ListReplies(ctx context.Context, commentID int64, page, pageSize int) ([]*model.Comment, int64, error) {
	var replies []*model.Comment
	var total int64
	offset := (page - 1) * pageSize

	if err := r.db.WithContext(ctx).Model(&model.Comment{}).
		Where("reply_to_id = ?", commentID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.WithContext(ctx).Where("reply_to_id = ?", commentID).
		Offset(offset).Limit(pageSize).
		Order("create_time ASC").
		Find(&replies).Error

	return replies, total, err
}

func (r *commentRepositoryImpl) WithTransaction(ctx context.Context, fn func(txRepo CommentRepository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := &commentRepositoryImpl{db: tx}
		return fn(txRepo)
	})
}

type likeRepositoryImpl struct {
	db *gorm.DB
}

func NewLikeRepository(db *gorm.DB) LikeRepository {
	return &likeRepositoryImpl{db: db}
}

func (r *likeRepositoryImpl) Create(ctx context.Context, like *model.Like) error {
	return r.db.WithContext(ctx).Create(like).Error
}

func (r *likeRepositoryImpl) Delete(ctx context.Context, userID, videoID int64) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND video_id = ?", userID, videoID).
		Delete(&model.Like{}).Error
}

func (r *likeRepositoryImpl) Find(ctx context.Context, userID, videoID int64) (*model.Like, error) {
	var like model.Like
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND video_id = ?", userID, videoID).
		First(&like).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &like, err
}

func (r *likeRepositoryImpl) Exists(ctx context.Context, userID, videoID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Like{}).
		Where("user_id = ? AND video_id = ?", userID, videoID).
		Count(&count).Error
	return count > 0, err
}

func (r *likeRepositoryImpl) CountByVideoID(ctx context.Context, videoID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Like{}).
		Where("video_id = ?", videoID).
		Count(&count).Error
	return count, err
}

func (r *likeRepositoryImpl) ListByUserID(ctx context.Context, userID int64, page, pageSize int) ([]*model.Like, int64, error) {
	var likes []*model.Like
	var total int64
	offset := (page - 1) * pageSize

	if err := r.db.WithContext(ctx).Model(&model.Like{}).
		Where("user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.WithContext(ctx).Where("user_id = ?", userID).
		Offset(offset).Limit(pageSize).
		Order("created_at DESC").
		Find(&likes).Error

	return likes, total, err
}

func (r *likeRepositoryImpl) ListByVideoIDs(ctx context.Context, videoIDs []int64) ([]*model.Like, error) {
	var likes []*model.Like
	err := r.db.WithContext(ctx).Where("video_id IN ?", videoIDs).Find(&likes).Error
	return likes, err
}

func (r *likeRepositoryImpl) WithTransaction(ctx context.Context, fn func(txRepo LikeRepository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := &likeRepositoryImpl{db: tx}
		return fn(txRepo)
	})
}

type starRepositoryImpl struct {
	db *gorm.DB
}

func NewStarRepository(db *gorm.DB) StarRepository {
	return &starRepositoryImpl{db: db}
}

func (r *starRepositoryImpl) Create(ctx context.Context, star *model.Star) error {
	return r.db.WithContext(ctx).Create(star).Error
}

func (r *starRepositoryImpl) Delete(ctx context.Context, userID, videoID int64) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND video_id = ?", userID, videoID).
		Delete(&model.Star{}).Error
}

func (r *starRepositoryImpl) Find(ctx context.Context, userID, videoID int64) (*model.Star, error) {
	var star model.Star
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND video_id = ?", userID, videoID).
		First(&star).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &star, err
}

func (r *starRepositoryImpl) Exists(ctx context.Context, userID, videoID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Star{}).
		Where("user_id = ? AND video_id = ?", userID, videoID).
		Count(&count).Error
	return count > 0, err
}

func (r *starRepositoryImpl) CountByVideoID(ctx context.Context, videoID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Star{}).
		Where("video_id = ?", videoID).
		Count(&count).Error
	return count, err
}

func (r *starRepositoryImpl) ListByUserID(ctx context.Context, userID int64, page, pageSize int) ([]*model.Star, int64, error) {
	var stars []*model.Star
	var total int64
	offset := (page - 1) * pageSize

	if err := r.db.WithContext(ctx).Model(&model.Star{}).
		Where("user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.WithContext(ctx).Where("user_id = ?", userID).
		Offset(offset).Limit(pageSize).
		Order("created_at DESC").
		Find(&stars).Error

	return stars, total, err
}

func (r *starRepositoryImpl) WithTransaction(ctx context.Context, fn func(txRepo StarRepository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := &starRepositoryImpl{db: tx}
		return fn(txRepo)
	})
}

type shareRepositoryImpl struct {
	db *gorm.DB
}

func NewShareRepository(db *gorm.DB) ShareRepository {
	return &shareRepositoryImpl{db: db}
}

func (r *shareRepositoryImpl) Create(ctx context.Context, share *model.Share) error {
	return r.db.WithContext(ctx).Create(share).Error
}

func (r *shareRepositoryImpl) CountByVideoID(ctx context.Context, videoID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Share{}).
		Where("video_id = ?", videoID).
		Count(&count).Error
	return count, err
}

func (r *shareRepositoryImpl) ListByUserID(ctx context.Context, userID int64, page, pageSize int) ([]*model.Share, int64, error) {
	var shares []*model.Share
	var total int64
	offset := (page - 1) * pageSize

	if err := r.db.WithContext(ctx).Model(&model.Share{}).
		Where("user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.WithContext(ctx).Where("user_id = ?", userID).
		Offset(offset).Limit(pageSize).
		Order("created_at DESC").
		Find(&shares).Error

	return shares, total, err
}

func (r *shareRepositoryImpl) WithTransaction(ctx context.Context, fn func(txRepo ShareRepository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := &shareRepositoryImpl{db: tx}
		return fn(txRepo)
	})
}

type videoInteractionStatsRepositoryImpl struct {
	db *gorm.DB
}

func NewVideoInteractionStatsRepository(db *gorm.DB) VideoInteractionStatsRepository {
	return &videoInteractionStatsRepositoryImpl{db: db}
}

func (r *videoInteractionStatsRepositoryImpl) CreateOrUpdate(ctx context.Context, stats *model.VideoInteractionStats) error {
	return r.db.WithContext(ctx).Save(stats).Error
}

func (r *videoInteractionStatsRepositoryImpl) FindByVideoID(ctx context.Context, videoID int64) (*model.VideoInteractionStats, error) {
	var stats model.VideoInteractionStats
	err := r.db.WithContext(ctx).Where("video_id = ?", videoID).First(&stats).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &stats, err
}

func (r *videoInteractionStatsRepositoryImpl) IncrementViewCount(ctx context.Context, videoID int64) error {
	return r.db.WithContext(ctx).Model(&model.VideoInteractionStats{}).
		Where("video_id = ?", videoID).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

func (r *videoInteractionStatsRepositoryImpl) IncrementLikeCount(ctx context.Context, videoID int64) error {
	return r.db.WithContext(ctx).Model(&model.VideoInteractionStats{}).
		Where("video_id = ?", videoID).
		UpdateColumn("like_count", gorm.Expr("like_count + 1")).Error
}

func (r *videoInteractionStatsRepositoryImpl) DecrementLikeCount(ctx context.Context, videoID int64) error {
	return r.db.WithContext(ctx).Model(&model.VideoInteractionStats{}).
		Where("video_id = ?", videoID).
		UpdateColumn("like_count", gorm.Expr("GREATEST(like_count - 1, 0)")).Error
}

func (r *videoInteractionStatsRepositoryImpl) IncrementCommentCount(ctx context.Context, videoID int64) error {
	return r.db.WithContext(ctx).Model(&model.VideoInteractionStats{}).
		Where("video_id = ?", videoID).
		UpdateColumn("comment_count", gorm.Expr("comment_count + 1")).Error
}

func (r *videoInteractionStatsRepositoryImpl) IncrementStarCount(ctx context.Context, videoID int64) error {
	return r.db.WithContext(ctx).Model(&model.VideoInteractionStats{}).
		Where("video_id = ?", videoID).
		UpdateColumn("star_count", gorm.Expr("star_count + 1")).Error
}

func (r *videoInteractionStatsRepositoryImpl) DecrementStarCount(ctx context.Context, videoID int64) error {
	return r.db.WithContext(ctx).Model(&model.VideoInteractionStats{}).
		Where("video_id = ?", videoID).
		UpdateColumn("star_count", gorm.Expr("GREATEST(star_count - 1, 0)")).Error
}

func (r *videoInteractionStatsRepositoryImpl) IncrementShareCount(ctx context.Context, videoID int64) error {
	return r.db.WithContext(ctx).Model(&model.VideoInteractionStats{}).
		Where("video_id = ?", videoID).
		UpdateColumn("share_count", gorm.Expr("share_count + 1")).Error
}

func (r *videoInteractionStatsRepositoryImpl) WithTransaction(ctx context.Context, fn func(txRepo VideoInteractionStatsRepository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := &videoInteractionStatsRepositoryImpl{db: tx}
		return fn(txRepo)
	})
}
