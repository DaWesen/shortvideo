package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"shortvideo/internal/video/dao"
	"shortvideo/internal/video/model"
	"shortvideo/pkg/cache"
	"shortvideo/pkg/logger"
	"shortvideo/pkg/mq"
	"shortvideo/pkg/storage"
	"time"
)

var (
	ErrVideoNotFound     = errors.New("视频不存在")
	ErrNotVideoOwner     = errors.New("不是视频所有者")
	ErrInternalServer    = errors.New("服务器内部错误")
	ErrInvalidVideoData  = errors.New("无效的视频数据")
	ErrVideoUploadFailed = errors.New("视频上传失败")
	ErrCoverUploadFailed = errors.New("封面上传失败")
	ErrInvalidFile       = errors.New("无效的文件")
)

type VideoService interface {
	//视频发布
	PublishVideo(ctx context.Context, userID int64, title, videoURL, coverURL, description string) (int64, error)

	//视频上传
	UploadVideo(ctx context.Context, userID int64, videoData []byte, coverData []byte, title, description string) (string, string, error)

	//视频详情
	GetVideoByID(ctx context.Context, videoID, currentUserID int64) (*model.Video, error)

	//用户视频列表
	GetUserVideos(ctx context.Context, userID, currentUserID int64, page, pageSize int) ([]*model.Video, int64, error)

	//视频流
	GetFeedVideos(ctx context.Context, currentUserID int64, latestTime int64, pageSize int) ([]*model.Video, int64, error)

	//搜索视频
	SearchVideos(ctx context.Context, keyword string, currentUserID int64, page, pageSize int) ([]*model.Video, int64, error)

	//批量获取视频
	BatchGetVideosByIDs(ctx context.Context, videoIDs []int64, currentUserID int64) (map[int64]*model.Video, error)

	//删除视频
	DeleteVideo(ctx context.Context, videoID, userID int64) error

	//更新视频信息
	UpdateVideo(ctx context.Context, videoID, userID int64, title, description string) error

	//视频统计
	GetVideoStats(ctx context.Context, videoID int64) (*model.VideoStats, error)
	UpdateLikeCount(ctx context.Context, videoID int64, delta int64) error
	UpdateCommentCount(ctx context.Context, videoID int64, delta int64) error
	UpdateShareCount(ctx context.Context, videoID int64, delta int64) error
	IncrementViewCount(ctx context.Context, videoID int64) error

	//热门视频
	GetHotVideos(ctx context.Context, currentUserID int64, pageSize int) ([]*model.Video, error)

	//统计相关
	CountVideosByUserID(ctx context.Context, userID int64) (int64, error)
	GetTotalVideoCount(ctx context.Context) (int64, error)

	//事务相关
	WithTransaction(ctx context.Context, fn func(txService VideoService) error) error
}

type videoServiceImpl struct {
	repo          dao.VideoRepository
	storage       storage.Storage
	kafkaProducer *mq.Producer
	cache         cache.Cache
}

func NewVideoService(repo dao.VideoRepository, storage storage.Storage, kafkaProducer *mq.Producer, cache cache.Cache) VideoService {
	return &videoServiceImpl{
		repo:          repo,
		storage:       storage,
		kafkaProducer: kafkaProducer,
		cache:         cache,
	}
}

func NewVideoServiceWithRepo(repo dao.VideoRepository) VideoService {
	return &videoServiceImpl{
		repo: repo,
	}
}

