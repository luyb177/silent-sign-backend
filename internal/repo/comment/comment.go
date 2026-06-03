package comment

import (
	"time"

	"gorm.io/plugin/soft_delete"
)

type Comment struct {
	ID        uint64 `gorm:"primarykey"`
	CreatedAt time.Time
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:nano"`

	TargetType uint8
	TargetID   uint64
	FatherID   uint64
	CreatorID  uint64
	Location   string // 发布动态时ip所在的位置

	Content string

	// 冗余字段
	LikeNum uint64
}

func (Comment) TableName() string {
	return "comments"
}
