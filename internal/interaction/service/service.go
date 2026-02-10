package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"shortvideo/internal/interaction/dao"
	"shortvideo/internal/interaction/model"
	videoModel "shortvideo/internal/video/model"
	"shortvideo/internal/video/service"
	"shortvideo/pkg/logger"
	"shortvideo/pkg/mq"
	"time"
)

var (
	ErrInteractionFailed     = errors.New("互动操作失败")
	ErrCommentNotFound       = errors.New("评论不存在")
	ErrNotCommentOwner       = errors.New("不是评论所有者")
	ErrInvalidCommentContent = errors.New("无效的评论内容")
	ErrInternalServer        = errors.New("服务器内部错误")
	ErrVideoNotFound         = errors.New("视频不存在")
	ErrAlreadyLiked          = errors.New("已经点赞过")
	ErrAlreadyStarred        = errors.New("已经收藏过")
	ErrNotLiked              = errors.New("未点赞")
	ErrNotStarred            = errors.New("未收藏")
)

type InteractionService interface {
	//点赞
	LikeAction(ctx context.Context, userID, videoID int64, action bool) error
	GetLikeVideoList(ctx context.Context, userID, currentUserID int64, page, pageSize int) ([]*videoModel.Video, int64, error)
	CheckLikeStatus(ctx context.Context, userID, videoID int64) (bool, error)
	//收藏
	StarAction(ctx context.Context, userID, videoID int64, action bool) error
	GetStarVideoList(ctx context.Context, userID, currentUserID int64, page, pageSize int) ([]*videoModel.Video, int64, error)
	CheckStarStatus(ctx context.Context, userID, videoID int64) (bool, error)
	//评论
	CommentAction(ctx context.Context, userID, videoID int64, content string, replyToID int64) (*model.Comment, error)
	GetCommentList(ctx context.Context, videoID, currentUserID int64, page, pageSize int) ([]*model.Comment, int64, error)
	DeleteComment(ctx context.Context, userID, videoID, commentID int64) error
	//分享操作
	ShareAction(ctx context.Context, userID, videoID int64) error
	//获取互动统计
	GetCount(ctx context.Context, videoID int64) (int64, int64, int64, int64, error)
	//事务支持
	WithTransaction(ctx context.Context, fn func(txService InteractionService) error) error
}

type interactionServiceImpl struct {
	likeRepo      dao.LikeRepository
	starRepo      dao.StarRepository
	commentRepo   dao.CommentRepository
	shareRepo     dao.ShareRepository
	statsRepo     dao.VideoInteractionStatsRepository
	videoService  service.VideoService
	kafkaProducer *mq.Producer
}

func NewInteractionService(
	likeRepo dao.LikeRepository,
	starRepo dao.StarRepository,
	commentRepo dao.CommentRepository,
	shareRepo dao.ShareRepository,
	statsRepo dao.VideoInteractionStatsRepository,
	videoService service.VideoService,
	kafkaProducer *mq.Producer,
) InteractionService {
	return &interactionServiceImpl{
		likeRepo:      likeRepo,
		starRepo:      starRepo,
		commentRepo:   commentRepo,
		shareRepo:     shareRepo,
		statsRepo:     statsRepo,
		videoService:  videoService,
		kafkaProducer: kafkaProducer,
	}
}

