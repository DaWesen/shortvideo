package dao

import (
	"context"
	"errors"
	"fmt"
	"shortvideo/internal/message/model"

	"gorm.io/gorm"
)

type MessageRepository interface {
	Create(ctx context.Context, message *model.Message) error
	FindByID(ctx context.Context, id int64) (*model.Message, error)
	Delete(ctx context.Context, messageID, userID int64) error
	GetChatHistory(ctx context.Context, userID1, userID2 int64, lastMessageID int64, pageSize int) ([]*model.Message, int64, error)
	GetLatestMessages(ctx context.Context, userID int64, limit int) ([]*model.Message, error)
	MarkMessageRead(ctx context.Context, userID, messageID int64) error
	MarkMessagesRead(ctx context.Context, userID, sendID int64) error
	GetUnreadCount(ctx context.Context, userID int64) (int64, error)
	GetUnreadCountBySender(ctx context.Context, userID, sendID int64) (int64, error)
	WithTransaction(ctx context.Context, fn func(txRepo MessageRepository) error) error
}

type NotificationRepository interface {
	Create(ctx context.Context, notification *model.SystemNotification) error
	FindByID(ctx context.Context, id int64) (*model.SystemNotification, error)
	GetNotifications(ctx context.Context, userID int64, page, pageSize int, notificationType *int32) ([]*model.SystemNotification, int64, error)
	MarkNotificationRead(ctx context.Context, userID, notificationID int64) error
	MarkAllNotificationsRead(ctx context.Context, userID int64) error
	GetUnreadNotificationCount(ctx context.Context, userID int64) (int64, error)
	Delete(ctx context.Context, notificationID, userID int64) error
	WithTransaction(ctx context.Context, fn func(txRepo NotificationRepository) error) error
}

type messageRepositoryImpl struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) MessageRepository {
	return &messageRepositoryImpl{db: db}
}

func (r *messageRepositoryImpl) Create(ctx context.Context, message *model.Message) error {
	return r.db.WithContext(ctx).Create(message).Error
}

func (r *messageRepositoryImpl) FindByID(ctx context.Context, id int64) (*model.Message, error) {
	var message model.Message
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&message).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &message, err
}

func (r *messageRepositoryImpl) Delete(ctx context.Context, messageID, userID int64) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND (send_id = ? OR receive_id = ?)", messageID, userID, userID).
		Delete(&model.Message{}).Error
}

func (r *messageRepositoryImpl) GetChatHistory(ctx context.Context, userID1, userID2 int64, lastMessageID int64, pageSize int) ([]*model.Message, int64, error) {
	var messages []*model.Message
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Message{}).
		Where("(send_id = ? AND receive_id = ?) OR (send_id = ? AND receive_id = ?)",
			userID1, userID2, userID2, userID1)

	if lastMessageID > 0 {
		query = query.Where("id < ?", lastMessageID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("id DESC").Limit(pageSize).Find(&messages).Error

	return messages, total, err
}

func (r *messageRepositoryImpl) GetLatestMessages(ctx context.Context, userID int64, limit int) ([]*model.Message, error) {
	var messages []*model.Message

	subQuery := r.db.WithContext(ctx).Model(&model.Message{}).
		Select("MAX(id) as max_id").
		Where("receive_id = ? OR send_id = ?", userID, userID).
		Group(fmt.Sprintf("CASE WHEN send_id = %d THEN receive_id ELSE send_id END", userID))

	err := r.db.WithContext(ctx).Where("id IN (?)", subQuery).
		Order("create_time DESC").
		Limit(limit).
		Find(&messages).Error

	return messages, err
}

func (r *messageRepositoryImpl) MarkMessageRead(ctx context.Context, userID, messageID int64) error {
	return r.db.WithContext(ctx).Model(&model.Message{}).
		Where("id = ? AND receive_id = ?", messageID, userID).
		Update("is_read", true).Error
}

func (r *messageRepositoryImpl) MarkMessagesRead(ctx context.Context, userID, sendID int64) error {
	return r.db.WithContext(ctx).Model(&model.Message{}).
		Where("receive_id = ? AND send_id = ? AND is_read = ?", userID, sendID, false).
		Update("is_read", true).Error
}

func (r *messageRepositoryImpl) GetUnreadCount(ctx context.Context, userID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Message{}).
		Where("receive_id = ? AND is_read = ?", userID, false).
		Count(&count).Error
	return count, err
}

func (r *messageRepositoryImpl) GetUnreadCountBySender(ctx context.Context, userID, sendID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Message{}).
		Where("receive_id = ? AND send_id = ? AND is_read = ?", userID, sendID, false).
		Count(&count).Error
	return count, err
}

func (r *messageRepositoryImpl) WithTransaction(ctx context.Context, fn func(txRepo MessageRepository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := &messageRepositoryImpl{db: tx}
		return fn(txRepo)
	})
}

type notificationRepositoryImpl struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepositoryImpl{db: db}
}

func (r *notificationRepositoryImpl) Create(ctx context.Context, notification *model.SystemNotification) error {
	return r.db.WithContext(ctx).Create(notification).Error
}

func (r *notificationRepositoryImpl) FindByID(ctx context.Context, id int64) (*model.SystemNotification, error) {
	var notification model.SystemNotification
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&notification).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &notification, err
}

func (r *notificationRepositoryImpl) GetNotifications(ctx context.Context, userID int64, page, pageSize int, notificationType *int32) ([]*model.SystemNotification, int64, error) {
	var notifications []*model.SystemNotification
	var total int64
	offset := (page - 1) * pageSize

	query := r.db.WithContext(ctx).Model(&model.SystemNotification{}).
		Where("user_id = ?", userID)

	if notificationType != nil {
		query = query.Where("type = ?", *notificationType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Offset(offset).Limit(pageSize).
		Order("create_time DESC").
		Find(&notifications).Error

	return notifications, total, err
}

func (r *notificationRepositoryImpl) MarkNotificationRead(ctx context.Context, userID, notificationID int64) error {
	return r.db.WithContext(ctx).Model(&model.SystemNotification{}).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Update("is_read", true).Error
}

func (r *notificationRepositoryImpl) MarkAllNotificationsRead(ctx context.Context, userID int64) error {
	return r.db.WithContext(ctx).Model(&model.SystemNotification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Update("is_read", true).Error
}

func (r *notificationRepositoryImpl) GetUnreadNotificationCount(ctx context.Context, userID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.SystemNotification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count).Error
	return count, err
}

func (r *notificationRepositoryImpl) Delete(ctx context.Context, notificationID, userID int64) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Delete(&model.SystemNotification{}).Error
}

func (r *notificationRepositoryImpl) WithTransaction(ctx context.Context, fn func(txRepo NotificationRepository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := &notificationRepositoryImpl{db: tx}
		return fn(txRepo)
	})
}
