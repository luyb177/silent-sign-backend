package moment

import (
	"context"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, m *Moment, tx ...*gorm.DB) error
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

func (r *repo) Create(ctx context.Context, m *Moment, tx ...*gorm.DB) error {
	db := r.getDB(ctx, tx...)
	return db.Create(m).Error
}
