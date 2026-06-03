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
	IncrementCommentNum(ctx context.Context, id uint64, tx ...*gorm.DB) error
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

// IncrementCommentNum 原子增加评论数并刷新热度分数
func (r *repo) IncrementCommentNum(ctx context.Context, id uint64, tx ...*gorm.DB) error {
	db := r.getDB(ctx, tx...)
	result := db.Model(&Moment{}).Where("id = ?", id).Updates(map[string]interface{}{
		"comment_num": gorm.Expr("comment_num + 1"),
		"hot_score":   gorm.Expr("ROUND((like_num * 3 + (comment_num + 1) * 5 + share_num * 8 + UNIX_TIMESTAMP(created_at) / 45000.0) * 100) / 100"),
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
