package user

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, user *User, tx ...*gorm.DB) error
	FindByEmail(ctx context.Context, email string, tx ...*gorm.DB) (*User, error)
}

type repo struct {
	client *redis.Client
	db     *gorm.DB
}

func NewRepository(client *redis.Client, db *gorm.DB) Repository {
	return &repo{
		client: client,
		db:     db,
	}
}

func (r *repo) Create(ctx context.Context, user *User, tx ...*gorm.DB) error {
	db := r.getDB(ctx, tx...)
	return db.Create(user).Error
}

func (r *repo) FindByEmail(ctx context.Context, email string, tx ...*gorm.DB) (*User, error) {
	db := r.getDB(ctx, tx...)

	var u User
	err := db.Where("email = ?", email).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &u, err
}
