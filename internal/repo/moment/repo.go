package moment

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, m *Moment, tx ...*gorm.DB) error
	Delete(ctx context.Context, userID, momentID uint64, tx ...*gorm.DB) error
	FindByID(ctx context.Context, id uint64, tx ...*gorm.DB) (*Moment, error)
	ListByCreatedAt(ctx context.Context, cursorID uint64, cursorTime time.Time, limit int, tx ...*gorm.DB) ([]*Moment, error)
	ListByHot(ctx context.Context, cursorID uint64, cursorScore float64, limit int, tx ...*gorm.DB) ([]*Moment, error)
	AdjustCommentNum(ctx context.Context, id uint64, delta int, tx ...*gorm.DB) error
	AdjustLikeNum(ctx context.Context, id uint64, delta int, tx ...*gorm.DB) error
	SetLikeNum(ctx context.Context, id uint64, count int64, tx ...*gorm.DB) error
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

// AdjustCommentNum 原子增减评论数并刷新热度（delta 为正增、为负减）
func (r *repo) AdjustCommentNum(ctx context.Context, id uint64, delta int, tx ...*gorm.DB) error {
	db := r.getDB(ctx, tx...)
	sign := "+"
	guard := ""
	if delta < 0 {
		sign = "-"
		guard = " AND comment_num > 0"
		delta = -delta
	}
	result := db.Model(&Moment{}).Where("id = ?"+guard, id).Updates(map[string]interface{}{
		"comment_num": gorm.Expr("comment_num "+sign+" ?", delta),
		"hot_score":   gorm.Expr("ROUND((like_num * 3 + (comment_num "+sign+" ?) * 5 + share_num * 8 + UNIX_TIMESTAMP(created_at) / 45000.0) * 100) / 100", delta),
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// AdjustLikeNum 原子增减点赞数并刷新热度
func (r *repo) AdjustLikeNum(ctx context.Context, id uint64, delta int, tx ...*gorm.DB) error {
	db := r.getDB(ctx, tx...)
	sign := "+"
	guard := ""
	if delta < 0 {
		sign = "-"
		guard = " AND like_num > 0"
		delta = -delta
	}
	result := db.Model(&Moment{}).Where("id = ?"+guard, id).Updates(map[string]interface{}{
		"like_num":  gorm.Expr("like_num "+sign+" ?", delta),
		"hot_score": gorm.Expr("ROUND(((like_num "+sign+" ?) * 3 + comment_num * 5 + share_num * 8 + UNIX_TIMESTAMP(created_at) / 45000.0) * 100) / 100", delta),
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// SetLikeNum 直接设置点赞数并刷新热度（用于 cron 全量同步）
func (r *repo) SetLikeNum(ctx context.Context, id uint64, count int64, tx ...*gorm.DB) error {
	db := r.getDB(ctx, tx...)
	result := db.Model(&Moment{}).Where("id = ?", id).Updates(map[string]interface{}{
		"like_num":  count,
		"hot_score": gorm.Expr("ROUND((? * 3 + comment_num * 5 + share_num * 8 + UNIX_TIMESTAMP(created_at) / 45000.0) * 100) / 100", count),
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ListByCreatedAt 按创建时间倒序游标分页（最新在前）
// cursorTime 为空时查首页
func (r *repo) ListByCreatedAt(ctx context.Context, cursorID uint64, cursorTime time.Time, limit int, tx ...*gorm.DB) ([]*Moment, error) {
	db := r.getDB(ctx, tx...)
	var moments []*Moment

	if cursorID == 0 {
		err := db.Order("created_at DESC, id DESC").Limit(limit).Find(&moments).Error
		return moments, err
	}

	err := db.Where(
		"created_at < ? OR (created_at = ? AND id < ?)", cursorTime, cursorTime, cursorID,
	).Order("created_at DESC, id DESC").Limit(limit).Find(&moments).Error
	return moments, err
}

// ListByHot 按热度倒序游标分页
func (r *repo) ListByHot(ctx context.Context, cursorID uint64, cursorScore float64, limit int, tx ...*gorm.DB) ([]*Moment, error) {
	db := r.getDB(ctx, tx...)
	var moments []*Moment

	if cursorID == 0 {
		err := db.Order("hot_score DESC, id DESC").Limit(limit).Find(&moments).Error
		return moments, err
	}

	err := db.Where(
		"hot_score < ? OR (hot_score = ? AND id < ?)", cursorScore, cursorScore, cursorID,
	).Order("hot_score DESC, id DESC").Limit(limit).Find(&moments).Error
	return moments, err
}
