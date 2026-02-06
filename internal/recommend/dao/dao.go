package dao

import (
	"context"
	"errors"
	"shortvideo/internal/recommend/model"
	"time"

	"gorm.io/gorm"
)

type UserActionRepository interface {
	Create(ctx context.Context, action *model.UserAction) error
	BatchCreate(ctx context.Context, actions []*model.UserAction) error
	FindByUserID(ctx context.Context, userID int64, limit int) ([]*model.UserAction, error)
	FindByItemID(ctx context.Context, itemID int64, itemType string, limit int) ([]*model.UserAction, error)
	FindRecentActions(ctx context.Context, userID int64, actionTypes []string, limit int) ([]*model.UserAction, error)
	CountByUserAndItem(ctx context.Context, userID, itemID int64, itemType, actionType string) (int64, error)
	GetActionStats(ctx context.Context, userID int64, startTime, endTime time.Time) (*UserActionStats, error)
	GetPopularItems(ctx context.Context, itemType string, days int, limit int) ([]*PopularItem, error)
	GetSimilarUsers(ctx context.Context, userID int64, limit int) ([]int64, error)
	WithTransaction(ctx context.Context, fn func(txRepo UserActionRepository) error) error
}

type VideoTagRepository interface {
	Create(ctx context.Context, tag *model.VideoTag) error
	BatchCreate(ctx context.Context, tags []*model.VideoTag) error
	DeleteByVideoID(ctx context.Context, videoID int64) error
	FindByVideoID(ctx context.Context, videoID int64) ([]*model.VideoTag, error)
	FindByTag(ctx context.Context, tag string, limit int) ([]*model.VideoTag, error)
	FindRelatedTags(ctx context.Context, tag string, limit int) ([]string, error)
	BatchGetVideoTags(ctx context.Context, videoIDs []int64) (map[int64][]string, error)
	WithTransaction(ctx context.Context, fn func(txRepo VideoTagRepository) error) error
}

type UserPreferenceRepository interface {
	CreateOrUpdate(ctx context.Context, preference *model.UserPreference) error
	FindByUserID(ctx context.Context, userID int64, limit int) ([]*model.UserPreference, error)
	FindByTag(ctx context.Context, tag string, limit int) ([]*model.UserPreference, error)
	GetUserTagWeight(ctx context.Context, userID int64, tag string) (float64, error)
	UpdateUserTagWeight(ctx context.Context, userID int64, tag string, delta float64) error
	BatchGetUserPreferences(ctx context.Context, userIDs []int64, limit int) (map[int64][]string, error)
	GetTopKPreferences(ctx context.Context, userID int64, k int) ([]*model.UserPreference, error)
	WithTransaction(ctx context.Context, fn func(txRepo UserPreferenceRepository) error) error
}

type UserActionStats struct {
	UserID       int64
	TotalActions int64
	LikeCount    int64
	ViewCount    int64
	CommentCount int64
	ShareCount   int64
	AvgDuration  float64
	StartTime    time.Time
	EndTime      time.Time
}

type PopularItem struct {
	ItemID      int64
	ItemType    string
	Score       float64
	ActionCount int64
}

type userActionRepositoryImpl struct {
	db *gorm.DB
}

func NewUserActionRepository(db *gorm.DB) UserActionRepository {
	return &userActionRepositoryImpl{db: db}
}

func (r *userActionRepositoryImpl) Create(ctx context.Context, action *model.UserAction) error {
	return r.db.WithContext(ctx).Create(action).Error
}

func (r *userActionRepositoryImpl) BatchCreate(ctx context.Context, actions []*model.UserAction) error {
	return r.db.WithContext(ctx).CreateInBatches(actions, 100).Error
}

func (r *userActionRepositoryImpl) FindByUserID(ctx context.Context, userID int64, limit int) ([]*model.UserAction, error) {
	var actions []*model.UserAction
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("timestamp DESC").
		Limit(limit).
		Find(&actions).Error
	return actions, err
}

func (r *userActionRepositoryImpl) FindByItemID(ctx context.Context, itemID int64, itemType string, limit int) ([]*model.UserAction, error) {
	var actions []*model.UserAction
	err := r.db.WithContext(ctx).
		Where("item_id = ? AND item_type = ?", itemID, itemType).
		Order("timestamp DESC").
		Limit(limit).
		Find(&actions).Error
	return actions, err
}

