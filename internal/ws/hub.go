package ws

import (
	"sync"
)

// Hub 管理所有 WebSocket 连接，按 userID 路由
type Hub struct {
	mu      sync.RWMutex
	clients map[uint64]*Client
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[uint64]*Client),
	}
}

// Register 注册客户端（替换同一用户的旧连接，通过 closeChan 安全通知退出）
func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if old, ok := h.clients[c.userID]; ok {
		old.Close()
	}
	h.clients[c.userID] = c
}

// Unregister 移除客户端并通知退出
func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if cur, ok := h.clients[c.userID]; ok && cur == c {
		delete(h.clients, c.userID)
	}
	c.Close()
}

// SendToUser 向指定用户推送消息（非阻塞）
func (h *Hub) SendToUser(userID uint64, data []byte) {
	h.mu.RLock()
	c, ok := h.clients[userID]
	if !ok {
		h.mu.RUnlock()
		return
	}
	select {
	case c.send <- data:
	default:
		// channel 满，丢弃
	}
	h.mu.RUnlock()
}