// 上传视频
func (s *videoServiceImpl) UploadVideo(ctx context.Context, userID int64, videoData []byte, coverData []byte, title, description string) (string, string, error) {
	logger.Info("上传视频请求",
		logger.Int64Field("user_id", userID),
		logger.StringField("title", title))

	if len(videoData) == 0 {
		logger.Warn("无效的视频文件",
			logger.Int64Field("user_id", userID))
		return "", "", ErrInvalidFile
	}

	timestamp := time.Now().Unix()
	videoObjectName := fmt.Sprintf("videos/%d_%d.mp4", userID, timestamp)
	coverObjectName := fmt.Sprintf("covers/%d_%d.jpg", userID, timestamp)

	if s.storage == nil {
		logger.Error("存储服务未初始化",
			logger.Int64Field("user_id", userID))
		return "", "", ErrVideoUploadFailed
	}

	videoReader := bytes.NewReader(videoData)
	videoURL, err := s.storage.Upload(ctx, "", videoObjectName, videoReader, int64(len(videoData)), "video/mp4")
	if err != nil {
		logger.Error("视频上传失败",
			logger.ErrorField(err),
			logger.Int64Field("user_id", userID))
		return "", "", ErrVideoUploadFailed
	}

	var coverURL string
	if len(coverData) > 0 {
		coverReader := bytes.NewReader(coverData)
		coverURL, err = s.storage.Upload(ctx, "", coverObjectName, coverReader, int64(len(coverData)), "image/jpeg")
		if err != nil {
			logger.Error("封面上传失败",
				logger.ErrorField(err),
				logger.Int64Field("user_id", userID))
			return "", "", ErrCoverUploadFailed
		}
	}

	video := &model.Video{
		AuthorID:    userID,
		URL:         videoURL,
		CoverURL:    coverURL,
		Title:       title,
		Description: description,
		PublishTime: time.Now().Unix(),
	}

	err = s.repo.Create(ctx, video)
	if err != nil {
		logger.Error("创建视频记录失败",
			logger.ErrorField(err),
			logger.Int64Field("user_id", userID),
			logger.StringField("title", title))
		return "", "", ErrInternalServer
	}

	if s.kafkaProducer != nil {
		eventData, _ := json.Marshal(map[string]interface{}{
			"video_id":    video.ID,
			"user_id":     userID,
			"video_url":   videoURL,
			"cover_url":   coverURL,
			"title":       title,
			"uploaded_at": time.Now(),
		})
		s.kafkaProducer.SendVideoEvent(ctx, fmt.Sprintf("%d", video.ID), eventData)
		logger.Info("发送视频上传事件",
			logger.Int64Field("video_id", video.ID),
			logger.Int64Field("user_id", userID))
	}

	logger.Info("视频上传成功",
		logger.Int64Field("user_id", userID),
		logger.StringField("title", title),
		logger.StringField("video_url", videoURL))

	return videoURL, coverURL, nil
}

// 获取视频详情
func (s *videoServiceImpl) GetVideoByID(ctx context.Context, videoID, currentUserID int64) (*model.Video, error) {
	logger.Info("获取视频详情请求",
		logger.Int64Field("video_id", videoID),
		logger.Int64Field("current_user_id", currentUserID))

	if s.cache != nil {
		videoKey := fmt.Sprintf("video:%d", videoID)
		cachedVideo, err := s.cache.Get(ctx, videoKey)
		if err == nil && cachedVideo != "" {
			var video model.Video
			if err := json.Unmarshal([]byte(cachedVideo), &video); err == nil {
				logger.Info("从缓存获取视频信息成功",
					logger.Int64Field("video_id", videoID))
				return &video, nil
			}
		}
	}

	video, err := s.repo.FindByID(ctx, videoID)
	if err != nil {
		logger.Error("查询视频失败",
			logger.ErrorField(err),
			logger.Int64Field("video_id", videoID))
		return nil, ErrInternalServer
	}
	if video == nil {
		logger.Warn("视频不存在",
			logger.Int64Field("video_id", videoID))
		return nil, ErrVideoNotFound
	}

	go func() {
		s.IncrementViewCount(context.Background(), videoID)
	}()

	if s.cache != nil {
		videoKey := fmt.Sprintf("video:%d", videoID)
		videoData, err := json.Marshal(video)
		if err == nil {
			s.cache.Set(ctx, videoKey, string(videoData), 5*time.Minute)
			logger.Info("视频信息存入缓存成功",
				logger.Int64Field("video_id", videoID))
		}
	}

	logger.Info("获取视频详情成功",
		logger.Int64Field("video_id", videoID),
		logger.StringField("title", video.Title))

	return video, nil
}

