package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"shortvideo/internal/social/dao"
	"shortvideo/internal/social/model"
	userService "shortvideo/internal/user/service"
	"shortvideo/pkg/logger"
	"shortvideo/pkg/mq"
	"time"
)

var (
	ErrFollowFailed         = errors.New("关注操作失败")
	ErrUnfollowFailed       = errors.New("取消关注操作失败")
	ErrUserNotFound         = errors.New("用户不存在")
	ErrCannotFollowYourself = errors.New("不能关注自己")
	ErrAlreadyFollowing     = errors.New("已经关注过了")
	ErrNotFollowing         = errors.New("没有关注过")
	ErrInternalServer       = errors.New("服务器内部错误")
)

type SocialService interface {
	//关注操作
	FollowAction(ctx context.Context, userID, targetUserID int64, action bool) error
	//获取关注列表
	GetFollowList(ctx context.Context, userID, currentUserID int64, page, pageSize int) ([]int64, int64, error)
	//获取粉丝列表
	GetFollowerList(ctx context.Context, userID, currentUserID int64, page, pageSize int) ([]int64, int64, error)
	//获取好友列表
	GetFriendList(ctx context.Context, userID int64, page, pageSize int) ([]int64, error)
	//检查关注状态
	CheckFollow(ctx context.Context, userID, targetUserID int64) (bool, error)
	//检查互相关注状态
	CheckMutualFollow(ctx context.Context, userID1, userID2 int64) (bool, error)
	//获取关注统计
	GetFollowStats(ctx context.Context, userID int64) (int64, int64, int64, error)
	//批量检查关注状态
	BatchCheckFollow(ctx context.Context, userID int64, targetUserIDs []int64) (map[int64]bool, error)
	//事务支持
	WithTransaction(ctx context.Context, fn func(txService SocialService) error) error
}

type socialServiceImpl struct {
	followRepo    dao.FollowRepository
	userService   userService.UserService
	kafkaProducer *mq.Producer
}

func NewSocialService(
	followRepo dao.FollowRepository,
	userService userService.UserService,
	kafkaProducer *mq.Producer,
) SocialService {
	return &socialServiceImpl{
		followRepo:    followRepo,
		userService:   userService,
		kafkaProducer: kafkaProducer,
	}
}

// 关注操作
func (s *socialServiceImpl) FollowAction(ctx context.Context, userID, targetUserID int64, action bool) error {
	logger.Info("关注操作请求",
		logger.Int64Field("user_id", userID),
		logger.Int64Field("target_user_id", targetUserID),
		logger.BoolField("action", action))

	if userID == targetUserID {
		return ErrCannotFollowYourself
	}

	_, err := s.userService.GetUserByID(ctx, userID)
	if err != nil {
		logger.Error("获取用户信息失败",
			logger.ErrorField(err),
			logger.Int64Field("user_id", userID))
		return ErrUserNotFound
	}

	_, err = s.userService.GetUserByID(ctx, targetUserID)
	if err != nil {
		logger.Error("获取目标用户信息失败",
			logger.ErrorField(err),
			logger.Int64Field("target_user_id", targetUserID))
		return ErrUserNotFound
	}

	exists, err := s.followRepo.Exists(ctx, userID, targetUserID)
	if err != nil {
		logger.Error("检查关注状态失败",
			logger.ErrorField(err),
			logger.Int64Field("user_id", userID),
			logger.Int64Field("target_user_id", targetUserID))
		return ErrInternalServer
	}

	if action {
		if exists {
			return ErrAlreadyFollowing
		}

		follow := &model.Follow{
			UserID:       userID,
			TargetUserID: targetUserID,
		}

		if err := s.followRepo.Create(ctx, follow); err != nil {
			logger.Error("创建关注记录失败",
				logger.ErrorField(err),
				logger.Int64Field("user_id", userID),
				logger.Int64Field("target_user_id", targetUserID))
			return ErrFollowFailed
		}
	} else {
		if !exists {
			return ErrNotFollowing
		}

		if err := s.followRepo.Delete(ctx, userID, targetUserID); err != nil {
			logger.Error("删除关注记录失败",
				logger.ErrorField(err),
				logger.Int64Field("user_id", userID),
				logger.Int64Field("target_user_id", targetUserID))
			return ErrUnfollowFailed
		}
	}

	if s.kafkaProducer != nil {
		eventData := map[string]interface{}{
			"user_id":        userID,
			"target_user_id": targetUserID,
			"action":         action,
			"created_at":     time.Now(),
		}
		data, _ := json.Marshal(eventData)
		s.kafkaProducer.SendSocialEvent(ctx, fmt.Sprintf("%d", userID), data)
	}

	logger.Info("关注操作成功",
		logger.Int64Field("user_id", userID),
		logger.Int64Field("target_user_id", targetUserID),
		logger.BoolField("action", action))

	return nil
}

