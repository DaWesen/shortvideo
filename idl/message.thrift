namespace go message

include "common.thrift"

struct SendMessageReq{
    1:i64 senderId
    2:i64 receiverId
    3:string content
}

struct SendMessageResp{
    1:common.BaseResp BaseResp
    2:i64 messageId
}

struct GetChatHistoryReq{
    1:i64 userId1
    2:i64 userId2
    3:i64 lastMessageId
    4:i32 pageSize
}

struct GetChatHistoryResp{
    1:common.BaseResp BaseResp
    2:list<common.Message> messages
    3:i64 nextMessageId
}

struct GetLatestMessagesReq{
    1:i64 userId
    2:i32 limit
}

struct LatestMessage{
    1:common.User user
    2:common.Message lastMessage
    3:i32 unreadCount
}

struct GetLatestMessagesResp{
    1:common.BaseResp BaseResp
    2:list<LatestMessage> messages
}

struct MarkMessageReadReq{
    1:i64 userId
    2:i64 messageId
}

struct MarkMessageReadResp{
    1:common.BaseResp BaseResp
}

struct DeleteMessageReq{
    1:i64 userId
    2:i64 messageId
}

struct DeleteMessageResp{
    1:common.BaseResp BaseResp
}

struct GetUnreadCountReq{
    1:i64 userId
}

struct GetUnreadCountResp{
    1:common.BaseResp BaseResp
    2:i64 totalUnread
}

struct SystemNotification{
    1:i64 id
    2:i64 userId
    3:string title
    4:string content
    5:i32 type
    6:i64 relatedId
    7:bool isRead
    8:string createTime
}

struct GetNotificationsReq{
    1:i64 userId
    2:i32 page
    3:i32 pageSize
    4:optional i32 type
}

struct GetNotificationsResp{
    1:common.BaseResp BaseResp
    2:list<SystemNotification> notifications
    3:i32 totalCount
}

struct MarkNotificationReadReq{
    1:i64 userId
    2:i64 notificationId
}

struct MarkNotificationReadResp{
    1:common.BaseResp BaseResp
}

service MessageService{
    SendMessageResp SendMessage(1:SendMessageReq req)
    GetChatHistoryResp GetChatHistory(1:GetChatHistoryReq req)
    GetLatestMessagesResp GetLatestMessages(1:GetLatestMessagesReq req)
    MarkMessageReadResp MarkMessageRead(1:MarkMessageReadReq req)
    DeleteMessageResp DeleteMessage(1:DeleteMessageReq req)
    GetUnreadCountResp GetUnreadCount(1:GetUnreadCountReq req)
    GetNotificationsResp GetNotifications(1:GetNotificationsReq req)
    MarkNotificationReadResp MarkNotificationRead(1:MarkNotificationReadReq req)
}