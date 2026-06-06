package friend

import (
	"time"

	"gorm.io/plugin/soft_delete"
)

// Friend 好友关系（双向各存一行）
type Friend struct {
	ID        uint64 `gorm:"primarykey"`
	CreatedAt time.Time
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:nano;uniqueIndex:idx_user_friend"`

	UserID   uint64 `gorm:"uniqueIndex:idx_user_friend"` // 用户
	FriendID uint64 `gorm:"uniqueIndex:idx_user_friend"` // 好友
}

func (Friend) TableName() string {
	return "friends"
}
