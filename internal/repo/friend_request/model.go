package friendrequest

import (
	"time"

	"gorm.io/plugin/soft_delete"
)

// FriendRequest 好友申请
type FriendRequest struct {
	ID        uint64 `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:nano"`

	FromUserID uint64 `gorm:"uniqueIndex:idx_from_to"` // 发起方
	ToUserID   uint64 `gorm:"uniqueIndex:idx_from_to"` // 接收方
	Status     uint8  `gorm:"default:1"`               // 1=pending, 2=accepted, 3=rejected
}

func (FriendRequest) TableName() string {
	return "friend_requests"
}
