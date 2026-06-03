package image

import (
	"time"

	"gorm.io/plugin/soft_delete"
)

// Image 通用图片，通过 TargetType + TargetID 关联到不同业务实体
type Image struct {
	ID        uint64 `gorm:"primarykey"`
	CreatedAt time.Time
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:nano"`

	TargetType uint8  `gorm:"index:idx_target"` // 关联类型，如 "moment"
	TargetID   uint64 `gorm:"index:idx_target"`
	URL        string
	Sort       uint8 `gorm:"default:0"` // 排序
}

func (Image) TableName() string {
	return "images"
}
