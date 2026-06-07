package image

import (
	"time"

	"gorm.io/plugin/soft_delete"
)

// Image 通用图片，通过 TargetType + TargetID 关联到不同业务实体
type Image struct {
	ID        uint64                `gorm:"primarykey;type:bigint unsigned auto_increment"`
	CreatedAt time.Time             `gorm:"type:datetime(3)"`
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:nano;type:bigint unsigned"`

	TargetType uint8  `gorm:"index:idx_target;type:tinyint unsigned"` // 关联类型，如 "moment"
	TargetID   uint64 `gorm:"index:idx_target;type:bigint unsigned"`
	URL        string `gorm:"type:varchar(512)"`
	Sort       uint8  `gorm:"default:0;type:tinyint unsigned"` // 排序
}

func (Image) TableName() string {
	return "images"
}
