package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"shortvideo/internal/video/dao"
	"shortvideo/internal/video/model"
	"shortvideo/pkg/mq"
	"shortvideo/pkg/storage"
	"time"
)

var (
	ErrVideoNotFound      = errors.New("视频不存在")
	ErrPermissionDenied   = errors.New("权限不足")
	ErrFileUploadFailed   = errors.New("文件上传失败")
	ErrInvalidFile        = errors.New("无效的文件")
	ErrInternalServer     = errors.New("服务器内部错误")
	ErrInvalidVideoStatus = errors.New("无效的视频状态")
)

type VideoService interface {
	//视频上传
	UploadVideo(ctx context.Context, userID int64, videoData []byte, coverData []byte, title, description string) (string, string, error)

	//视频管理
	GetVideoByID(ctx context.Context, videoID int64) (*model.Video, error)
	UpdateVideo(ctx context.Context, videoID int64, userID int64, title, description string) error
	DeleteVideo(ctx context.Context, videoID int64, userID int64) error

	//视频列表
	GetUserVideos(ctx context.Context, userID int64, page, pageSize int) ([]*model.Video, int64, error)
	GetFeedVideos(ctx context.Context, latestTime int64, pageSize int) ([]*model.Video, error)
	SearchVideos(ctx context.Context, keyword string, page, pageSize int) ([]*model.Video, int64, error)
	BatchGetVideos(ctx context.Context, videoIDs []int64) (map[int64]*model.Video, error)

	//视频统计
	GetVideoStats(ctx context.Context, videoID int64) (*model.VideoStats, error)
	UpdateLikeCount(ctx context.Context, videoID int64, delta int64) error
	UpdateCommentCount(ctx context.Context, videoID int64, delta int64) error
	IncrementViewCount(ctx context.Context, videoID int64) error

	//事务支持
	WithTransaction(ctx context.Context, fn func(txService VideoService) error) error
}

type videoServiceImpl struct {
	repo          dao.VideoRepository
	storage       storage.Storage
	kafkaProducer *mq.Producer
}

func NewVideoService(repo dao.VideoRepository, storage storage.Storage, kafkaProducer *mq.Producer) VideoService {
	return &videoServiceImpl{
		repo:          repo,
		storage:       storage,
		kafkaProducer: kafkaProducer,
	}
}

func NewVideoServiceWithRepo(repo dao.VideoRepository) VideoService {
	return &videoServiceImpl{
		repo:          repo,
		storage:       nil,
		kafkaProducer: nil,
	}
}

