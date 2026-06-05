package sse

import (
	"sync"
)

// Hub 管理所有 SSE 客户端连接，按 userID 路由推送
type Hub struct {
	mu      sync.RWMutex
	clients map[uint64]chan []byte
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[uint64]chan []byte),
	}
}

// Register 为用户注册一个 SSE 通道，返回只读 channel（同一用户重复注册会替换旧连接）
func (h *Hub) Register(userID uint64) <-chan []byte {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 关闭旧连接（如果存在），避免泄漏
	if old, ok := h.clients[userID]; ok {
		close(old)
	}
	ch := make(chan []byte, 64)
	h.clients[userID] = ch
	return ch
}

// Unregister 移除用户的 SSE 通道并关闭
func (h *Hub) Unregister(userID uint64) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if ch, ok := h.clients[userID]; ok {
		close(ch)
		delete(h.clients, userID)
	}
}

// PushToUser 向指定用户推送数据（非阻塞，channel 满时丢弃）
func (h *Hub) PushToUser(userID uint64, data []byte) {
	h.mu.RLock()
	ch, ok := h.clients[userID]
	h.mu.RUnlock()
	if !ok {
		return
	}
	select {
	case ch <- data:
	default:
		// channel 满，丢弃（避免阻塞业务逻辑）
	}
}
