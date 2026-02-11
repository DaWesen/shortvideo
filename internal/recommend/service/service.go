package service

import (
	"context"
	"errors"
	"shortvideo/internal/recommend/dao"
	"shortvideo/internal/recommend/model"
	"shortvideo/kitex_gen/common"
	"shortvideo/kitex_gen/recommend"
	"shortvideo/pkg/logger"
	"time"
)

var (
	ErrInternalServer   = errors.New("服务器内部错误")
	ErrInvalidParameter = errors.New("参数错误")
	ErrItemNotFound     = errors.New("项目不存在")
	ErrActionFailed     = errors.New("操作失败")
)

type RecommendService interface {
	//推荐相关
	GetRecommendVideos(ctx context.Context, userID int64, pageSize int32, offset *int64) ([]*common.Video, int64, error)
	GetRecommendUsers(ctx context.Context, userID int64, count int32) ([]*common.User, error)
	GetPersonalizedFeed(ctx context.Context, userID int64, pageSize int32, lastVideoID *int64) ([]*common.Video, int64, error)
	//用户行为相关
	RecordUserAction(ctx context.Context, userID, itemID int64, itemType, actionType, timestamp string, duration *int32, score *float64) error
	//标签相关
	GetHotTags(ctx context.Context, count int32) ([]*recommend.TagInfo, error)
	GetTagVideos(ctx context.Context, tag string, userID int64, pageSize int32) ([]*common.Video, error)
	//事务相关
	WithTransaction(ctx context.Context, fn func(txService RecommendService) error) error
}

type recommendServiceImpl struct {
	actionRepo     dao.UserActionRepository
	videoTagRepo   dao.VideoTagRepository
	preferenceRepo dao.UserPreferenceRepository
}

func NewRecommendService(
	actionRepo dao.UserActionRepository,
	videoTagRepo dao.VideoTagRepository,
	preferenceRepo dao.UserPreferenceRepository,
) RecommendService {
	return &recommendServiceImpl{
		actionRepo:     actionRepo,
		videoTagRepo:   videoTagRepo,
		preferenceRepo: preferenceRepo,
	}
}

func NewRecommendServiceWithRepo(
	actionRepo dao.UserActionRepository,
	videoTagRepo dao.VideoTagRepository,
	preferenceRepo dao.UserPreferenceRepository,
) RecommendService {
	return &recommendServiceImpl{
		actionRepo:     actionRepo,
		videoTagRepo:   videoTagRepo,
		preferenceRepo: preferenceRepo,
	}
}

// 获取推荐视频
func (s *recommendServiceImpl) GetRecommendVideos(ctx context.Context, userID int64, pageSize int32, offset *int64) ([]*common.Video, int64, error) {
	logger.Info("GetRecommendVideos request",
		logger.Int64Field("user_id", userID),
		logger.AnyField("page_size", pageSize),
		logger.AnyField("offset", offset))

	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	popularItems, err := s.actionRepo.GetPopularItems(ctx, "video", 7, int(pageSize))
	if err != nil {
		logger.Error("GetPopularItems failed", logger.ErrorField(err))
		return nil, 0, ErrInternalServer
	}

	videos := make([]*common.Video, 0, len(popularItems))
	for _, item := range popularItems {
		video := &common.Video{
			Id:           item.ItemID,
			Title:        "推荐视频",
			Description:  "这是一个推荐视频",
			CoverUrl:     "https://example.com/cover.jpg",
			Url:          "https://example.com/video.mp4",
			AuthorId:     1,
			LikeCount:    int64(item.ActionCount),
			CommentCount: 0,
			IsLike:       false,
			PublishTime:  time.Now().Unix(),
		}
		videos = append(videos, video)
	}

	nextOffset := int64(0)
	if offset != nil {
		nextOffset = *offset + int64(len(videos))
	} else {
		nextOffset = int64(len(videos))
	}

	logger.Info("GetRecommendVideos success",
		logger.Int64Field("user_id", userID),
		logger.IntField("video_count", len(videos)),
		logger.Int64Field("next_offset", nextOffset))

	return videos, nextOffset, nil
}