// 更新视频信息
func (s *videoServiceImpl) UpdateVideo(ctx context.Context, videoID int64, userID int64, title, description string) error {
	logger.Info("更新视频信息请求",
		logger.Int64Field("video_id", videoID),
		logger.Int64Field("user_id", userID),
		logger.StringField("title", title))

	video, err := s.repo.FindByID(ctx, videoID)
	if err != nil {
		logger.Error("查询视频失败",
			logger.ErrorField(err),
			logger.Int64Field("video_id", videoID),
			logger.Int64Field("user_id", userID))
		return ErrInternalServer
	}
	if video == nil {
		logger.Warn("视频不存在",
			logger.Int64Field("video_id", videoID),
			logger.Int64Field("user_id", userID))
		return ErrVideoNotFound
	}
	if video.AuthorID != userID {
		logger.Warn("不是视频所有者",
			logger.Int64Field("video_id", videoID),
			logger.Int64Field("user_id", userID),
			logger.Int64Field("actual_author_id", video.AuthorID))
		return ErrNotVideoOwner
	}

	video.Title = title
	video.Description = description

	err = s.repo.Update(ctx, video)
	if err != nil {
		logger.Error("更新视频失败",
			logger.ErrorField(err),
			logger.Int64Field("video_id", videoID),
			logger.Int64Field("user_id", userID))
		return ErrInternalServer
	}

	if s.cache != nil {
		videoKey := fmt.Sprintf("video:%d", videoID)
		s.cache.Delete(ctx, videoKey)
		logger.Info("删除视频缓存成功",
			logger.Int64Field("video_id", videoID))
	}

	if s.kafkaProducer != nil {
		eventData, _ := json.Marshal(map[string]interface{}{
			"video_id":    videoID,
			"user_id":     userID,
			"title":       title,
			"description": description,
			"updated_at":  time.Now(),
		})
		s.kafkaProducer.SendVideoEvent(ctx, fmt.Sprintf("%d", videoID), eventData)
		logger.Info("发送视频更新事件",
			logger.Int64Field("video_id", videoID),
			logger.Int64Field("user_id", userID))
	}

	logger.Info("更新视频信息成功",
		logger.Int64Field("video_id", videoID),
		logger.Int64Field("user_id", userID),
		logger.StringField("title", title))

	return nil
}

// 删除视频
func (s *videoServiceImpl) DeleteVideo(ctx context.Context, videoID int64, userID int64) error {
	logger.Info("删除视频请求",
		logger.Int64Field("video_id", videoID),
		logger.Int64Field("user_id", userID))

	video, err := s.repo.FindByID(ctx, videoID)
	if err != nil {
		logger.Error("查询视频失败",
			logger.ErrorField(err),
			logger.Int64Field("video_id", videoID),
			logger.Int64Field("user_id", userID))
		return ErrInternalServer
	}
	if video == nil {
		logger.Warn("视频不存在",
			logger.Int64Field("video_id", videoID),
			logger.Int64Field("user_id", userID))
		return ErrVideoNotFound
	}
	if video.AuthorID != userID {
		logger.Warn("不是视频所有者",
			logger.Int64Field("video_id", videoID),
			logger.Int64Field("user_id", userID),
			logger.Int64Field("actual_author_id", video.AuthorID))
		return ErrNotVideoOwner
	}

	err = s.repo.Delete(ctx, videoID, userID)
	if err != nil {
		logger.Error("删除视频失败",
			logger.ErrorField(err),
			logger.Int64Field("video_id", videoID),
			logger.Int64Field("user_id", userID))
		return ErrInternalServer
	}

	if s.cache != nil {
		videoKey := fmt.Sprintf("video:%d", videoID)
		s.cache.Delete(ctx, videoKey)
		logger.Info("删除视频缓存成功",
			logger.Int64Field("video_id", videoID))
	}

	if s.kafkaProducer != nil {
		eventData, _ := json.Marshal(map[string]interface{}{
			"video_id":    videoID,
			"user_id":     userID,
			"video_title": video.Title,
			"deleted_at":  time.Now(),
		})
		s.kafkaProducer.SendVideoEvent(ctx, fmt.Sprintf("%d", videoID), eventData)
		logger.Info("发送视频删除事件",
			logger.Int64Field("video_id", videoID),
			logger.Int64Field("user_id", userID))
	}

	logger.Info("删除视频成功",
		logger.Int64Field("video_id", videoID),
		logger.Int64Field("user_id", userID),
		logger.StringField("video_title", video.Title))

	return nil
}