// 获取关注列表
func (s *socialServiceImpl) GetFollowList(ctx context.Context, userID, currentUserID int64, page, pageSize int) ([]int64, int64, error) {
	logger.Info("获取关注列表请求",
		logger.Int64Field("user_id", userID),
		logger.Int64Field("current_user_id", currentUserID),
		logger.IntField("page", page),
		logger.IntField("page_size", pageSize))

	follows, total, err := s.followRepo.FindFollowing(ctx, userID, page, pageSize)
	if err != nil {
		logger.Error("获取关注列表失败",
			logger.ErrorField(err),
			logger.Int64Field("user_id", userID))
		return nil, 0, ErrInternalServer
	}

	targetUserIDs := make([]int64, len(follows))
	for i, follow := range follows {
		targetUserIDs[i] = follow.TargetUserID
	}

	logger.Info("获取关注列表成功",
		logger.Int64Field("user_id", userID),
		logger.IntField("follow_count", len(targetUserIDs)),
		logger.Int64Field("total_count", total))

	return targetUserIDs, total, nil
}

// 获取粉丝列表
func (s *socialServiceImpl) GetFollowerList(ctx context.Context, userID, currentUserID int64, page, pageSize int) ([]int64, int64, error) {
	logger.Info("获取粉丝列表请求",
		logger.Int64Field("user_id", userID),
		logger.Int64Field("current_user_id", currentUserID),
		logger.IntField("page", page),
		logger.IntField("page_size", pageSize))

	follows, total, err := s.followRepo.FindFollowers(ctx, userID, page, pageSize)
	if err != nil {
		logger.Error("获取粉丝列表失败",
			logger.ErrorField(err),
			logger.Int64Field("user_id", userID))
		return nil, 0, ErrInternalServer
	}

	followerIDs := make([]int64, len(follows))
	for i, follow := range follows {
		followerIDs[i] = follow.UserID
	}

	logger.Info("获取粉丝列表成功",
		logger.Int64Field("user_id", userID),
		logger.IntField("follower_count", len(followerIDs)),
		logger.Int64Field("total_count", total))

	return followerIDs, total, nil
}

// 获取好友列表
func (s *socialServiceImpl) GetFriendList(ctx context.Context, userID int64, page, pageSize int) ([]int64, error) {
	logger.Info("获取好友列表请求",
		logger.Int64Field("user_id", userID),
		logger.IntField("page", page),
		logger.IntField("page_size", pageSize))

	follows, err := s.followRepo.FindFriends(ctx, userID, page, pageSize)
	if err != nil {
		logger.Error("获取好友列表失败",
			logger.ErrorField(err),
			logger.Int64Field("user_id", userID))
		return nil, ErrInternalServer
	}

	friendIDs := make([]int64, len(follows))
	for i, follow := range follows {
		friendIDs[i] = follow.TargetUserID
	}

	logger.Info("获取好友列表成功",
		logger.Int64Field("user_id", userID),
		logger.IntField("friend_count", len(friendIDs)))

	return friendIDs, nil
}

// 检查关注状态
func (s *socialServiceImpl) CheckFollow(ctx context.Context, userID, targetUserID int64) (bool, error) {
	logger.Info("检查关注状态请求",
		logger.Int64Field("user_id", userID),
		logger.Int64Field("target_user_id", targetUserID))

	exists, err := s.followRepo.Exists(ctx, userID, targetUserID)
	if err != nil {
		logger.Error("检查关注状态失败",
			logger.ErrorField(err),
			logger.Int64Field("user_id", userID),
			logger.Int64Field("target_user_id", targetUserID))
		return false, ErrInternalServer
	}

	logger.Info("检查关注状态成功",
		logger.Int64Field("user_id", userID),
		logger.Int64Field("target_user_id", targetUserID),
		logger.BoolField("is_following", exists))

	return exists, nil
}

