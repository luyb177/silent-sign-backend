package comment

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository interface {
	Create(ctx context.Context, c *Comment, tx ...*gorm.DB) error
	Delete(ctx context.Context, creatorID, commentID uint64, tx ...*gorm.DB) error
	FindByID(ctx context.Context, id uint64, tx ...*gorm.DB) (*Comment, error)
	FindByIDForUpdate(ctx context.Context, id uint64, tx ...*gorm.DB) (*Comment, error)
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

func (r *repo) FindByID(ctx context.Context, id uint64, tx ...*gorm.DB) (*Comment, error) {
	db := r.getDB(ctx, tx...)

	var c Comment
	err := db.Where("id = ?", id).First(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &c, err
}

func (r *repo) FindByIDForUpdate(ctx context.Context, id uint64, tx ...*gorm.DB) (*Comment, error) {
	db := r.getDB(ctx, tx...)

	var c Comment
	err := db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", id).
		First(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &c, err
}
