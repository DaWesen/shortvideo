package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"shortvideo/internal/message/dao"
	"shortvideo/internal/message/model"
	userService "shortvideo/internal/user/service"
	"shortvideo/pkg/logger"
	"shortvideo/pkg/mq"
	"time"
)

var (
	ErrMessageNotFound       = errors.New("消息不存在")
	ErrNotMessageOwner       = errors.New("不是消息所有者")
	ErrInvalidMessageContent = errors.New("无效的消息内容")
	ErrMessageSendFailed     = errors.New("消息发送失败")
	ErrNotificationNotFound  = errors.New("通知不存在")
	ErrNotNotificationOwner  = errors.New("不是通知所有者")
	ErrInternalServer        = errors.New("服务器内部错误")
	ErrUserNotFound          = errors.New("用户不存在")
)

type MessageService interface {
	//发送消息
	SendMessage(ctx context.Context, senderID, receiverID int64, content string) (int64, error)
	//获取聊天历史
	GetChatHistory(ctx context.Context, userID1, userID2, lastMessageID int64, pageSize int) ([]*model.Message, int64, error)
	//获取最新消息
	GetLatestMessages(ctx context.Context, userID int64, limit int) ([]*model.Message, error)
	//标记消息已读
	MarkMessageRead(ctx context.Context, userID, messageID int64) error
	//删除消息
	DeleteMessage(ctx context.Context, userID, messageID int64) error
	//获取未读消息数
	GetUnreadCount(ctx context.Context, userID int64) (int64, error)
	//获取通知列表
	GetNotifications(ctx context.Context, userID int64, page, pageSize int, notificationType *int32) ([]*model.SystemNotification, int64, error)
	//标记通知已读
	MarkNotificationRead(ctx context.Context, userID, notificationID int64) error
	//创建系统通知
	CreateNotification(ctx context.Context, userID int64, title, content string, notificationType int32, relatedID int64) (int64, error)
	//事务支持
	WithTransaction(ctx context.Context, fn func(txService MessageService) error) error
}

type messageServiceImpl struct {
	messageRepo      dao.MessageRepository
	notificationRepo dao.NotificationRepository
	userService      userService.UserService
	kafkaProducer    *mq.Producer
}

func NewMessageService(
	messageRepo dao.MessageRepository,
	notificationRepo dao.NotificationRepository,
	userService userService.UserService,
	kafkaProducer *mq.Producer,
) MessageService {
	return &messageServiceImpl{
		messageRepo:      messageRepo,
		notificationRepo: notificationRepo,
		userService:      userService,
		kafkaProducer:    kafkaProducer,
	}
}

// 发送消息
func (s *messageServiceImpl) SendMessage(ctx context.Context, senderID, receiverID int64, content string) (int64, error) {
	logger.Info("发送消息请求",
		logger.Int64Field("sender_id", senderID),
		logger.Int64Field("receiver_id", receiverID),
		logger.StringField("content", content))

	if content == "" {
		return 0, ErrInvalidMessageContent
	}

	var err error
	_, err = s.userService.GetUserByID(ctx, senderID)
	if err != nil {
		logger.Error("获取发送者信息失败",
			logger.ErrorField(err),
			logger.Int64Field("sender_id", senderID))
		return 0, ErrUserNotFound
	}

	_, err = s.userService.GetUserByID(ctx, receiverID)
	if err != nil {
		logger.Error("获取接收者信息失败",
			logger.ErrorField(err),
			logger.Int64Field("receiver_id", receiverID))
		return 0, ErrUserNotFound
	}

	message := &model.Message{
		SendID:     senderID,
		ReceiveID:  receiverID,
		Content:    content,
		CreateTime: time.Now().Format("2006-01-02 15:04:05"),
		IsRead:     false,
	}

	if err := s.messageRepo.Create(ctx, message); err != nil {
		logger.Error("创建消息失败",
			logger.ErrorField(err),
			logger.Int64Field("sender_id", senderID),
			logger.Int64Field("receiver_id", receiverID))
		return 0, ErrMessageSendFailed
	}

	if s.kafkaProducer != nil {
		eventData := map[string]interface{}{
			"message_id":  message.ID,
			"sender_id":   senderID,
			"receiver_id": receiverID,
			"content":     content,
			"created_at":  time.Now(),
		}
		data, _ := json.Marshal(eventData)
		s.kafkaProducer.SendMessageEvent(ctx, fmt.Sprintf("%d", message.ID), data)
	}

	logger.Info("发送消息成功",
		logger.Int64Field("message_id", message.ID),
		logger.Int64Field("sender_id", senderID),
		logger.Int64Field("receiver_id", receiverID))

	return message.ID, nil
}