// 检查互相关注状态
func (s *socialServiceImpl) CheckMutualFollow(ctx context.Context, userID1, userID2 int64) (bool, error) {
	logger.Info("检查互相关注状态请求",
		logger.Int64Field("user_id1", userID1),
		logger.Int64Field("user_id2", userID2))

	following1, err := s.followRepo.Exists(ctx, userID1, userID2)
	if err != nil {
		logger.Error("检查关注状态失败",
			logger.ErrorField(err),
			logger.Int64Field("user_id1", userID1),
			logger.Int64Field("user_id2", userID2))
		return false, ErrInternalServer
	}

	following2, err := s.followRepo.Exists(ctx, userID2, userID1)
	if err != nil {
		logger.Error("检查关注状态失败",
			logger.ErrorField(err),
			logger.Int64Field("user_id2", userID2),
			logger.Int64Field("user_id1", userID1))
		return false, ErrInternalServer
	}

	isMutual := following1 && following2

	logger.Info("检查互相关注状态成功",
		logger.Int64Field("user_id1", userID1),
		logger.Int64Field("user_id2", userID2),
		logger.BoolField("is_mutual_follow", isMutual))

	return isMutual, nil
}

// 获取关注统计
func (s *socialServiceImpl) GetFollowStats(ctx context.Context, userID int64) (int64, int64, int64, error) {
	logger.Info("获取关注统计请求",
		logger.Int64Field("user_id", userID))

	followCount, err := s.followRepo.CountFollowing(ctx, userID)
	if err != nil {
		logger.Error("获取关注数失败",
			logger.ErrorField(err),
			logger.Int64Field("user_id", userID))
		return 0, 0, 0, ErrInternalServer
	}

	followerCount, err := s.followRepo.CountFollowers(ctx, userID)
	if err != nil {
		logger.Error("获取粉丝数失败",
			logger.ErrorField(err),
			logger.Int64Field("user_id", userID))
		return 0, 0, 0, ErrInternalServer
	}

	friendCount, err := s.followRepo.CountFriends(ctx, userID)
	if err != nil {
		logger.Error("获取好友数失败",
			logger.ErrorField(err),
			logger.Int64Field("user_id", userID))
		return 0, 0, 0, ErrInternalServer
	}

	logger.Info("获取关注统计成功",
		logger.Int64Field("user_id", userID),
		logger.Int64Field("follow_count", followCount),
		logger.Int64Field("follower_count", followerCount),
		logger.Int64Field("friend_count", friendCount))

	return followCount, followerCount, friendCount, nil
}

// 批量检查关注状态
func (s *socialServiceImpl) BatchCheckFollow(ctx context.Context, userID int64, targetUserIDs []int64) (map[int64]bool, error) {
	logger.Info("批量检查关注状态请求",
		logger.Int64Field("user_id", userID),
		logger.IntField("target_user_count", len(targetUserIDs)))

	result, err := s.followRepo.BatchCheckFollow(ctx, userID, targetUserIDs)
	if err != nil {
		logger.Error("批量检查关注状态失败",
			logger.ErrorField(err),
			logger.Int64Field("user_id", userID))
		return nil, ErrInternalServer
	}

	logger.Info("批量检查关注状态成功",
		logger.Int64Field("user_id", userID),
		logger.IntField("result_count", len(result)))

	return result, nil
}

// 事务支持
func (s *socialServiceImpl) WithTransaction(ctx context.Context, fn func(txService SocialService) error) error {
	return s.followRepo.WithTransaction(ctx, func(txFollowRepo dao.FollowRepository) error {
		txService := &socialServiceImpl{
			followRepo:    txFollowRepo,
			userService:   s.userService,
			kafkaProducer: s.kafkaProducer,
		}

		return fn(txService)
	})
}