// 获取用户视频列表
func (s *videoServiceImpl) GetUserVideos(ctx context.Context, userID, currentUserID int64, page, pageSize int) ([]*model.Video, int64, error) {
	logger.Info("获取用户视频列表请求",
		logger.Int64Field("user_id", userID),
		logger.Int64Field("current_user_id", currentUserID),
		logger.IntField("page", page),
		logger.IntField("page_size", pageSize))

	videos, total, err := s.repo.ListByAuthorID(ctx, userID, page, pageSize)
	if err != nil {
		logger.Error("查询用户视频失败",
			logger.ErrorField(err),
			logger.Int64Field("user_id", userID),
			logger.Int64Field("current_user_id", currentUserID))
		return nil, 0, ErrInternalServer
	}

	logger.Info("获取用户视频列表成功",
		logger.Int64Field("user_id", userID),
		logger.IntField("video_count", len(videos)),
		logger.Int64Field("total_count", total))

	return videos, total, nil
}

// 获取视频流
func (s *videoServiceImpl) GetFeedVideos(ctx context.Context, currentUserID int64, latestTime int64, pageSize int) ([]*model.Video, int64, error) {
	logger.Info("获取视频流请求",
		logger.Int64Field("current_user_id", currentUserID),
		logger.Int64Field("latest_time", latestTime),
		logger.IntField("page_size", pageSize))

	videos, err := s.repo.ListFeedVideos(ctx, latestTime, pageSize)
	if err != nil {
		logger.Error("查询视频流失败",
			logger.ErrorField(err),
			logger.Int64Field("current_user_id", currentUserID))
		return nil, 0, ErrInternalServer
	}

	nextTime := time.Now().Unix()
	if len(videos) > 0 {
		nextTime = videos[len(videos)-1].PublishTime
	}

	logger.Info("获取视频流成功",
		logger.Int64Field("current_user_id", currentUserID),
		logger.IntField("video_count", len(videos)),
		logger.Int64Field("next_time", nextTime))

	return videos, nextTime, nil
}

// 搜索视频
func (s *videoServiceImpl) SearchVideos(ctx context.Context, keyword string, currentUserID int64, page, pageSize int) ([]*model.Video, int64, error) {
	logger.Info("搜索视频请求",
		logger.StringField("keyword", keyword),
		logger.Int64Field("current_user_id", currentUserID),
		logger.IntField("page", page),
		logger.IntField("page_size", pageSize))

	videos, total, err := s.repo.Search(ctx, keyword, page, pageSize)
	if err != nil {
		logger.Error("搜索视频失败",
			logger.ErrorField(err),
			logger.StringField("keyword", keyword),
			logger.Int64Field("current_user_id", currentUserID))
		return nil, 0, ErrInternalServer
	}

	logger.Info("搜索视频成功",
		logger.StringField("keyword", keyword),
		logger.Int64Field("current_user_id", currentUserID),
		logger.IntField("video_count", len(videos)),
		logger.Int64Field("total_count", total))

	return videos, total, nil
}