// 上传视频
func (s *videoServiceImpl) UploadVideo(ctx context.Context, userID int64, videoData []byte, coverData []byte, title, description string) (string, string, error) {
	if len(videoData) == 0 {
		return "", "", ErrInvalidFile
	}

	timestamp := time.Now().Unix()
	videoObjectName := fmt.Sprintf("videos/%d_%d.mp4", userID, timestamp)
	coverObjectName := fmt.Sprintf("covers/%d_%d.jpg", userID, timestamp)

	if s.storage == nil {
		return "", "", ErrFileUploadFailed
	}

	videoReader := bytes.NewReader(videoData)
	videoURL, err := s.storage.Upload(ctx, "", videoObjectName, videoReader, int64(len(videoData)), "video/mp4")
	if err != nil {
		return "", "", ErrFileUploadFailed
	}

	var coverURL string
	if len(coverData) > 0 {
		coverReader := bytes.NewReader(coverData)
		coverURL, err = s.storage.Upload(ctx, "", coverObjectName, coverReader, int64(len(coverData)), "image/jpeg")
		if err != nil {
			return "", "", ErrFileUploadFailed
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
	}

	return videoURL, coverURL, nil
}

// 获取视频详情
func (s *videoServiceImpl) GetVideoByID(ctx context.Context, videoID int64) (*model.Video, error) {
	video, err := s.repo.FindByID(ctx, videoID)
	if err != nil {
		return nil, ErrInternalServer
	}
	if video == nil {
		return nil, ErrVideoNotFound
	}
	return video, nil
}

// 更新视频信息
func (s *videoServiceImpl) UpdateVideo(ctx context.Context, videoID int64, userID int64, title, description string) error {
	video, err := s.repo.FindByID(ctx, videoID)
	if err != nil {
		return ErrInternalServer
	}
	if video == nil {
		return ErrVideoNotFound
	}
	if video.AuthorID != userID {
		return ErrPermissionDenied
	}

	video.Title = title
	video.Description = description

	err = s.repo.Update(ctx, video)
	if err != nil {
		return ErrInternalServer
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
	}

	return nil
}

// 删除视频
func (s *videoServiceImpl) DeleteVideo(ctx context.Context, videoID int64, userID int64) error {
	video, err := s.repo.FindByID(ctx, videoID)
	if err != nil {
		return ErrInternalServer
	}
	if video == nil {
		return ErrVideoNotFound
	}
	if video.AuthorID != userID {
		return ErrPermissionDenied
	}

	err = s.repo.Delete(ctx, videoID, userID)
	if err != nil {
		return ErrInternalServer
	}

	if s.kafkaProducer != nil {
		eventData, _ := json.Marshal(map[string]interface{}{
			"video_id":   videoID,
			"user_id":    userID,
			"deleted_at": time.Now(),
		})
		s.kafkaProducer.SendVideoEvent(ctx, fmt.Sprintf("%d", videoID), eventData)
	}

	return nil
}

// 获取用户视频列表
func (s *videoServiceImpl) GetUserVideos(ctx context.Context, userID int64, page, pageSize int) ([]*model.Video, int64, error) {
	videos, total, err := s.repo.ListByAuthorID(ctx, userID, page, pageSize)
	if err != nil {
		return nil, 0, ErrInternalServer
	}
	return videos, total, nil
}

// 获取视频流
func (s *videoServiceImpl) GetFeedVideos(ctx context.Context, latestTime int64, pageSize int) ([]*model.Video, error) {
	videos, err := s.repo.ListFeedVideos(ctx, latestTime, pageSize)
	if err != nil {
		return nil, ErrInternalServer
	}
	return videos, nil
}

// 搜索视频
func (s *videoServiceImpl) SearchVideos(ctx context.Context, keyword string, page, pageSize int) ([]*model.Video, int64, error) {
	videos, total, err := s.repo.Search(ctx, keyword, page, pageSize)
	if err != nil {
		return nil, 0, ErrInternalServer
	}
	return videos, total, nil
}

// 批量获取视频
func (s *videoServiceImpl) BatchGetVideos(ctx context.Context, videoIDs []int64) (map[int64]*model.Video, error) {
	videos, err := s.repo.BatchGetByIDs(ctx, videoIDs)
	if err != nil {
		return nil, ErrInternalServer
	}
	return videos, nil
}

// 获取视频统计信息
func (s *videoServiceImpl) GetVideoStats(ctx context.Context, videoID int64) (*model.VideoStats, error) {
	_, err := s.repo.FindByID(ctx, videoID)
	if err != nil {
		return nil, ErrInternalServer
	}

	stats, err := s.repo.GetStats(ctx, videoID)
	if err != nil {
		return nil, ErrInternalServer
	}
	return stats, nil
}

// 更新视频点赞数
func (s *videoServiceImpl) UpdateLikeCount(ctx context.Context, videoID int64, delta int64) error {
	_, err := s.repo.FindByID(ctx, videoID)
	if err != nil {
		return ErrInternalServer
	}

	err = s.repo.UpdateLikeCount(ctx, videoID, delta)
	if err != nil {
		return ErrInternalServer
	}

	if s.kafkaProducer != nil {
		eventData, _ := json.Marshal(map[string]interface{}{
			"video_id":   videoID,
			"delta":      delta,
			"updated_at": time.Now(),
		})
		s.kafkaProducer.SendVideoEvent(ctx, fmt.Sprintf("%d", videoID), eventData)
	}

	return nil
}

// 更新视频评论数
func (s *videoServiceImpl) UpdateCommentCount(ctx context.Context, videoID int64, delta int64) error {
	_, err := s.repo.FindByID(ctx, videoID)
	if err != nil {
		return ErrInternalServer
	}

	err = s.repo.UpdateCommentCount(ctx, videoID, delta)
	if err != nil {
		return ErrInternalServer
	}

	if s.kafkaProducer != nil {
		eventData, _ := json.Marshal(map[string]interface{}{
			"video_id":   videoID,
			"delta":      delta,
			"updated_at": time.Now(),
		})
		s.kafkaProducer.SendVideoEvent(ctx, fmt.Sprintf("%d", videoID), eventData)
	}

	return nil
}

// 增加视频观看数
func (s *videoServiceImpl) IncrementViewCount(ctx context.Context, videoID int64) error {
	_, err := s.repo.FindByID(ctx, videoID)
	if err != nil {
		return ErrInternalServer
	}

	err = s.repo.IncrementViewCount(ctx, videoID)
	if err != nil {
		return ErrInternalServer
	}

	return nil
}

// 事务支持
func (s *videoServiceImpl) WithTransaction(ctx context.Context, fn func(txService VideoService) error) error {
	return s.repo.WithTransaction(ctx, func(txRepo dao.VideoRepository) error {
		txService := NewVideoServiceWithRepo(txRepo)
		return fn(txService)
	})
}
