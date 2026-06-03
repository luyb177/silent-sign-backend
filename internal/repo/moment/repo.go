package moment

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, m *Moment, tx ...*gorm.DB) error
	Delete(ctx context.Context, userID, momentID uint64, tx ...*gorm.DB) error
	FindByID(ctx context.Context, id uint64, tx ...*gorm.DB) (*Moment, error)
	Update(ctx context.Context, id uint64, updates map[string]interface{}, tx ...*gorm.DB) error
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

func (r *repo) Delete(ctx context.Context, userID, momentID uint64, tx ...*gorm.DB) error {
	db := r.getDB(ctx, tx...)
	return db.Where("user_id = ? AND id = ?", userID, momentID).Delete(&Moment{}).Error
}

func (r *repo) FindByID(ctx context.Context, id uint64, tx ...*gorm.DB) (*Moment, error) {
	db := r.getDB(ctx, tx...)

	var m Moment
	err := db.Where("id = ?", id).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &m, err
}

func (r *repo) Update(ctx context.Context, id uint64, updates map[string]interface{}, tx ...*gorm.DB) error {
	db := r.getDB(ctx, tx...)
	return db.Model(&Moment{}).Where("id = ?", id).Updates(updates).Error
}
