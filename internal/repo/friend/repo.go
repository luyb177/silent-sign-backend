package friend

import (
	"context"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, f *Friend, tx ...*gorm.DB) error
	Delete(ctx context.Context, userID, friendID uint64, tx ...*gorm.DB) error
	Exists(ctx context.Context, userID, friendID uint64, tx ...*gorm.DB) (bool, error)
	ListByUser(ctx context.Context, userID uint64, limit int, cursorID uint64, tx ...*gorm.DB) ([]*Friend, error)
}

type repo struct {
	db     *gorm.DB
	client *redis.Client
}

func NewRepository(db *gorm.DB, client *redis.Client) Repository {
	return &repo{db: db, client: client}
}

func (r *repo) Create(ctx context.Context, f *Friend, tx ...*gorm.DB) error {
	db := r.getDB(ctx, tx...)
	return db.Create(f).Error
}

func (r *repo) Delete(ctx context.Context, userID, friendID uint64, tx ...*gorm.DB) error {
	db := r.getDB(ctx, tx...)
	return db.Where("user_id = ? AND friend_id = ?", userID, friendID).Delete(&Friend{}).Error
}

func (r *repo) Exists(ctx context.Context, userID, friendID uint64, tx ...*gorm.DB) (bool, error) {
	db := r.getDB(ctx, tx...)
	var count int64
	err := db.Model(&Friend{}).Where("user_id = ? AND friend_id = ?", userID, friendID).Count(&count).Error
	return count > 0, err
}

// ListByUser 按创建时间倒序游标分页查询用户的好友列表
func (r *repo) ListByUser(ctx context.Context, userID uint64, limit int, cursorID uint64, tx ...*gorm.DB) ([]*Friend, error) {
	db := r.getDB(ctx, tx...)
	var friends []*Friend

	query := db.Where("user_id = ?", userID)
	if cursorID != 0 {
		query = query.Where("id < ?", cursorID)
	}
	err := query.Order("id DESC").Limit(limit).Find(&friends).Error
	return friends, err
}
