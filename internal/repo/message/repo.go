package message

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, m *Message, tx ...*gorm.DB) error
	FindByID(ctx context.Context, id uint64, tx ...*gorm.DB) (*Message, error)
	ListByReceiver(ctx context.Context, receiverID uint64, cursorID uint64, limit int, tx ...*gorm.DB) ([]*Message, error)
	// ListByPartner 查询两人之间的私聊消息（双向，按时间倒序游标分页），msgType=0 不过滤类型
	ListByPartner(ctx context.Context, userID, partnerID uint64, msgType uint8, cursorID uint64, limit int, tx ...*gorm.DB) ([]*Message, error)
	MarkRead(ctx context.Context, receiverID uint64, messageIDs []uint64, tx ...*gorm.DB) error
	UnreadCount(ctx context.Context, receiverID uint64, tx ...*gorm.DB) (int64, error)

	// 群消息已读（Redis）
	MarkGroupRead(ctx context.Context, messageID, userID uint64) error
	IsGroupRead(ctx context.Context, messageID, userID uint64) (bool, error)
	BatchIsGroupRead(ctx context.Context, messageID uint64, userIDs []uint64) (map[uint64]bool, error)
	GroupReadCount(ctx context.Context, messageID uint64) (int64, error)
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

func (r *repo) Create(ctx context.Context, m *Message, tx ...*gorm.DB) error {
	db := r.getDB(ctx, tx...)
	return db.Create(m).Error
}

func (r *repo) FindByID(ctx context.Context, id uint64, tx ...*gorm.DB) (*Message, error) {
	db := r.getDB(ctx, tx...)

	var m Message
	err := db.Where("id = ?", id).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &m, err
}

// ListByReceiver 按创建时间倒序分页查询用户收到的消息
func (r *repo) ListByReceiver(ctx context.Context, receiverID uint64, cursorID uint64, limit int, tx ...*gorm.DB) ([]*Message, error) {
	db := r.getDB(ctx, tx...)
	var messages []*Message

	query := db.Where("receiver_id = ?", receiverID)
	if cursorID != 0 {
		query = query.Where("id < ?", cursorID)
	}
	err := query.Order("id DESC").Limit(limit).Find(&messages).Error
	return messages, err
}

// ListByPartner 查询两人之间的私聊消息（双向筛选，按时间倒序游标分页）
// msgType=0 表示不过滤类型
func (r *repo) ListByPartner(ctx context.Context, userID, partnerID uint64, msgType uint8, cursorID uint64, limit int, tx ...*gorm.DB) ([]*Message, error) {
	db := r.getDB(ctx, tx...)
	var messages []*Message

	query := db.Where(
		"(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
		userID, partnerID, partnerID, userID,
	)
	if msgType != 0 {
		query = query.Where("type = ?", msgType)
	}
	if cursorID != 0 {
		query = query.Where("id < ?", cursorID)
	}
	err := query.Order("id DESC").Limit(limit).Find(&messages).Error
	return messages, err
}

// MarkRead 批量标记消息为已读
func (r *repo) MarkRead(ctx context.Context, receiverID uint64, messageIDs []uint64, tx ...*gorm.DB) error {
	if len(messageIDs) == 0 {
		return nil
	}
	db := r.getDB(ctx, tx...)
	return db.Model(&Message{}).
		Where("receiver_id = ? AND id IN ? AND is_read = false", receiverID, messageIDs).
		Update("is_read", true).Error
}

// UnreadCount 获取用户未读消息数
func (r *repo) UnreadCount(ctx context.Context, receiverID uint64, tx ...*gorm.DB) (int64, error) {
	db := r.getDB(ctx, tx...)
	var count int64
	err := db.Model(&Message{}).
		Where("receiver_id = ? AND is_read = false", receiverID).
		Count(&count).Error
	return count, err
}

// ══════════════════════════════════════════════
// 群消息已读（Redis）
// ══════════════════════════════════════════════

func groupReadKey(messageID uint64) string {
	return fmt.Sprintf("msg_read:%d", messageID)
}

// MarkGroupRead 标记用户已读某条群消息
func (r *repo) MarkGroupRead(ctx context.Context, messageID, userID uint64) error {
	return r.client.SAdd(ctx, groupReadKey(messageID), strconv.FormatUint(userID, 10)).Err()
}

// IsGroupRead 检查用户是否已读某条群消息
func (r *repo) IsGroupRead(ctx context.Context, messageID, userID uint64) (bool, error) {
	return r.client.SIsMember(ctx, groupReadKey(messageID), strconv.FormatUint(userID, 10)).Result()
}

// BatchIsGroupRead 批量检查多个用户是否已读同一条群消息
func (r *repo) BatchIsGroupRead(ctx context.Context, messageID uint64, userIDs []uint64) (map[uint64]bool, error) {
	key := groupReadKey(messageID)
	result := make(map[uint64]bool, len(userIDs))
	if len(userIDs) == 0 {
		return result, nil
	}

	pipe := r.client.Pipeline()
	cmds := make([]*redis.BoolCmd, len(userIDs))
	for i, uid := range userIDs {
		cmds[i] = pipe.SIsMember(ctx, key, strconv.FormatUint(uid, 10))
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return nil, err
	}

	for i, uid := range userIDs {
		ok, err := cmds[i].Result()
		if err != nil {
			return nil, err
		}
		result[uid] = ok
	}
	return result, nil
}

// GroupReadCount 获取群消息已读人数
func (r *repo) GroupReadCount(ctx context.Context, messageID uint64) (int64, error) {
	return r.client.SCard(ctx, groupReadKey(messageID)).Result()
}
