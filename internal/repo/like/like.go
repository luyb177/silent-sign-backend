package like

import (
	"time"

	"gorm.io/plugin/soft_delete"
)

type Like struct {
	ID        uint64 `gorm:"primarykey"`
	CreatedAt time.Time
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:nano"`

	TargetType uint8  `gorm:"uniqueIndex:idx_target_user"`
	TargetID   uint64 `gorm:"uniqueIndex:idx_target_user"`
	UserID     uint64 `gorm:"uniqueIndex:idx_target_user"`
}

func (Like) TableName() string {
	return "likes"
}
