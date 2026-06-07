package friendrequest

import (
	"context"
	"errors"

	"github.com/luyb177/silent-sign-backend/internal/constvar"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, r *FriendRequest, tx ...*gorm.DB) error
	FindByID(ctx context.Context, id uint64, tx ...*gorm.DB) (*FriendRequest, error)
	FindPending(ctx context.Context, fromUserID, toUserID uint64, tx ...*gorm.DB) (*FriendRequest, error)
	UpdateStatus(ctx context.Context, id uint64, status uint8, tx ...*gorm.DB) error
	ListPendingToUser(ctx context.Context, toUserID uint64, limit int, cursorID uint64, tx ...*gorm.DB) ([]*FriendRequest, error)
	// ListToUser 查询发给某用户的好友申请（支持按状态筛选，0 表示全部）
	ListToUser(ctx context.Context, toUserID uint64, status uint8, limit int, cursorID uint64, tx ...*gorm.DB) ([]*FriendRequest, error)
}

type repo struct {
	db     *gorm.DB
	client *redis.Client
}

func NewRepository(db *gorm.DB, client *redis.Client) Repository {
	return &repo{db: db, client: client}
}

func (r *repo) Create(ctx context.Context, fr *FriendRequest, tx ...*gorm.DB) error {
	db := r.getDB(ctx, tx...)
	return db.Create(fr).Error
}

func (r *repo) FindByID(ctx context.Context, id uint64, tx ...*gorm.DB) (*FriendRequest, error) {
	db := r.getDB(ctx, tx...)
	var fr FriendRequest
	err := db.Where("id = ?", id).First(&fr).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &fr, err
}

// FindPending 查找两人之间 pending 状态的申请（双向检查）
func (r *repo) FindPending(ctx context.Context, fromUserID, toUserID uint64, tx ...*gorm.DB) (*FriendRequest, error) {
	db := r.getDB(ctx, tx...)
	var fr FriendRequest
	err := db.Where(
		"((from_user_id = ? AND to_user_id = ?) OR (from_user_id = ? AND to_user_id = ?)) AND status = ?",
		fromUserID, toUserID, toUserID, fromUserID, constvar.FriendRequestStatusPending,
	).First(&fr).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &fr, err
}

func (r *repo) UpdateStatus(ctx context.Context, id uint64, status uint8, tx ...*gorm.DB) error {
	db := r.getDB(ctx, tx...)
	return db.Model(&FriendRequest{}).Where("id = ?", id).Update("status", status).Error
}

// ListPendingToUser 查询发给某用户的所有待处理申请（按时间倒序游标分页）
func (r *repo) ListPendingToUser(ctx context.Context, toUserID uint64, limit int, cursorID uint64, tx ...*gorm.DB) ([]*FriendRequest, error) {
	db := r.getDB(ctx, tx...)
	var requests []*FriendRequest

	query := db.Where("to_user_id = ? AND status = ?", toUserID, constvar.FriendRequestStatusPending)
	if cursorID != 0 {
		query = query.Where("id < ?", cursorID)
	}
	err := query.Order("id DESC").Limit(limit).Find(&requests).Error
	return requests, err
}

// ListToUser 查询发给某用户的好友申请（支持按状态筛选，status=0 表示全部，按时间倒序游标分页）
func (r *repo) ListToUser(ctx context.Context, toUserID uint64, status uint8, limit int, cursorID uint64, tx ...*gorm.DB) ([]*FriendRequest, error) {
	db := r.getDB(ctx, tx...)
	var requests []*FriendRequest

	query := db.Where("to_user_id = ?", toUserID)
	if status != 0 {
		query = query.Where("status = ?", status)
	}
	if cursorID != 0 {
		query = query.Where("id < ?", cursorID)
	}
	err := query.Order("id DESC").Limit(limit).Find(&requests).Error
	return requests, err
}
