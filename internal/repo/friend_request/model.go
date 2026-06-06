package friendrequest

import (
	"time"

	"gorm.io/plugin/soft_delete"
)

// FriendRequest 好友申请
type FriendRequest struct {
	ID        uint64                `gorm:"primarykey;type:bigint unsigned auto_increment"`
	CreatedAt time.Time             `gorm:"type:datetime(3)"`
	UpdatedAt time.Time             `gorm:"type:datetime(3)"`
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:nano;type:bigint unsigned"`

	FromUserID uint64 `gorm:"uniqueIndex:idx_from_to;type:bigint unsigned"` // 发起方
	ToUserID   uint64 `gorm:"uniqueIndex:idx_from_to;type:bigint unsigned"` // 接收方
	Status     uint8  `gorm:"default:1;type:tinyint unsigned"`              // 1=pending, 2=accepted, 3=rejected
}

func (FriendRequest) TableName() string {
	return "friend_requests"
}