// 点赞操作
func (s *interactionServiceImpl) LikeAction(ctx context.Context, userID, videoID int64, action bool) error {
	logger.Info("点赞操作请求",
		logger.Int64Field("user_id", userID),
		logger.Int64Field("video_id", videoID),
		logger.BoolField("action", action))

	exists, err := s.likeRepo.Exists(ctx, userID, videoID)
	if err != nil {
		logger.Error("检查点赞状态失败",
			logger.ErrorField(err),
			logger.Int64Field("user_id", userID),
			logger.Int64Field("video_id", videoID))
		return ErrInternalServer
	}

	if action {
		if exists {
			return ErrAlreadyLiked
		}

		like := &model.Like{
			UserID:  userID,
			VideoID: videoID,
		}

		if err := s.likeRepo.Create(ctx, like); err != nil {
			logger.Error("创建点赞记录失败",
				logger.ErrorField(err),
				logger.Int64Field("user_id", userID),
				logger.Int64Field("video_id", videoID))
			return ErrInteractionFailed
		}

		if err := s.statsRepo.IncrementLikeCount(ctx, videoID); err != nil {
			logger.Error("增加视频点赞数失败",
				logger.ErrorField(err),
				logger.Int64Field("video_id", videoID))
		}
	} else {
		if !exists {
			return ErrNotLiked
		}

		if err := s.likeRepo.Delete(ctx, userID, videoID); err != nil {
			logger.Error("删除点赞记录失败",
				logger.ErrorField(err),
				logger.Int64Field("user_id", userID),
				logger.Int64Field("video_id", videoID))
			return ErrInteractionFailed
		}

		if err := s.statsRepo.DecrementLikeCount(ctx, videoID); err != nil {
			logger.Error("减少视频点赞数失败",
				logger.ErrorField(err),
				logger.Int64Field("video_id", videoID))
		}
	}

	if s.kafkaProducer != nil {
		eventData := map[string]interface{}{
			"user_id":    userID,
			"video_id":   videoID,
			"action":     action,
			"created_at": time.Now(),
		}
		data, _ := json.Marshal(eventData)
		s.kafkaProducer.SendInteractionEvent(ctx, fmt.Sprintf("%d", userID), data)
	}

	logger.Info("点赞操作成功",
		logger.Int64Field("user_id", userID),
		logger.Int64Field("video_id", videoID),
		logger.BoolField("action", action))

	return nil
}

// 获取用户点赞视频列表
func (s *interactionServiceImpl) GetLikeVideoList(ctx context.Context, userID, currentUserID int64, page, pageSize int) ([]*videoModel.Video, int64, error) {
	logger.Info("获取用户点赞视频列表请求",
		logger.Int64Field("user_id", userID),
		logger.Int64Field("current_user_id", currentUserID),
		logger.IntField("page", page),
		logger.IntField("page_size", pageSize))

	likes, total, err := s.likeRepo.ListByUserID(ctx, userID, page, pageSize)
	if err != nil {
		logger.Error("获取用户点赞记录失败",
			logger.ErrorField(err),
			logger.Int64Field("user_id", userID))
		return nil, 0, ErrInternalServer
	}

	if len(likes) == 0 {
		return []*videoModel.Video{}, total, nil
	}

	videoIDs := make([]int64, len(likes))
	for i, like := range likes {
		videoIDs[i] = like.VideoID
	}

	videoMap, err := s.videoService.BatchGetVideosByIDs(ctx, videoIDs, currentUserID)
	if err != nil {
		logger.Error("批量获取视频信息失败",
			logger.ErrorField(err),
			logger.Int64Field("user_id", userID))
		return nil, 0, ErrInternalServer
	}

	videos := make([]*videoModel.Video, 0, len(likes))
	for _, like := range likes {
		if video, ok := videoMap[like.VideoID]; ok {
			videos = append(videos, video)
		}
	}

	logger.Info("获取用户点赞视频列表成功",
		logger.Int64Field("user_id", userID),
		logger.IntField("video_count", len(videos)),
		logger.Int64Field("total_count", total))

	return videos, total, nil
}

// 收藏操作
func (s *interactionServiceImpl) StarAction(ctx context.Context, userID, videoID int64, action bool) error {
	logger.Info("收藏操作请求",
		logger.Int64Field("user_id", userID),
		logger.Int64Field("video_id", videoID),
		logger.BoolField("action", action))

	exists, err := s.starRepo.Exists(ctx, userID, videoID)
	if err != nil {
		logger.Error("检查收藏状态失败",
			logger.ErrorField(err),
			logger.Int64Field("user_id", userID),
			logger.Int64Field("video_id", videoID))
		return ErrInternalServer
	}

	if action {
		if exists {
			return ErrAlreadyStarred
		}

		star := &model.Star{
			UserID:  userID,
			VideoID: videoID,
		}

		if err := s.starRepo.Create(ctx, star); err != nil {
			logger.Error("创建收藏记录失败",
				logger.ErrorField(err),
				logger.Int64Field("user_id", userID),
				logger.Int64Field("video_id", videoID))
			return ErrInteractionFailed
		}

		if err := s.statsRepo.IncrementStarCount(ctx, videoID); err != nil {
			logger.Error("增加视频收藏数失败",
				logger.ErrorField(err),
				logger.Int64Field("video_id", videoID))
		}
	} else {
		if !exists {
			return ErrNotStarred
		}

		if err := s.starRepo.Delete(ctx, userID, videoID); err != nil {
			logger.Error("删除收藏记录失败",
				logger.ErrorField(err),
				logger.Int64Field("user_id", userID),
				logger.Int64Field("video_id", videoID))
			return ErrInteractionFailed
		}

		if err := s.statsRepo.DecrementStarCount(ctx, videoID); err != nil {
			logger.Error("减少视频收藏数失败",
				logger.ErrorField(err),
				logger.Int64Field("video_id", videoID))
		}
	}

	if s.kafkaProducer != nil {
		eventData := map[string]interface{}{
			"user_id":    userID,
			"video_id":   videoID,
			"action":     action,
			"created_at": time.Now(),
		}
		data, _ := json.Marshal(eventData)
		s.kafkaProducer.SendInteractionEvent(ctx, fmt.Sprintf("%d", userID), data)
	}

	logger.Info("收藏操作成功",
		logger.Int64Field("user_id", userID),
		logger.Int64Field("video_id", videoID),
		logger.BoolField("action", action))

	return nil
}

