package like

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// key 格式: like:{targetType}:{targetID}
const likeKeyPrefix = "like"

func likeKey(targetType uint8, targetID uint64) string {
	return fmt.Sprintf("%s:%d:%d", likeKeyPrefix, targetType, targetID)
}

// ParseLikeKey 解析 like key 返回 targetType 和 targetID
func ParseLikeKey(key string) (targetType uint8, targetID uint64, ok bool) {
	parts := strings.Split(key, ":")
	if len(parts) != 3 || parts[0] != likeKeyPrefix {
		return 0, 0, false
	}
	t, err := strconv.ParseUint(parts[1], 10, 8)
	if err != nil {
		return 0, 0, false
	}
	id, err := strconv.ParseUint(parts[2], 10, 64)
	if err != nil {
		return 0, 0, false
	}
	return uint8(t), id, true
}

type Repository interface {
	Toggle(ctx context.Context, targetType uint8, targetID, userID uint64) (liked bool, count int64, err error)
	IsLiked(ctx context.Context, targetType uint8, targetID, userID uint64) (bool, error)
	Count(ctx context.Context, targetType uint8, targetID uint64) (int64, error)
	// SyncTarget 将指定目标的 Redis 点赞数据全量同步到 MySQL
	SyncTarget(ctx context.Context, targetType uint8, targetID uint64, tx ...*gorm.DB) error
	// ScanKeys 扫描所有 like key（供定时任务使用）
	ScanKeys(ctx context.Context, cursor uint64, count int64) (keys []string, nextCursor uint64, err error)
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

// Toggle 点赞/取消点赞（Redis Set），返回当前是否已赞及点赞总数
func (r *repo) Toggle(ctx context.Context, targetType uint8, targetID, userID uint64) (liked bool, count int64, err error) {
	key := likeKey(targetType, targetID)
	uid := strconv.FormatUint(userID, 10)

	// 先检查是否已点赞
	exists, err := r.client.SIsMember(ctx, key, uid).Result()
	if err != nil {
		return false, 0, err
	}

	if exists {
		// 取消点赞
		if err = r.client.SRem(ctx, key, uid).Err(); err != nil {
			return false, 0, err
		}
	} else {
		// 点赞
		if err = r.client.SAdd(ctx, key, uid).Err(); err != nil {
			return false, 0, err
		}
	}

	count, err = r.client.SCard(ctx, key).Result()
	if err != nil {
		return false, 0, err
	}
	return !exists, count, nil
}

// IsLiked 检查用户是否已点赞
func (r *repo) IsLiked(ctx context.Context, targetType uint8, targetID, userID uint64) (bool, error) {
	key := likeKey(targetType, targetID)
	return r.client.SIsMember(ctx, key, strconv.FormatUint(userID, 10)).Result()
}

// Count 获取目标点赞数
func (r *repo) Count(ctx context.Context, targetType uint8, targetID uint64) (int64, error) {
	return r.client.SCard(ctx, likeKey(targetType, targetID)).Result()
}

// SyncTarget 将指定目标的 Redis 点赞数据全量同步到 MySQL
func (r *repo) SyncTarget(ctx context.Context, targetType uint8, targetID uint64, tx ...*gorm.DB) error {
	db := r.getDB(ctx, tx...)
	key := likeKey(targetType, targetID)

	// 获取 Redis 中所有点赞用户 ID
	members, err := r.client.SMembers(ctx, key).Result()
	if err != nil {
		return err
	}

	redisUserIDs := make(map[uint64]struct{}, len(members))
	for _, m := range members {
		uid, _ := strconv.ParseUint(m, 10, 64)
		redisUserIDs[uid] = struct{}{}
	}

	// 获取 MySQL 中该目标的所有有效点赞
	var dbLikes []Like
	if err := db.Where("target_type = ? AND target_id = ?", targetType, targetID).Find(&dbLikes).Error; err != nil {
		return err
	}
	dbUserIDs := make(map[uint64]uint64, len(dbLikes)) // userID → likeID
	for _, l := range dbLikes {
		dbUserIDs[l.UserID] = l.ID
	}

	// 新增 Redis 有但 MySQL 没有的
	for uid := range redisUserIDs {
		if _, ok := dbUserIDs[uid]; !ok {
			if err := db.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "target_type"}, {Name: "target_id"}, {Name: "user_id"}},
				DoNothing: true,
			}).Create(&Like{
				TargetType: targetType,
				TargetID:   targetID,
				UserID:     uid,
			}).Error; err != nil {
				return err
			}
		}
	}

	// 删除 MySQL 有但 Redis 没有的（已取消点赞）
	for uid, likeID := range dbUserIDs {
		if _, ok := redisUserIDs[uid]; !ok {
			if err := db.Where("id = ?", likeID).Delete(&Like{}).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

// ScanKeys 扫描所有 like:* key
func (r *repo) ScanKeys(ctx context.Context, cursor uint64, count int64) (keys []string, nextCursor uint64, err error) {
	keys, nextCursor, err = r.client.Scan(ctx, cursor, likeKeyPrefix+":*", count).Result()
	return
}