// 批量获取视频
func (s *videoServiceImpl) BatchGetVideosByIDs(ctx context.Context, videoIDs []int64, currentUserID int64) (map[int64]*model.Video, error) {
	logger.Info("批量获取视频请求",
		logger.AnyField("video_ids", videoIDs),
		logger.Int64Field("current_user_id", currentUserID))

	videos, err := s.repo.BatchGetByIDs(ctx, videoIDs)
	if err != nil {
		logger.Error("批量查询视频失败",
			logger.ErrorField(err),
			logger.AnyField("video_ids", videoIDs),
			logger.Int64Field("current_user_id", currentUserID))
		return nil, ErrInternalServer
	}

	logger.Info("批量获取视频成功",
		logger.Int64Field("current_user_id", currentUserID),
		logger.IntField("video_count", len(videos)))

	return videos, nil
}

// 获取视频统计信息
func (s *videoServiceImpl) GetVideoStats(ctx context.Context, videoID int64) (*model.VideoStats, error) {
	logger.Info("获取视频统计数据请求",
		logger.Int64Field("video_id", videoID))

	_, err := s.repo.FindByID(ctx, videoID)
	if err != nil {
		logger.Error("查询视频失败",
			logger.ErrorField(err),
			logger.Int64Field("video_id", videoID))
		return nil, ErrInternalServer
	}

	stats, err := s.repo.GetStats(ctx, videoID)
	if err != nil {
		logger.Error("查询视频统计数据失败",
			logger.ErrorField(err),
			logger.Int64Field("video_id", videoID))
		return nil, ErrInternalServer
	}

	logger.Info("获取视频统计数据成功",
		logger.Int64Field("video_id", videoID),
		logger.Int64Field("like_count", stats.LikeCount),
		logger.Int64Field("comment_count", stats.CommentCount))

	return stats, nil
}

// 更新视频点赞数
func (s *videoServiceImpl) UpdateLikeCount(ctx context.Context, videoID int64, delta int64) error {
	logger.Info("更新视频点赞数请求",
		logger.Int64Field("video_id", videoID),
		logger.Int64Field("delta", delta))

	_, err := s.repo.FindByID(ctx, videoID)
	if err != nil {
		logger.Error("查询视频失败",
			logger.ErrorField(err),
			logger.Int64Field("video_id", videoID))
		return ErrInternalServer
	}

	err = s.repo.UpdateLikeCount(ctx, videoID, delta)
	if err != nil {
		logger.Error("更新视频点赞数失败",
			logger.ErrorField(err),
			logger.Int64Field("video_id", videoID),
			logger.Int64Field("delta", delta))
		return ErrInternalServer
	}

	if s.cache != nil {
		videoKey := fmt.Sprintf("video:%d", videoID)
		s.cache.Delete(ctx, videoKey)
		logger.Info("删除视频缓存成功",
			logger.Int64Field("video_id", videoID))
	}

	if s.kafkaProducer != nil {
		eventData, _ := json.Marshal(map[string]interface{}{
			"video_id":   videoID,
			"delta":      delta,
			"updated_at": time.Now(),
		})
		s.kafkaProducer.SendVideoEvent(ctx, fmt.Sprintf("%d", videoID), eventData)
		logger.Info("发送视频点赞事件",
			logger.Int64Field("video_id", videoID),
			logger.Int64Field("delta", delta))
	}

	logger.Info("更新视频点赞数成功",
		logger.Int64Field("video_id", videoID),
		logger.Int64Field("delta", delta))

	return nil
}

