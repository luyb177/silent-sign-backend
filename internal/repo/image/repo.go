package image

import (
	"context"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Repository interface {
	CreateBatch(ctx context.Context, images []*Image, tx ...*gorm.DB) error
	FindByTarget(ctx context.Context, targetType uint8, targetID uint64, tx ...*gorm.DB) ([]*Image, error)
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

func (r *repo) CreateBatch(ctx context.Context, images []*Image, tx ...*gorm.DB) error {
	db := r.getDB(ctx, tx...)
	return db.Create(&images).Error
}

func (r *repo) FindByTarget(ctx context.Context, targetType uint8, targetID uint64, tx ...*gorm.DB) ([]*Image, error) {
	db := r.getDB(ctx, tx...)
	var images []*Image
	err := db.Where("target_type = ? AND target_id = ?", targetType, targetID).
		Order("sort ASC").Find(&images).Error
	return images, err
}
