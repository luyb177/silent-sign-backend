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
	FindByID(ctx context.Context, id uint64, tx ...*gorm.DB) (*User, error)
	Update(ctx context.Context, id uint64, updates map[string]interface{}, tx ...*gorm.DB) error
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

func (r *repo) FindByID(ctx context.Context, id uint64, tx ...*gorm.DB) (*User, error) {
	db := r.getDB(ctx, tx...)

	var u User
	err := db.Where("id = ?", id).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &u, err
}

func (r *repo) Update(ctx context.Context, id uint64, updates map[string]interface{}, tx ...*gorm.DB) error {
	db := r.getDB(ctx, tx...)
	return db.Model(&User{}).Where("id = ?", id).Updates(updates).Error
}