// 获取推荐用户
func (s *recommendServiceImpl) GetRecommendUsers(ctx context.Context, userID int64, count int32) ([]*common.User, error) {
	logger.Info("GetRecommendUsers request",
		logger.Int64Field("user_id", userID),
		logger.AnyField("count", count))

	if count <= 0 || count > 50 {
		count = 10
	}

	similarUsers, err := s.actionRepo.GetSimilarUsers(ctx, userID, int(count))
	if err != nil {
		logger.Error("GetSimilarUsers failed", logger.ErrorField(err))
		return nil, ErrInternalServer
	}

	users := make([]*common.User, 0, len(similarUsers))
	for _, similarUserID := range similarUsers {
		avatarUrl := "https://example.com/avatar.jpg"
		about := "This is a recommended user"
		user := &common.User{
			Id:            similarUserID,
			Username:      "User" + string(rune(similarUserID)),
			Password:      "",
			Avatar:        &avatarUrl,
			About:         &about,
			FollowCount:   0,
			FollowerCount: 0,
			IsFollow:      false,
		}
		users = append(users, user)
	}

	if len(users) == 0 {
		for i := 0; i < int(count); i++ {
			avatarUrl := "https://example.com/avatar.jpg"
			about := "This is a hot user"
			user := &common.User{
				Id:            int64(i + 1),
				Username:      "HotUser" + string(rune(i+1)),
				Password:      "",
				Avatar:        &avatarUrl,
				About:         &about,
				FollowCount:   0,
				FollowerCount: 0,
				IsFollow:      false,
			}
			users = append(users, user)
		}
	}

	logger.Info("GetRecommendUsers success",
		logger.Int64Field("user_id", userID),
		logger.IntField("user_count", len(users)))

	return users, nil
}

// 记录用户行为
func (s *recommendServiceImpl) RecordUserAction(ctx context.Context, userID, itemID int64, itemType, actionType, timestamp string, duration *int32, score *float64) error {
	logger.Info("RecordUserAction request",
		logger.Int64Field("user_id", userID),
		logger.Int64Field("item_id", itemID),
		logger.StringField("item_type", itemType),
		logger.StringField("action_type", actionType),
		logger.StringField("timestamp", timestamp))

	if userID <= 0 || itemID <= 0 || itemType == "" || actionType == "" {
		return ErrInvalidParameter
	}

	action := &model.UserAction{
		UserID:     userID,
		ItemID:     itemID,
		ItemType:   itemType,
		ActionType: actionType,
		Timestamp:  timestamp,
	}

	if duration != nil {
		action.Duration = *duration
	}

	if score != nil {
		action.Score = *score
	}

	if err := s.actionRepo.Create(ctx, action); err != nil {
		logger.Error("CreateUserAction failed", logger.ErrorField(err))
		return ErrInternalServer
	}

	if itemType == "video" {
		tags, err := s.videoTagRepo.FindByVideoID(ctx, itemID)
		if err != nil {
			logger.Error("FindByVideoID failed", logger.ErrorField(err))
		} else {
			for _, tag := range tags {
				weight, err := s.preferenceRepo.GetUserTagWeight(ctx, userID, tag.TagName)
				if err != nil {
					logger.Error("GetUserTagWeight failed", logger.ErrorField(err))
					continue
				}

				delta := 0.0
				switch actionType {
				case "like":
					delta = 1.0
				case "comment":
					delta = 0.8
				case "share":
					delta = 1.2
				case "view":
					delta = 0.1
				}

				if weight == 0.0 {
					preference := &model.UserPreference{
						UserID:  userID,
						TagName: tag.TagName,
						Weight:  delta,
					}
					if err := s.preferenceRepo.CreateOrUpdate(ctx, preference); err != nil {
						logger.Error("CreateOrUpdatePreference failed", logger.ErrorField(err))
					}
				} else {
					if err := s.preferenceRepo.UpdateUserTagWeight(ctx, userID, tag.TagName, delta); err != nil {
						logger.Error("UpdateUserTagWeight failed", logger.ErrorField(err))
					}
				}
			}
		}
	}

	logger.Info("RecordUserAction success",
		logger.Int64Field("user_id", userID),
		logger.Int64Field("item_id", itemID))

	return nil
}

// 获取热门标签
func (s *recommendServiceImpl) GetHotTags(ctx context.Context, count int32) ([]*recommend.TagInfo, error) {
	logger.Info("GetHotTags request", logger.AnyField("count", count))

	if count <= 0 || count > 50 {
		count = 20
	}

	tags := []*recommend.TagInfo{
		{TagName: "热门", VideoCount: 1000, ViewCount: 100000},
		{TagName: "音乐", VideoCount: 800, ViewCount: 80000},
		{TagName: "舞蹈", VideoCount: 600, ViewCount: 60000},
		{TagName: "美食", VideoCount: 500, ViewCount: 50000},
		{TagName: "旅行", VideoCount: 400, ViewCount: 40000},
		{TagName: "健身", VideoCount: 300, ViewCount: 30000},
		{TagName: "游戏", VideoCount: 200, ViewCount: 20000},
		{TagName: "学习", VideoCount: 100, ViewCount: 10000},
	}

	if int(count) < len(tags) {
		tags = tags[:count]
	}

	logger.Info("GetHotTags success", logger.IntField("tag_count", len(tags)))

	return tags, nil
}

