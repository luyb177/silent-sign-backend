package friend

import (
	"time"

	"gorm.io/plugin/soft_delete"
)

// Friend 好友关系（双向各存一行）
type Friend struct {
	ID        uint64                `gorm:"primarykey;type:bigint unsigned auto_increment"`
	CreatedAt time.Time             `gorm:"type:datetime(3)"`
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:nano;uniqueIndex:idx_user_friend;type:bigint unsigned"`

	UserID   uint64 `gorm:"uniqueIndex:idx_user_friend;type:bigint unsigned"` // 用户
	FriendID uint64 `gorm:"uniqueIndex:idx_user_friend;type:bigint unsigned"` // 好友
}

func (Friend) TableName() string {
	return "friends"
}
