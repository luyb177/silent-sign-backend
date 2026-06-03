package comment

import (
	"context"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, c *Comment, tx ...*gorm.DB) error
	Delete(ctx context.Context, creatorID, commentID uint64, tx ...*gorm.DB) error
}

type repo struct {
	db     *gorm.DB
	client *redis.Client
}

func NewRepository(db *gorm.DB, client *redis.Client) Repository {
	return &repo{
		db:     db,
		client: client,
	}
}

func (r *repo) Create(ctx context.Context, c *Comment, tx ...*gorm.DB) error {
	db := r.getDB(ctx, tx...)
	return db.Create(c).Error
}

func (r *repo) Delete(ctx context.Context, creatorID, commentID uint64, tx ...*gorm.DB) error {
	db := r.getDB(ctx, tx...)
	return db.Where("creator_id = ? AND id = ?", creatorID, commentID).Delete(&Comment{}).Error
}