// 更新视频评论数
func (s *videoServiceImpl) UpdateCommentCount(ctx context.Context, videoID int64, delta int64) error {
	logger.Info("更新视频评论数请求",
		logger.Int64Field("video_id", videoID),
		logger.Int64Field("delta", delta))

	_, err := s.repo.FindByID(ctx, videoID)
	if err != nil {
		logger.Error("查询视频失败",
			logger.ErrorField(err),
			logger.Int64Field("video_id", videoID))
		return ErrInternalServer
	}

	err = s.repo.UpdateCommentCount(ctx, videoID, delta)
	if err != nil {
		logger.Error("更新视频评论数失败",
			logger.ErrorField(err),
			logger.Int64Field("video_id", videoID),
			logger.Int64Field("delta", delta))
		return ErrInternalServer
	}

	if s.cache != nil {
		videoKey := fmt.Sprintf("video:%d", videoID)
		s.cache.Delete(ctx, videoKey)
		logger.Info("删除视频缓存成功",
			logger.Int64Field("video_id", videoID))
	}

	if s.kafkaProducer != nil {
		eventData, _ := json.Marshal(map[string]interface{}{
			"video_id":   videoID,
			"delta":      delta,
			"updated_at": time.Now(),
		})
		s.kafkaProducer.SendVideoEvent(ctx, fmt.Sprintf("%d", videoID), eventData)
		logger.Info("发送视频评论事件",
			logger.Int64Field("video_id", videoID),
			logger.Int64Field("delta", delta))
	}

	logger.Info("更新视频评论数成功",
		logger.Int64Field("video_id", videoID),
		logger.Int64Field("delta", delta))

	return nil
}

// 增加视频观看数
func (s *videoServiceImpl) IncrementViewCount(ctx context.Context, videoID int64) error {
	logger.Info("增加视频观看次数请求",
		logger.Int64Field("video_id", videoID))

	_, err := s.repo.FindByID(ctx, videoID)
	if err != nil {
		logger.Error("查询视频失败",
			logger.ErrorField(err),
			logger.Int64Field("video_id", videoID))
		return ErrInternalServer
	}

	err = s.repo.IncrementViewCount(ctx, videoID)
	if err != nil {
		logger.Error("增加视频观看次数失败",
			logger.ErrorField(err),
			logger.Int64Field("video_id", videoID))
		return ErrInternalServer
	}

	logger.Info("增加视频观看次数成功",
		logger.Int64Field("video_id", videoID))

	return nil
}

// 事务支持
func (s *videoServiceImpl) WithTransaction(ctx context.Context, fn func(txService VideoService) error) error {
	return s.repo.WithTransaction(ctx, func(txRepo dao.VideoRepository) error {
		txService := NewVideoServiceWithRepo(txRepo)
		return fn(txService)
	})
}

// 发布视频
func (s *videoServiceImpl) PublishVideo(ctx context.Context, userID int64, title, videoURL, coverURL, description string) (int64, error) {
	logger.Info("发布视频请求",
		logger.Int64Field("user_id", userID),
		logger.StringField("title", title))

	if title == "" || videoURL == "" {
		logger.Warn("无效的视频数据",
			logger.Int64Field("user_id", userID))
		return 0, ErrInvalidVideoData
	}

	video := &model.Video{
		AuthorID:     userID,
		Title:        title,
		URL:          videoURL,
		CoverURL:     coverURL,
		Description:  description,
		PublishTime:  time.Now().Unix(),
		LikeCount:    0,
		CommentCount: 0,
	}

	if err := s.repo.Create(ctx, video); err != nil {
		logger.Error("创建视频失败",
			logger.ErrorField(err),
			logger.Int64Field("user_id", userID),
			logger.StringField("title", title))
		return 0, ErrInternalServer
	}

	if s.kafkaProducer != nil {
		eventData, _ := json.Marshal(map[string]interface{}{
			"video_id":     video.ID,
			"user_id":      userID,
			"title":        title,
			"video_url":    videoURL,
			"cover_url":    coverURL,
			"published_at": time.Now(),
		})
		s.kafkaProducer.SendVideoEvent(ctx, fmt.Sprintf("%d", video.ID), eventData)
		logger.Info("发送视频发布事件",
			logger.Int64Field("video_id", video.ID),
			logger.Int64Field("user_id", userID))
	}

	logger.Info("视频发布成功",
		logger.Int64Field("video_id", video.ID),
		logger.Int64Field("user_id", userID),
		logger.StringField("title", title))

	return video.ID, nil
}

