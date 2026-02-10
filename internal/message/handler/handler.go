package handler

import (
	"context"
	"shortvideo/internal/message/service"
	"shortvideo/kitex_gen/common"
	message "shortvideo/kitex_gen/message"
)

// MessageServiceImpl implements the last service interface defined in the IDL.
type MessageServiceImpl struct {
	messageService service.MessageService
}

func NewMessageService(messageService service.MessageService) *MessageServiceImpl {
	return &MessageServiceImpl{
		messageService: messageService,
	}
}

// SendMessage implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) SendMessage(ctx context.Context, req *message.SendMessageReq) (resp *message.SendMessageResp, err error) {
	successMsg := "成功"
	resp = &message.SendMessageResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
		MessageId: 0,
	}

	messageID, err := s.messageService.SendMessage(ctx, req.SenderId, req.ReceiverId, req.Content)
	if err != nil {
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	resp.MessageId = messageID
	return resp, nil
}

// GetChatHistory implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) GetChatHistory(ctx context.Context, req *message.GetChatHistoryReq) (resp *message.GetChatHistoryResp, err error) {
	successMsg := "成功"
	resp = &message.GetChatHistoryResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
		Messages:      []*common.Message{},
		NextMessageId: 0,
	}

	messages, nextMessageID, err := s.messageService.GetChatHistory(ctx, req.UserId1, req.UserId2, req.LastMessageId, int(req.PageSize))
	if err != nil {
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	commonMessages := make([]*common.Message, len(messages))
	for i, msg := range messages {
		commonMessages[i] = &common.Message{
			Id:         msg.ID,
			SendId:     msg.SendID,
			ReceiveId:  msg.ReceiveID,
			Content:    msg.Content,
			CreateTime: msg.CreateTime,
			IsRead:     msg.IsRead,
		}
	}

	resp.Messages = commonMessages
	resp.NextMessageId = nextMessageID
	return resp, nil
}

// GetLatestMessages implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) GetLatestMessages(ctx context.Context, req *message.GetLatestMessagesReq) (resp *message.GetLatestMessagesResp, err error) {
	successMsg := "成功"
	resp = &message.GetLatestMessagesResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
		Messages: []*message.LatestMessage{},
	}

	messages, err := s.messageService.GetLatestMessages(ctx, req.UserId, int(req.Limit))
	if err != nil {
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	latestMessages := make([]*message.LatestMessage, len(messages))
	for i, msg := range messages {
		otherUserID := msg.SendID
		if msg.SendID == req.UserId {
			otherUserID = msg.ReceiveID
		}

		user := &common.User{
			Id: otherUserID,
		}

		latestMessages[i] = &message.LatestMessage{
			User: user,
			LastMessage: &common.Message{
				Id:         msg.ID,
				SendId:     msg.SendID,
				ReceiveId:  msg.ReceiveID,
				Content:    msg.Content,
				CreateTime: msg.CreateTime,
				IsRead:     msg.IsRead,
			},
			UnreadCount: 0,
		}
	}

	resp.Messages = latestMessages
	return resp, nil
}

// MarkMessageRead implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) MarkMessageRead(ctx context.Context, req *message.MarkMessageReadReq) (resp *message.MarkMessageReadResp, err error) {
	successMsg := "成功"
	resp = &message.MarkMessageReadResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
	}

	err = s.messageService.MarkMessageRead(ctx, req.UserId, req.MessageId)
	if err != nil {
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	return resp, nil
}

// DeleteMessage implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) DeleteMessage(ctx context.Context, req *message.DeleteMessageReq) (resp *message.DeleteMessageResp, err error) {
	successMsg := "成功"
	resp = &message.DeleteMessageResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
	}

	err = s.messageService.DeleteMessage(ctx, req.UserId, req.MessageId)
	if err != nil {
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	return resp, nil
}

// GetUnreadCount implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) GetUnreadCount(ctx context.Context, req *message.GetUnreadCountReq) (resp *message.GetUnreadCountResp, err error) {
	successMsg := "成功"
	resp = &message.GetUnreadCountResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
		TotalUnread: 0,
	}

	count, err := s.messageService.GetUnreadCount(ctx, req.UserId)
	if err != nil {
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	resp.TotalUnread = count
	return resp, nil
}

// GetNotifications implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) GetNotifications(ctx context.Context, req *message.GetNotificationsReq) (resp *message.GetNotificationsResp, err error) {
	successMsg := "成功"
	resp = &message.GetNotificationsResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
		Notifications: []*message.SystemNotification{},
		TotalCount:    0,
	}

	notifications, total, err := s.messageService.GetNotifications(ctx, req.UserId, int(req.Page), int(req.PageSize), req.Type)
	if err != nil {
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	systemNotifications := make([]*message.SystemNotification, len(notifications))
	for i, notification := range notifications {
		systemNotifications[i] = &message.SystemNotification{
			Id:         notification.ID,
			UserId:     notification.UserID,
			Title:      notification.Title,
			Content:    notification.Content,
			Type:       notification.Type,
			RelatedId:  notification.RelatedID,
			IsRead:     notification.IsRead,
			CreateTime: notification.CreateTime,
		}
	}

	resp.Notifications = systemNotifications
	resp.TotalCount = int32(total)
	return resp, nil
}

// MarkNotificationRead implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) MarkNotificationRead(ctx context.Context, req *message.MarkNotificationReadReq) (resp *message.MarkNotificationReadResp, err error) {
	successMsg := "成功"
	resp = &message.MarkNotificationReadResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
	}

	err = s.messageService.MarkNotificationRead(ctx, req.UserId, req.NotificationId)
	if err != nil {
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	return resp, nil
}