// 获取标签相关视频
func (s *recommendServiceImpl) GetTagVideos(ctx context.Context, tag string, userID int64, pageSize int32) ([]*common.Video, error) {
	logger.Info("GetTagVideos request",
		logger.StringField("tag", tag),
		logger.Int64Field("user_id", userID),
		logger.AnyField("page_size", pageSize))

	if tag == "" {
		return nil, ErrInvalidParameter
	}

	if pageSize <= 0 || pageSize > 50 {
		pageSize = 20
	}

	tagVideos, err := s.videoTagRepo.FindByTag(ctx, tag, int(pageSize))
	if err != nil {
		logger.Error("FindByTag failed", logger.ErrorField(err))
		return nil, ErrInternalServer
	}

	videos := make([]*common.Video, 0, len(tagVideos))
	for _, tagVideo := range tagVideos {
		video := &common.Video{
			Id:           tagVideo.VideoID,
			Title:        "标签视频",
			Description:  "这是一个标签相关视频",
			CoverUrl:     "https://example.com/cover.jpg",
			Url:          "https://example.com/video.mp4",
			AuthorId:     1,
			LikeCount:    0,
			CommentCount: 0,
			IsLike:       false,
			PublishTime:  time.Now().Unix(),
		}
		videos = append(videos, video)
	}

	logger.Info("GetTagVideos success",
		logger.StringField("tag", tag),
		logger.IntField("video_count", len(videos)))

	return videos, nil
}

// 获取个性化推荐流
func (s *recommendServiceImpl) GetPersonalizedFeed(ctx context.Context, userID int64, pageSize int32, lastVideoID *int64) ([]*common.Video, int64, error) {
	logger.Info("GetPersonalizedFeed request",
		logger.Int64Field("user_id", userID),
		logger.AnyField("page_size", pageSize),
		logger.AnyField("last_video_id", lastVideoID))

	if pageSize <= 0 || pageSize > 50 {
		pageSize = 20
	}

	preferences, err := s.preferenceRepo.FindByUserID(ctx, userID, 10)
	if err != nil {
		logger.Error("FindByUserID failed", logger.ErrorField(err))
	}

	preferredTags := make([]string, 0, len(preferences))
	for _, pref := range preferences {
		preferredTags = append(preferredTags, pref.TagName)
	}

	popularItems, err := s.actionRepo.GetPopularItems(ctx, "video", 7, int(pageSize))
	if err != nil {
		logger.Error("GetPopularItems failed", logger.ErrorField(err))
		return nil, 0, ErrInternalServer
	}

	videos := make([]*common.Video, 0, len(popularItems))
	var nextLastVideoID int64 = 0

	for _, item := range popularItems {
		video := &common.Video{
			Id:           item.ItemID,
			Title:        "个性化推荐视频",
			Description:  "这是一个个性化推荐视频",
			CoverUrl:     "https://example.com/cover.jpg",
			Url:          "https://example.com/video.mp4",
			AuthorId:     1,
			LikeCount:    int64(item.ActionCount),
			CommentCount: 0,
			IsLike:       false,
			PublishTime:  time.Now().Unix(),
		}
		videos = append(videos, video)
		nextLastVideoID = item.ItemID
	}

	logger.Info("GetPersonalizedFeed success",
		logger.Int64Field("user_id", userID),
		logger.IntField("video_count", len(videos)),
		logger.Int64Field("next_last_video_id", nextLastVideoID))

	return videos, nextLastVideoID, nil
}

// 事务相关
func (s *recommendServiceImpl) WithTransaction(ctx context.Context, fn func(txService RecommendService) error) error {
	return s.actionRepo.WithTransaction(ctx, func(txActionRepo dao.UserActionRepository) error {
		txService := &recommendServiceImpl{
			actionRepo:     txActionRepo,
			videoTagRepo:   s.videoTagRepo,
			preferenceRepo: s.preferenceRepo,
		}
		return fn(txService)
	})
}