// 更新视频分享数
func (s *videoServiceImpl) UpdateShareCount(ctx context.Context, videoID int64, delta int64) error {
	logger.Info("更新视频分享数请求",
		logger.Int64Field("video_id", videoID),
		logger.Int64Field("delta", delta))

	_, err := s.repo.FindByID(ctx, videoID)
	if err != nil {
		logger.Error("查询视频失败",
			logger.ErrorField(err),
			logger.Int64Field("video_id", videoID))
		return ErrInternalServer
	}

	err = s.repo.UpdateShareCount(ctx, videoID, delta)
	if err != nil {
		logger.Error("更新视频分享数失败",
			logger.ErrorField(err),
			logger.Int64Field("video_id", videoID),
			logger.Int64Field("delta", delta))
		return ErrInternalServer
	}

	if s.cache != nil {
		videoKey := fmt.Sprintf("video:%d", videoID)
		s.cache.Delete(ctx, videoKey)
		logger.Info("删除视频缓存成功",
			logger.Int64Field("video_id", videoID))
	}

	if s.kafkaProducer != nil {
		eventData, _ := json.Marshal(map[string]interface{}{
			"video_id":   videoID,
			"delta":      delta,
			"updated_at": time.Now(),
		})
		s.kafkaProducer.SendVideoEvent(ctx, fmt.Sprintf("%d", videoID), eventData)
		logger.Info("发送视频分享事件",
			logger.Int64Field("video_id", videoID),
			logger.Int64Field("delta", delta))
	}

	logger.Info("更新视频分享数成功",
		logger.Int64Field("video_id", videoID),
		logger.Int64Field("delta", delta))

	return nil
}

// 获取热门视频
func (s *videoServiceImpl) GetHotVideos(ctx context.Context, currentUserID int64, pageSize int) ([]*model.Video, error) {
	logger.Info("获取热门视频请求",
		logger.Int64Field("current_user_id", currentUserID),
		logger.IntField("page_size", pageSize))

	videos, err := s.repo.ListFeedVideos(ctx, 0, pageSize)
	if err != nil {
		logger.Error("查询热门视频失败",
			logger.ErrorField(err),
			logger.Int64Field("current_user_id", currentUserID))
		return nil, ErrInternalServer
	}

	logger.Info("获取热门视频成功",
		logger.Int64Field("current_user_id", currentUserID),
		logger.IntField("video_count", len(videos)))

	return videos, nil
}

// 获取用户视频总数
func (s *videoServiceImpl) CountVideosByUserID(ctx context.Context, userID int64) (int64, error) {
	logger.Info("获取用户视频总数请求",
		logger.Int64Field("user_id", userID))

	count, err := s.repo.CountByAuthorID(ctx, userID)
	if err != nil {
		logger.Error("查询用户视频总数失败",
			logger.ErrorField(err),
			logger.Int64Field("user_id", userID))
		return 0, ErrInternalServer
	}

	logger.Info("获取用户视频总数成功",
		logger.Int64Field("user_id", userID),
		logger.Int64Field("count", count))

	return count, nil
}

// 获取总视频数
func (s *videoServiceImpl) GetTotalVideoCount(ctx context.Context) (int64, error) {
	logger.Info("获取总视频数请求")

	count, err := s.repo.GetTotalVideoCount(ctx)
	if err != nil {
		logger.Error("查询总视频数失败",
			logger.ErrorField(err))
		return 0, ErrInternalServer
	}

	logger.Info("获取总视频数成功",
		logger.Int64Field("count", count))

	return count, nil
}
