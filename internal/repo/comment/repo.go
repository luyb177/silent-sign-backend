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
	DeleteByFatherID(ctx context.Context, fatherID uint64, tx ...*gorm.DB) error
	FindByID(ctx context.Context, id uint64, tx ...*gorm.DB) (*Comment, error)
	FindByIDForUpdate(ctx context.Context, id uint64, tx ...*gorm.DB) (*Comment, error)
	AdjustSubNum(ctx context.Context, id uint64, delta int, tx ...*gorm.DB) error
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

// DeleteByFatherID 删除父评论下的所有子评论（软删除）
func (r *repo) DeleteByFatherID(ctx context.Context, fatherID uint64, tx ...*gorm.DB) error {
	db := r.getDB(ctx, tx...)
	return db.Where("father_id = ?", fatherID).Delete(&Comment{}).Error
}

// AdjustSubNum 原子增减子评论数（delta 为正增、为负减）
func (r *repo) AdjustSubNum(ctx context.Context, id uint64, delta int, tx ...*gorm.DB) error {
	db := r.getDB(ctx, tx...)
	sign := "+"
	guard := ""
	if delta < 0 {
		sign = "-"
		guard = " AND sub_num > 0"
		delta = -delta
	}
	result := db.Model(&Comment{}).Where("id = ?"+guard, id).
		Update("sub_num", gorm.Expr("sub_num "+sign+" ?", delta))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
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
