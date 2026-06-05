package ws

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
	"github.com/luyb177/silent-sign-backend/internal/constvar"
	"github.com/luyb177/silent-sign-backend/internal/repo"
	messageRepo "github.com/luyb177/silent-sign-backend/internal/repo/message"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4096
)

// WSMessage WebSocket 消息格式
type WSMessage struct {
	Type     string `json:"type"`       // "chat"
	ToUserID uint64 `json:"to_user_id"` // 客户端→服务端
	Content  string `json:"content"`    // 客户端→服务端 / 服务端→客户端

	// 服务端→客户端 填充
	MessageID  uint64 `json:"message_id,omitempty"`
	FromUserID uint64 `json:"from_user_id,omitempty"`
	CreatedAt  string `json:"created_at,omitempty"`
}

// Client 代表一个 WebSocket 连接
type Client struct {
	userID uint64
	hub    *Hub
	repos  *repo.Repositories
	conn   *websocket.Conn
	send   chan []byte
}

func NewClient(userID uint64, hub *Hub, repos *repo.Repositories, conn *websocket.Conn) *Client {
	return &Client{
		userID: userID,
		hub:    hub,
		repos:  repos,
		conn:   conn,
		send:   make(chan []byte, 64),
	}
}

// ReadPump 从 WebSocket 读取消息
func (c *Client) ReadPump() {
	defer func() {
		c.hub.Unregister(c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	err := c.conn.SetReadDeadline(time.Now().Add(pongWait))
	if err != nil {
		return
	}
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			break
		}

		var wsMsg WSMessage
		if err := json.Unmarshal(msg, &wsMsg); err != nil {
			continue
		}

		c.handleMessage(&wsMsg)
	}
}

// WritePump 向 WebSocket 写入消息
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case data, ok := <-c.send:
			err := c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err != nil {
				return
			}
			if !ok {
				err := c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				if err != nil {
					return
				}
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}
		case <-ticker.C:
			err := c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err != nil {
				return
			}
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage 处理收到的聊天消息
func (c *Client) handleMessage(wsMsg *WSMessage) {
	if wsMsg.Type != "chat" || wsMsg.Content == "" || wsMsg.ToUserID == 0 {
		return
	}

	// 存消息
	m := &messageRepo.Message{
		SenderID:   c.userID,
		ReceiverID: wsMsg.ToUserID,
		Type:       constvar.MsgTypeChat,
		Content:    wsMsg.Content,
	}
	if err := c.repos.Message.Create(context.Background(), m); err != nil {
		logx.Errorf("ws save message failed: %v", err)
		return
	}

	// 组装推送消息
	resp := WSMessage{
		Type:       "chat",
		MessageID:  m.ID,
		FromUserID: c.userID,
		Content:    wsMsg.Content,
		CreatedAt:  m.CreatedAt.Format(time.DateTime),
	}
	data, _ := json.Marshal(resp)

	// 推送给接收方
	c.hub.SendToUser(wsMsg.ToUserID, data)
	// 回显给发送方
	c.hub.SendToUser(c.userID, data)
}
