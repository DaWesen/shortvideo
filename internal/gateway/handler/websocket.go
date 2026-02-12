package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"shortvideo/kitex_gen/danmu"
	"shortvideo/kitex_gen/message"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/hertz-contrib/websocket"
)

// 全局WebSocket管理器
var (
	wsManager *WSManager
	once      sync.Once
)

// WebSocket消息类型
const (
	MessageTypeChat         = "chat"
	MessageTypeDanmu        = "danmu"
	MessageTypeNotification = "notification"
	MessageTypeLiveStatus   = "live_status"
)

// WebSocket消息结构
type WSMessage struct {
	Type    string          `json:"type"`
	Content json.RawMessage `json:"content"`
}

// WebSocket客户端
type WSClient struct {
	conn     *websocket.Conn
	userID   int64
	liveID   int64
	joinTime time.Time
	mutex    sync.Mutex
	isClosed bool
}

// WebSocket管理器
type WSManager struct {
	clients    map[*WSClient]bool
	broadcast  chan []byte
	register   chan *WSClient
	unregister chan *WSClient
	mutex      sync.Mutex
}

// 创建WebSocket管理器
func NewWSManager() *WSManager {
	return &WSManager{
		clients:    make(map[*WSClient]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *WSClient),
		unregister: make(chan *WSClient),
	}
}

// 初始化WebSocket管理器
func InitWSManager() *WSManager {
	once.Do(func() {
		wsManager = NewWSManager()
		wsManager.Start()
	})
	return wsManager
}

// 启动WebSocket管理器
func (manager *WSManager) Start() {
	go func() {
		for {
			select {
			case client := <-manager.register:
				manager.mutex.Lock()
				manager.clients[client] = true
				manager.mutex.Unlock()
				log.Printf("新的WebSocket连接: 用户ID=%d, 直播ID=%d", client.userID, client.liveID)
			case client := <-manager.unregister:
				if _, ok := manager.clients[client]; ok {
					manager.mutex.Lock()
					delete(manager.clients, client)
					manager.mutex.Unlock()
					client.conn.Close()
					log.Printf("WebSocket连接关闭: 用户ID=%d, 直播ID=%d", client.userID, client.liveID)
				}
			case message := <-manager.broadcast:
				manager.mutex.Lock()
				for client := range manager.clients {
					if err := client.conn.WriteMessage(websocket.TextMessage, message); err != nil {
						client.conn.Close()
						delete(manager.clients, client)
					}
				}
				manager.mutex.Unlock()
			}
		}
	}()
}

// 向指定直播的所有客户端广播消息
func (manager *WSManager) BroadcastToLive(liveID int64, message []byte) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	for client := range manager.clients {
		if client.liveID == liveID {
			if err := client.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				client.conn.Close()
				delete(manager.clients, client)
			}
		}
	}
}

// 向指定用户广播消息
func (manager *WSManager) BroadcastToUser(userID int64, message []byte) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	for client := range manager.clients {
		if client.userID == userID {
			if err := client.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				client.conn.Close()
				delete(manager.clients, client)
			}
		}
	}
}

// 处理WebSocket连接
func (h *HTTPHandler) HandleWebSocket(c context.Context, ctx *app.RequestContext) {
	userID := int64(0)
	liveID := int64(0)

	userIDStr := ctx.Query("user_id")
	liveIDStr := ctx.Query("live_id")

	if userIDStr != "" {
		if id, err := parseInt64(userIDStr); err == nil {
			userID = id
		}
	}

	if liveIDStr != "" {
		if id, err := parseInt64(liveIDStr); err == nil {
			liveID = id
		}
	}

	//升级HTTP连接为WebSocket连接
	upgrader := websocket.HertzUpgrader{}
	upgrader.Upgrade(ctx, func(conn *websocket.Conn) {
		//创建WebSocket客户端
		client := &WSClient{
			conn:     conn,
			userID:   userID,
			liveID:   liveID,
			joinTime: time.Now(),
			isClosed: false,
		}

		//注册客户端
		wsManager.register <- client

		//处理WebSocket消息
		for {
			//读取消息
			msgType, msg, err := conn.ReadMessage()
			if err != nil {
				if !websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket读取错误: %v", err)
				}
				break
			}

			//处理消息
			h.handleWSMessage(client, msgType, msg)
		}

		//注销客户端
		client.isClosed = true
		wsManager.unregister <- client
	})
}