// 获取用户收藏视频列表
func (s *interactionServiceImpl) GetStarVideoList(ctx context.Context, userID, currentUserID int64, page, pageSize int) ([]*videoModel.Video, int64, error) {
	logger.Info("获取用户收藏视频列表请求",
		logger.Int64Field("user_id", userID),
		logger.Int64Field("current_user_id", currentUserID),
		logger.IntField("page", page),
		logger.IntField("page_size", pageSize))

	stars, total, err := s.starRepo.ListByUserID(ctx, userID, page, pageSize)
	if err != nil {
		logger.Error("获取用户收藏记录失败",
			logger.ErrorField(err),
			logger.Int64Field("user_id", userID))
		return nil, 0, ErrInternalServer
	}

	if len(stars) == 0 {
		return []*videoModel.Video{}, total, nil
	}

	videoIDs := make([]int64, len(stars))
	for i, star := range stars {
		videoIDs[i] = star.VideoID
	}

	videoMap, err := s.videoService.BatchGetVideosByIDs(ctx, videoIDs, currentUserID)
	if err != nil {
		logger.Error("批量获取视频信息失败",
			logger.ErrorField(err),
			logger.Int64Field("user_id", userID))
		return nil, 0, ErrInternalServer
	}

	videos := make([]*videoModel.Video, 0, len(stars))
	for _, star := range stars {
		if video, ok := videoMap[star.VideoID]; ok {
			videos = append(videos, video)
		}
	}

	logger.Info("获取用户收藏视频列表成功",
		logger.Int64Field("user_id", userID),
		logger.IntField("video_count", len(videos)),
		logger.Int64Field("total_count", total))

	return videos, total, nil
}

// 评论操作
func (s *interactionServiceImpl) CommentAction(ctx context.Context, userID, videoID int64, content string, replyToID int64) (*model.Comment, error) {
	logger.Info("评论操作请求",
		logger.Int64Field("user_id", userID),
		logger.Int64Field("video_id", videoID),
		logger.StringField("content", content),
		logger.Int64Field("reply_to_id", replyToID))

	if content == "" {
		return nil, ErrInvalidCommentContent
	}

	comment := &model.Comment{
		UserID:     userID,
		VideoID:    videoID,
		Content:    content,
		ReplyToID:  replyToID,
		CreateTime: time.Now().Format("2006-01-02 15:04:05"),
	}

	if err := s.commentRepo.Create(ctx, comment); err != nil {
		logger.Error("创建评论记录失败",
			logger.ErrorField(err),
			logger.Int64Field("user_id", userID),
			logger.Int64Field("video_id", videoID))
		return nil, ErrInteractionFailed
	}

	if err := s.statsRepo.IncrementCommentCount(ctx, videoID); err != nil {
		logger.Error("增加视频评论数失败",
			logger.ErrorField(err),
			logger.Int64Field("video_id", videoID))
	}

	if s.kafkaProducer != nil {
		eventData := map[string]interface{}{
			"comment_id":  comment.ID,
			"user_id":     userID,
			"video_id":    videoID,
			"content":     content,
			"reply_to_id": replyToID,
			"created_at":  time.Now(),
		}
		data, _ := json.Marshal(eventData)
		s.kafkaProducer.SendInteractionEvent(ctx, fmt.Sprintf("%d", comment.ID), data)
	}

	logger.Info("评论操作成功",
		logger.Int64Field("user_id", userID),
		logger.Int64Field("video_id", videoID),
		logger.Int64Field("comment_id", comment.ID))

	return comment, nil
}

