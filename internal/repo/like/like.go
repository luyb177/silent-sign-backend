package like

import (
	"time"

	"gorm.io/plugin/soft_delete"
)

type Like struct {
	ID        uint64                `gorm:"primarykey;type:bigint unsigned auto_increment"`
	CreatedAt time.Time             `gorm:"type:datetime(3)"`
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:nano;type:bigint unsigned"`

	TargetType uint8  `gorm:"uniqueIndex:idx_target_user;type:tinyint unsigned"`
	TargetID   uint64 `gorm:"uniqueIndex:idx_target_user;type:bigint unsigned"`
	UserID     uint64 `gorm:"uniqueIndex:idx_target_user;type:bigint unsigned"`
}

func (Like) TableName() string {
	return "likes"
}