// 获取聊天历史
func (s *messageServiceImpl) GetChatHistory(ctx context.Context, userID1, userID2, lastMessageID int64, pageSize int) ([]*model.Message, int64, error) {
	logger.Info("获取聊天历史请求",
		logger.Int64Field("user_id1", userID1),
		logger.Int64Field("user_id2", userID2),
		logger.Int64Field("last_message_id", lastMessageID),
		logger.IntField("page_size", pageSize))

	messages, total, err := s.messageRepo.GetChatHistory(ctx, userID1, userID2, lastMessageID, pageSize)
	if err != nil {
		logger.Error("获取聊天历史失败",
			logger.ErrorField(err),
			logger.Int64Field("user_id1", userID1),
			logger.Int64Field("user_id2", userID2))
		return nil, 0, ErrInternalServer
	}

	if len(messages) > 0 {
		go func() {
			if err := s.messageRepo.MarkMessagesRead(context.Background(), userID1, userID2); err != nil {
				logger.Error("标记消息已读失败",
					logger.ErrorField(err),
					logger.Int64Field("user_id1", userID1),
					logger.Int64Field("user_id2", userID2))
			}
		}()
	}

	var nextMessageID int64
	if len(messages) > 0 {
		nextMessageID = messages[len(messages)-1].ID
	}

	logger.Info("获取聊天历史成功",
		logger.Int64Field("user_id1", userID1),
		logger.Int64Field("user_id2", userID2),
		logger.IntField("message_count", len(messages)),
		logger.Int64Field("total_count", total))

	return messages, nextMessageID, nil
}

// 获取最新消息
func (s *messageServiceImpl) GetLatestMessages(ctx context.Context, userID int64, limit int) ([]*model.Message, error) {
	logger.Info("获取最新消息请求",
		logger.Int64Field("user_id", userID),
		logger.IntField("limit", limit))

	messages, err := s.messageRepo.GetLatestMessages(ctx, userID, limit)
	if err != nil {
		logger.Error("获取最新消息失败",
			logger.ErrorField(err),
			logger.Int64Field("user_id", userID))
		return nil, ErrInternalServer
	}

	logger.Info("获取最新消息成功",
		logger.Int64Field("user_id", userID),
		logger.IntField("message_count", len(messages)))

	return messages, nil
}

// 标记消息已读
func (s *messageServiceImpl) MarkMessageRead(ctx context.Context, userID, messageID int64) error {
	logger.Info("标记消息已读请求",
		logger.Int64Field("user_id", userID),
		logger.Int64Field("message_id", messageID))

	message, err := s.messageRepo.FindByID(ctx, messageID)
	if err != nil {
		logger.Error("查询消息失败",
			logger.ErrorField(err),
			logger.Int64Field("message_id", messageID))
		return ErrInternalServer
	}

	if message == nil {
		return ErrMessageNotFound
	}

	if message.ReceiveID != userID {
		return ErrNotMessageOwner
	}

	if err := s.messageRepo.MarkMessageRead(ctx, userID, messageID); err != nil {
		logger.Error("标记消息已读失败",
			logger.ErrorField(err),
			logger.Int64Field("user_id", userID),
			logger.Int64Field("message_id", messageID))
		return ErrInternalServer
	}

	logger.Info("标记消息已读成功",
		logger.Int64Field("message_id", messageID),
		logger.Int64Field("user_id", userID))

	return nil
}

// 删除消息
func (s *messageServiceImpl) DeleteMessage(ctx context.Context, userID, messageID int64) error {
	logger.Info("删除消息请求",
		logger.Int64Field("user_id", userID),
		logger.Int64Field("message_id", messageID))

	message, err := s.messageRepo.FindByID(ctx, messageID)
	if err != nil {
		logger.Error("查询消息失败",
			logger.ErrorField(err),
			logger.Int64Field("message_id", messageID))
		return ErrInternalServer
	}

	if message == nil {
		return ErrMessageNotFound
	}

	if message.SendID != userID && message.ReceiveID != userID {
		return ErrNotMessageOwner
	}

	if err := s.messageRepo.Delete(ctx, messageID, userID); err != nil {
		logger.Error("删除消息失败",
			logger.ErrorField(err),
			logger.Int64Field("user_id", userID),
			logger.Int64Field("message_id", messageID))
		return ErrInternalServer
	}

	logger.Info("删除消息成功",
		logger.Int64Field("message_id", messageID),
		logger.Int64Field("user_id", userID))

	return nil
}