func (r *userActionRepositoryImpl) FindRecentActions(ctx context.Context, userID int64, actionTypes []string, limit int) ([]*model.UserAction, error) {
	var actions []*model.UserAction
	query := r.db.WithContext(ctx).Where("user_id = ?", userID)

	if len(actionTypes) > 0 {
		query = query.Where("action_type IN ?", actionTypes)
	}

	err := query.Order("timestamp DESC").
		Limit(limit).
		Find(&actions).Error
	return actions, err
}

func (r *userActionRepositoryImpl) CountByUserAndItem(ctx context.Context, userID, itemID int64, itemType, actionType string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.UserAction{}).
		Where("user_id = ? AND item_id = ? AND item_type = ? AND action_type = ?",
			userID, itemID, itemType, actionType).
		Count(&count).Error
	return count, err
}

func (r *userActionRepositoryImpl) GetActionStats(ctx context.Context, userID int64, startTime, endTime time.Time) (*UserActionStats, error) {
	var stats UserActionStats
	stats.UserID = userID
	stats.StartTime = startTime
	stats.EndTime = endTime

	query := r.db.WithContext(ctx).Model(&model.UserAction{}).
		Where("user_id = ?", userID)

	if !startTime.IsZero() {
		query = query.Where("created_at >= ?", startTime)
	}
	if !endTime.IsZero() {
		query = query.Where("created_at <= ?", endTime)
	}

	query.Count(&stats.TotalActions)

	query.Where("action_type = ?", "like").Count(&stats.LikeCount)
	query.Where("action_type = ?", "view").Count(&stats.ViewCount)
	query.Where("action_type = ?", "comment").Count(&stats.CommentCount)
	query.Where("action_type = ?", "share").Count(&stats.ShareCount)

	var avgDuration float64
	r.db.WithContext(ctx).Model(&model.UserAction{}).
		Select("COALESCE(AVG(duration), 0)").
		Where("user_id = ? AND action_type = ?", userID, "view").
		Scan(&avgDuration)
	stats.AvgDuration = avgDuration

	return &stats, nil
}

func (r *userActionRepositoryImpl) GetPopularItems(ctx context.Context, itemType string, days int, limit int) ([]*PopularItem, error) {
	var popularItems []*PopularItem

	startTime := time.Now().AddDate(0, 0, -days)

	err := r.db.WithContext(ctx).Model(&model.UserAction{}).
		Select("item_id, item_type, COUNT(*) as action_count, "+
			"SUM(CASE WHEN action_type = 'like' THEN 2 "+
			"WHEN action_type = 'comment' THEN 1.5 "+
			"WHEN action_type = 'share' THEN 3 "+
			"ELSE 1 END) as score").
		Where("item_type = ? AND created_at >= ?", itemType, startTime).
		Group("item_id, item_type").
		Order("score DESC").
		Limit(limit).
		Scan(&popularItems).Error

	return popularItems, err
}

func (r *userActionRepositoryImpl) GetSimilarUsers(ctx context.Context, userID int64, limit int) ([]int64, error) {
	var similarUsers []int64
	return similarUsers, nil
}

func (r *userActionRepositoryImpl) WithTransaction(ctx context.Context, fn func(txRepo UserActionRepository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := &userActionRepositoryImpl{db: tx}
		return fn(txRepo)
	})
}

type videoTagRepositoryImpl struct {
	db *gorm.DB
}

func NewVideoTagRepository(db *gorm.DB) VideoTagRepository {
	return &videoTagRepositoryImpl{db: db}
}

func (r *videoTagRepositoryImpl) Create(ctx context.Context, tag *model.VideoTag) error {
	return r.db.WithContext(ctx).Create(tag).Error
}

func (r *videoTagRepositoryImpl) BatchCreate(ctx context.Context, tags []*model.VideoTag) error {
	return r.db.WithContext(ctx).CreateInBatches(tags, 100).Error
}

func (r *videoTagRepositoryImpl) DeleteByVideoID(ctx context.Context, videoID int64) error {
	return r.db.WithContext(ctx).
		Where("video_id = ?", videoID).
		Delete(&model.VideoTag{}).Error
}

func (r *videoTagRepositoryImpl) FindByVideoID(ctx context.Context, videoID int64) ([]*model.VideoTag, error) {
	var tags []*model.VideoTag
	err := r.db.WithContext(ctx).Where("video_id = ?", videoID).Find(&tags).Error
	return tags, err
}

func (r *videoTagRepositoryImpl) FindByTag(ctx context.Context, tag string, limit int) ([]*model.VideoTag, error) {
	var tags []*model.VideoTag
	err := r.db.WithContext(ctx).Where("tag_name = ?", tag).
		Limit(limit).
		Find(&tags).Error
	return tags, err
}

