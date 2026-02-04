package handler

import (
	"context"
	message "shortvideo/kitex_gen/message"
)

// MessageServiceImpl implements the last service interface defined in the IDL.
type MessageServiceImpl struct{}

func NewMessageService() *MessageServiceImpl {
	return &MessageServiceImpl{}
}

// SendMessage implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) SendMessage(ctx context.Context, req *message.SendMessageReq) (resp *message.SendMessageResp, err error) {
	// TODO: Your code here...
	return
}

// GetChatHistory implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) GetChatHistory(ctx context.Context, req *message.GetChatHistoryReq) (resp *message.GetChatHistoryResp, err error) {
	// TODO: Your code here...
	return
}

// GetLatestMessages implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) GetLatestMessages(ctx context.Context, req *message.GetLatestMessagesReq) (resp *message.GetLatestMessagesResp, err error) {
	// TODO: Your code here...
	return
}

// MarkMessageRead implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) MarkMessageRead(ctx context.Context, req *message.MarkMessageReadReq) (resp *message.MarkMessageReadResp, err error) {
	// TODO: Your code here...
	return
}

// DeleteMessage implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) DeleteMessage(ctx context.Context, req *message.DeleteMessageReq) (resp *message.DeleteMessageResp, err error) {
	// TODO: Your code here...
	return
}

// GetUnreadCount implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) GetUnreadCount(ctx context.Context, req *message.GetUnreadCountReq) (resp *message.GetUnreadCountResp, err error) {
	// TODO: Your code here...
	return
}

// GetNotifications implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) GetNotifications(ctx context.Context, req *message.GetNotificationsReq) (resp *message.GetNotificationsResp, err error) {
	// TODO: Your code here...
	return
}

// MarkNotificationRead implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) MarkNotificationRead(ctx context.Context, req *message.MarkNotificationReadReq) (resp *message.MarkNotificationReadResp, err error) {
	// TODO: Your code here...
	return
}
