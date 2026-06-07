package comment

import (
	"time"

	"gorm.io/plugin/soft_delete"
)

type Comment struct {
	ID        uint64                `gorm:"primarykey;type:bigint unsigned auto_increment"`
	CreatedAt time.Time             `gorm:"type:datetime(3)"`
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:nano;type:bigint unsigned"`

	TargetType uint8  `gorm:"type:tinyint unsigned"`
	TargetID   uint64 `gorm:"type:bigint unsigned"`
	FatherID   uint64 `gorm:"type:bigint unsigned;default:0"`
	CreatorID  uint64 `gorm:"type:bigint unsigned"`
	Location   string `gorm:"type:varchar(255)"` // 发布动态时ip所在的位置

	Content string `gorm:"type:text"`

	// 冗余字段
	LikeNum uint64 `gorm:"type:bigint unsigned;default:0"`
	SubNum  uint64 `gorm:"type:bigint unsigned;default:0"`
}

func (Comment) TableName() string {
	return "comments"
}