// 获取未读消息数
func (s *messageServiceImpl) GetUnreadCount(ctx context.Context, userID int64) (int64, error) {
	logger.Info("获取未读消息数请求",
		logger.Int64Field("user_id", userID))

	count, err := s.messageRepo.GetUnreadCount(ctx, userID)
	if err != nil {
		logger.Error("获取未读消息数失败",
			logger.ErrorField(err),
			logger.Int64Field("user_id", userID))
		return 0, ErrInternalServer
	}

	logger.Info("获取未读消息数成功",
		logger.Int64Field("user_id", userID),
		logger.Int64Field("unread_count", count))

	return count, nil
}

// 获取通知列表
func (s *messageServiceImpl) GetNotifications(ctx context.Context, userID int64, page, pageSize int, notificationType *int32) ([]*model.SystemNotification, int64, error) {
	logger.Info("获取通知列表请求",
		logger.Int64Field("user_id", userID),
		logger.IntField("page", page),
		logger.IntField("page_size", pageSize))

	notifications, total, err := s.notificationRepo.GetNotifications(ctx, userID, page, pageSize, notificationType)
	if err != nil {
		logger.Error("获取通知列表失败",
			logger.ErrorField(err),
			logger.Int64Field("user_id", userID))
		return nil, 0, ErrInternalServer
	}

	logger.Info("获取通知列表成功",
		logger.Int64Field("user_id", userID),
		logger.IntField("notification_count", len(notifications)),
		logger.Int64Field("total_count", total))

	return notifications, total, nil
}

// 标记通知已读
func (s *messageServiceImpl) MarkNotificationRead(ctx context.Context, userID, notificationID int64) error {
	logger.Info("标记通知已读请求",
		logger.Int64Field("user_id", userID),
		logger.Int64Field("notification_id", notificationID))

	notification, err := s.notificationRepo.FindByID(ctx, notificationID)
	if err != nil {
		logger.Error("查询通知失败",
			logger.ErrorField(err),
			logger.Int64Field("notification_id", notificationID))
		return ErrInternalServer
	}

	if notification == nil {
		return ErrNotificationNotFound
	}

	if notification.UserID != userID {
		return ErrNotNotificationOwner
	}

	if err := s.notificationRepo.MarkNotificationRead(ctx, userID, notificationID); err != nil {
		logger.Error("标记通知已读失败",
			logger.ErrorField(err),
			logger.Int64Field("user_id", userID),
			logger.Int64Field("notification_id", notificationID))
		return ErrInternalServer
	}

	logger.Info("标记通知已读成功",
		logger.Int64Field("notification_id", notificationID),
		logger.Int64Field("user_id", userID))

	return nil
}

// 创建系统通知
func (s *messageServiceImpl) CreateNotification(ctx context.Context, userID int64, title, content string, notificationType int32, relatedID int64) (int64, error) {
	logger.Info("创建系统通知请求",
		logger.Int64Field("user_id", userID),
		logger.StringField("title", title),
		logger.IntField("type", int(notificationType)))

	var err error
	_, err = s.userService.GetUserByID(ctx, userID)
	if err != nil {
		return 0, ErrUserNotFound
	}

	notification := &model.SystemNotification{
		UserID:     userID,
		Title:      title,
		Content:    content,
		Type:       notificationType,
		RelatedID:  relatedID,
		IsRead:     false,
		CreateTime: time.Now().Format("2006-01-02 15:04:05"),
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		logger.Error("创建系统通知失败",
			logger.ErrorField(err),
			logger.Int64Field("user_id", userID))
		return 0, ErrInternalServer
	}

	logger.Info("创建系统通知成功",
		logger.Int64Field("notification_id", notification.ID),
		logger.Int64Field("user_id", userID))

	return notification.ID, nil
}

// 事务支持
func (s *messageServiceImpl) WithTransaction(ctx context.Context, fn func(txService MessageService) error) error {
	return s.messageRepo.WithTransaction(ctx, func(txMessageRepo dao.MessageRepository) error {
		var txNotificationRepo dao.NotificationRepository

		err := s.notificationRepo.WithTransaction(ctx, func(repo dao.NotificationRepository) error {
			txNotificationRepo = repo
			return nil
		})
		if err != nil {
			return err
		}

		txService := &messageServiceImpl{
			messageRepo:      txMessageRepo,
			notificationRepo: txNotificationRepo,
			userService:      s.userService,
			kafkaProducer:    s.kafkaProducer,
		}

		return fn(txService)
	})
}
