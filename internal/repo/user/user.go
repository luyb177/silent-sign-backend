package user

import (
	"time"

	"gorm.io/plugin/soft_delete"
)

type User struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:nano"`

	// 可修改字段
	Avatar   string
	Name     string
	Email    string `gorm:"uniqueIndex"`
	Password string

	// 部分冗余字段，用于快速查询
}

func (User) TableName() string {
	return "users"
}