func (r *videoTagRepositoryImpl) FindRelatedTags(ctx context.Context, tag string, limit int) ([]string, error) {
	var relatedTags []string

	subQuery := r.db.WithContext(ctx).Model(&model.VideoTag{}).
		Select("video_id").
		Where("tag_name = ?", tag)

	err := r.db.WithContext(ctx).Model(&model.VideoTag{}).
		Select("tag_name").
		Where("video_id IN (?) AND tag_name != ?", subQuery, tag).
		Group("tag_name").
		Order("COUNT(*) DESC").
		Limit(limit).
		Pluck("tag_name", &relatedTags).Error

	return relatedTags, err
}

func (r *videoTagRepositoryImpl) BatchGetVideoTags(ctx context.Context, videoIDs []int64) (map[int64][]string, error) {
	var tags []*model.VideoTag
	result := make(map[int64][]string)

	err := r.db.WithContext(ctx).Where("video_id IN ?", videoIDs).Find(&tags).Error
	if err != nil {
		return nil, err
	}

	for _, tag := range tags {
		if _, exists := result[tag.VideoID]; !exists {
			result[tag.VideoID] = []string{}
		}
		result[tag.VideoID] = append(result[tag.VideoID], tag.TagName)
	}

	return result, nil
}

func (r *videoTagRepositoryImpl) WithTransaction(ctx context.Context, fn func(txRepo VideoTagRepository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := &videoTagRepositoryImpl{db: tx}
		return fn(txRepo)
	})
}

type userPreferenceRepositoryImpl struct {
	db *gorm.DB
}

func NewUserPreferenceRepository(db *gorm.DB) UserPreferenceRepository {
	return &userPreferenceRepositoryImpl{db: db}
}

func (r *userPreferenceRepositoryImpl) CreateOrUpdate(ctx context.Context, preference *model.UserPreference) error {
	return r.db.WithContext(ctx).Save(preference).Error
}

func (r *userPreferenceRepositoryImpl) FindByUserID(ctx context.Context, userID int64, limit int) ([]*model.UserPreference, error) {
	var preferences []*model.UserPreference
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("weight DESC").
		Limit(limit).
		Find(&preferences).Error
	return preferences, err
}

func (r *userPreferenceRepositoryImpl) FindByTag(ctx context.Context, tag string, limit int) ([]*model.UserPreference, error) {
	var preferences []*model.UserPreference
	err := r.db.WithContext(ctx).
		Where("tag_name = ?", tag).
		Order("weight DESC").
		Limit(limit).
		Find(&preferences).Error
	return preferences, err
}

func (r *userPreferenceRepositoryImpl) GetUserTagWeight(ctx context.Context, userID int64, tag string) (float64, error) {
	var preference model.UserPreference
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND tag_name = ?", userID, tag).
		First(&preference).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0.0, nil
	}

	return preference.Weight, err
}

func (r *userPreferenceRepositoryImpl) UpdateUserTagWeight(ctx context.Context, userID int64, tag string, delta float64) error {
	return r.db.WithContext(ctx).Model(&model.UserPreference{}).
		Where("user_id = ? AND tag_name = ?", userID, tag).
		UpdateColumn("weight", gorm.Expr("weight + ?", delta)).Error
}

func (r *userPreferenceRepositoryImpl) BatchGetUserPreferences(ctx context.Context, userIDs []int64, limit int) (map[int64][]string, error) {
	var preferences []*model.UserPreference
	result := make(map[int64][]string)

	err := r.db.WithContext(ctx).
		Where("user_id IN ?", userIDs).
		Order("user_id, weight DESC").
		Find(&preferences).Error

	if err != nil {
		return nil, err
	}

	userPrefCount := make(map[int64]int)
	for _, pref := range preferences {
		if _, exists := result[pref.UserID]; !exists {
			result[pref.UserID] = []string{}
		}

		if userPrefCount[pref.UserID] < limit {
			result[pref.UserID] = append(result[pref.UserID], pref.TagName)
			userPrefCount[pref.UserID]++
		}
	}

	return result, nil
}

func (r *userPreferenceRepositoryImpl) GetTopKPreferences(ctx context.Context, userID int64, k int) ([]*model.UserPreference, error) {
	var preferences []*model.UserPreference
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("weight DESC").
		Limit(k).
		Find(&preferences).Error
	return preferences, err
}

func (r *userPreferenceRepositoryImpl) WithTransaction(ctx context.Context, fn func(txRepo UserPreferenceRepository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := &userPreferenceRepositoryImpl{db: tx}
		return fn(txRepo)
	})
}