// 处理WebSocket消息
func (h *HTTPHandler) handleWSMessage(client *WSClient, msgType int, msg []byte) {
	//解析消息
	var wsMsg WSMessage
	if err := json.Unmarshal(msg, &wsMsg); err != nil {
		log.Printf("解析WebSocket消息失败: %v", err)
		return
	}

	//根据消息类型处理
	switch wsMsg.Type {
	case MessageTypeChat:
		h.handleChatMessage(client, wsMsg.Content)
	case MessageTypeDanmu:
		h.handleDanmuMessage(client, wsMsg.Content)
	default:
		log.Printf("未知的消息类型: %s", wsMsg.Type)
	}
}

// 处理聊天消息
func (h *HTTPHandler) handleChatMessage(client *WSClient, content json.RawMessage) {
	//解析聊天消息
	var chatMsg struct {
		ReceiverID int64  `json:"receiver_id"`
		Content    string `json:"content"`
	}

	if err := json.Unmarshal(content, &chatMsg); err != nil {
		log.Printf("解析聊天消息失败: %v", err)
		return
	}

	//发送消息
	if h.clients.MessageClient != nil {
		sendReq := &message.SendMessageReq{
			SenderId:   client.userID,
			ReceiverId: chatMsg.ReceiverID,
			Content:    chatMsg.Content,
		}

		_, err := h.clients.MessageClient.SendMessage(context.Background(), sendReq)
		if err != nil {
			log.Printf("发送消息失败: %v", err)
			return
		}
	}

	//广播消息给接收者
	responseMsg := WSMessage{
		Type:    MessageTypeChat,
		Content: content,
	}

	responseData, _ := json.Marshal(responseMsg)
	wsManager.BroadcastToUser(chatMsg.ReceiverID, responseData)
}

// 处理弹幕消息
func (h *HTTPHandler) handleDanmuMessage(client *WSClient, content json.RawMessage) {
	//解析弹幕消息
	var danmuMsg struct {
		LiveID   int64  `json:"live_id"`
		Content  string `json:"content"`
		Color    string `json:"color"`
		Position int32  `json:"position"`
	}

	if err := json.Unmarshal(content, &danmuMsg); err != nil {
		log.Printf("解析弹幕消息失败: %v", err)
		return
	}

	//发送弹幕
	if h.clients.DanmuClient != nil {
		danmuReq := &danmu.SendDanmuReq{
			UserId:  client.userID,
			LiveId:  danmuMsg.LiveID,
			Content: danmuMsg.Content,
		}

		if danmuMsg.Color != "" {
			danmuReq.Color = &danmuMsg.Color
		}
		if danmuMsg.Position > 0 {
			danmuReq.Position = &danmuMsg.Position
		}

		_, err := h.clients.DanmuClient.SendDanmu(context.Background(), danmuReq)
		if err != nil {
			log.Printf("发送弹幕失败: %v", err)
			return
		}
	}

	//广播弹幕给直播间所有用户
	responseMsg := WSMessage{
		Type:    MessageTypeDanmu,
		Content: content,
	}

	responseData, _ := json.Marshal(responseMsg)
	wsManager.BroadcastToLive(danmuMsg.LiveID, responseData)
}

// 解析int64
func parseInt64(s string) (int64, error) {
	var id int64
	_, err := fmt.Sscanf(s, "%d", &id)
	return id, err
}
