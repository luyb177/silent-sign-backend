package message

import (
	"time"

	"gorm.io/plugin/soft_delete"
)

// Message 通用消息/通知
type Message struct {
	ID        uint64 `gorm:"primarykey"`
	CreatedAt time.Time
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:nano"`

	SenderID   uint64 // 发送者，0 表示系统消息
	ReceiverID uint64 `gorm:"index:idx_receiver_read"` // 私信接收者，群消息时为 0
	GroupID    uint64 `gorm:"index;default:0"`         // 群ID，0 表示私信
	Type       uint8  // 消息类型
	Content    string
	IsRead     bool `gorm:"index:idx_receiver_read;default:false"` // 是否已读（仅私信有效，群消息已读见 Redis msg_read:{id}）
}

func (Message) TableName() string {
	return "messages"
}
