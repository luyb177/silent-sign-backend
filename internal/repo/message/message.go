package message

import (
	"time"

	"gorm.io/plugin/soft_delete"
)

// Message 通用消息/通知
type Message struct {
	ID        uint64                `gorm:"primarykey;type:bigint unsigned auto_increment"`
	CreatedAt time.Time             `gorm:"type:datetime(3)"`
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:nano;type:bigint unsigned"`

	SenderID   uint64 `gorm:"type:bigint unsigned;default:0"`               // 发送者，0 表示系统消息
	ReceiverID uint64 `gorm:"index:idx_receiver_read;type:bigint unsigned"` // 私信接收者，群消息时为 0
	GroupID    uint64 `gorm:"index;default:0;type:bigint unsigned"`         // 群ID，0 表示私信
	Type       uint8  `gorm:"type:tinyint unsigned"`                        // 消息类型
	Content    string `gorm:"type:text"`
	IsRead     bool   `gorm:"index:idx_receiver_read;default:false;type:tinyint(1)"` // 是否已读
}

func (Message) TableName() string {
	return "messages"
}