// 获取评论列表
func (s *interactionServiceImpl) GetCommentList(ctx context.Context, videoID, currentUserID int64, page, pageSize int) ([]*model.Comment, int64, error) {
	logger.Info("获取评论列表请求",
		logger.Int64Field("video_id", videoID),
		logger.Int64Field("current_user_id", currentUserID),
		logger.IntField("page", page),
		logger.IntField("page_size", pageSize))

	comments, total, err := s.commentRepo.ListByVideoID(ctx, videoID, page, pageSize)
	if err != nil {
		logger.Error("获取评论列表失败",
			logger.ErrorField(err),
			logger.Int64Field("video_id", videoID))
		return nil, 0, ErrInternalServer
	}

	logger.Info("获取评论列表成功",
		logger.Int64Field("video_id", videoID),
		logger.IntField("comment_count", len(comments)),
		logger.Int64Field("total_count", total))

	return comments, total, nil
}

// 删除评论
func (s *interactionServiceImpl) DeleteComment(ctx context.Context, userID, videoID, commentID int64) error {
	logger.Info("删除评论请求",
		logger.Int64Field("user_id", userID),
		logger.Int64Field("video_id", videoID),
		logger.Int64Field("comment_id", commentID))

	comment, err := s.commentRepo.FindByID(ctx, commentID)
	if err != nil {
		logger.Error("查询评论失败",
			logger.ErrorField(err),
			logger.Int64Field("comment_id", commentID))
		return ErrInternalServer
	}

	if comment == nil {
		return ErrCommentNotFound
	}

	if comment.UserID != userID {
		return ErrNotCommentOwner
	}

	if err := s.commentRepo.Delete(ctx, commentID, userID, videoID); err != nil {
		logger.Error("删除评论失败",
			logger.ErrorField(err),
			logger.Int64Field("comment_id", commentID),
			logger.Int64Field("user_id", userID),
			logger.Int64Field("video_id", videoID))
		return ErrInteractionFailed
	}

	if s.kafkaProducer != nil {
		eventData := map[string]interface{}{
			"comment_id": commentID,
			"user_id":    userID,
			"video_id":   videoID,
			"deleted_at": time.Now(),
		}
		data, _ := json.Marshal(eventData)
		s.kafkaProducer.SendInteractionEvent(ctx, fmt.Sprintf("%d", commentID), data)
	}

	logger.Info("删除评论成功",
		logger.Int64Field("comment_id", commentID),
		logger.Int64Field("user_id", userID),
		logger.Int64Field("video_id", videoID))

	return nil
}

// 分享操作
func (s *interactionServiceImpl) ShareAction(ctx context.Context, userID, videoID int64) error {
	logger.Info("分享操作请求",
		logger.Int64Field("user_id", userID),
		logger.Int64Field("video_id", videoID))

	share := &model.Share{
		UserID:  userID,
		VideoID: videoID,
	}

	if err := s.shareRepo.Create(ctx, share); err != nil {
		logger.Error("创建分享记录失败",
			logger.ErrorField(err),
			logger.Int64Field("user_id", userID),
			logger.Int64Field("video_id", videoID))
		return ErrInteractionFailed
	}

	if err := s.statsRepo.IncrementShareCount(ctx, videoID); err != nil {
		logger.Error("增加视频分享数失败",
			logger.ErrorField(err),
			logger.Int64Field("video_id", videoID))
	}

	if s.kafkaProducer != nil {
		eventData := map[string]interface{}{
			"share_id":  share.ID,
			"user_id":   userID,
			"video_id":  videoID,
			"shared_at": time.Now(),
		}
		data, _ := json.Marshal(eventData)
		s.kafkaProducer.SendInteractionEvent(ctx, fmt.Sprintf("%d", share.ID), data)
	}

	logger.Info("分享操作成功",
		logger.Int64Field("user_id", userID),
		logger.Int64Field("video_id", videoID),
		logger.Int64Field("share_id", share.ID))

	return nil
}

// 获取互动统计
func (s *interactionServiceImpl) GetCount(ctx context.Context, videoID int64) (int64, int64, int64, int64, error) {
	logger.Info("获取互动统计请求",
		logger.Int64Field("video_id", videoID))

	stats, err := s.statsRepo.FindByVideoID(ctx, videoID)
	if err != nil {
		logger.Error("查询互动统计失败",
			logger.ErrorField(err),
			logger.Int64Field("video_id", videoID))
		return 0, 0, 0, 0, ErrInternalServer
	}

	if stats == nil {
		return 0, 0, 0, 0, nil
	}

	logger.Info("获取互动统计成功",
		logger.Int64Field("video_id", videoID),
		logger.Int64Field("like_count", stats.LikeCount),
		logger.Int64Field("comment_count", stats.CommentCount),
		logger.Int64Field("star_count", stats.StarCount),
		logger.Int64Field("share_count", stats.ShareCount))

	return stats.LikeCount, stats.CommentCount, stats.StarCount, stats.ShareCount, nil
}

// 检查点赞状态
func (s *interactionServiceImpl) CheckLikeStatus(ctx context.Context, userID, videoID int64) (bool, error) {
	logger.Info("检查点赞状态请求",
		logger.Int64Field("user_id", userID),
		logger.Int64Field("video_id", videoID))

	exists, err := s.likeRepo.Exists(ctx, userID, videoID)
	if err != nil {
		logger.Error("检查点赞状态失败",
			logger.ErrorField(err),
			logger.Int64Field("user_id", userID),
			logger.Int64Field("video_id", videoID))
		return false, ErrInternalServer
	}

	logger.Info("检查点赞状态成功",
		logger.Int64Field("user_id", userID),
		logger.Int64Field("video_id", videoID),
		logger.BoolField("is_liked", exists))

	return exists, nil
}

// 检查收藏状态
func (s *interactionServiceImpl) CheckStarStatus(ctx context.Context, userID, videoID int64) (bool, error) {
	logger.Info("检查收藏状态请求",
		logger.Int64Field("user_id", userID),
		logger.Int64Field("video_id", videoID))

	exists, err := s.starRepo.Exists(ctx, userID, videoID)
	if err != nil {
		logger.Error("检查收藏状态失败",
			logger.ErrorField(err),
			logger.Int64Field("user_id", userID),
			logger.Int64Field("video_id", videoID))
		return false, ErrInternalServer
	}

	logger.Info("检查收藏状态成功",
		logger.Int64Field("user_id", userID),
		logger.Int64Field("video_id", videoID),
		logger.BoolField("is_starred", exists))

	return exists, nil
}

// 事务支持
func (s *interactionServiceImpl) WithTransaction(ctx context.Context, fn func(txService InteractionService) error) error {
	return s.likeRepo.WithTransaction(ctx, func(txLikeRepo dao.LikeRepository) error {
		var txStarRepo dao.StarRepository
		var txCommentRepo dao.CommentRepository
		var txShareRepo dao.ShareRepository
		var txStatsRepo dao.VideoInteractionStatsRepository

		err := s.starRepo.WithTransaction(ctx, func(repo dao.StarRepository) error {
			txStarRepo = repo
			return nil
		})
		if err != nil {
			return err
		}

		err = s.commentRepo.WithTransaction(ctx, func(repo dao.CommentRepository) error {
			txCommentRepo = repo
			return nil
		})
		if err != nil {
			return err
		}

		err = s.shareRepo.WithTransaction(ctx, func(repo dao.ShareRepository) error {
			txShareRepo = repo
			return nil
		})
		if err != nil {
			return err
		}

		err = s.statsRepo.WithTransaction(ctx, func(repo dao.VideoInteractionStatsRepository) error {
			txStatsRepo = repo
			return nil
		})
		if err != nil {
			return err
		}

		txService := &interactionServiceImpl{
			likeRepo:      txLikeRepo,
			starRepo:      txStarRepo,
			commentRepo:   txCommentRepo,
			shareRepo:     txShareRepo,
			statsRepo:     txStatsRepo,
			videoService:  s.videoService,
			kafkaProducer: s.kafkaProducer,
		}

		return fn(txService)
	})
}
