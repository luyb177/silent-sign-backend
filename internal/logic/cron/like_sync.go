package cron

import (
	"context"

	likerepo "github.com/luyb177/silent-sign-backend/internal/repo/like"

	"github.com/luyb177/silent-sign-backend/internal/constvar"
	"github.com/luyb177/silent-sign-backend/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

// SyncLikes 全量同步 Redis 点赞数据到 MySQL，并更新目标 LikeNum + HotScore
func SyncLikes(svcCtx *svc.ServiceContext) error {
	ctx := context.Background()
	var cursor uint64

	for {
		keys, nextCursor, err := svcCtx.Repos.Like.ScanKeys(ctx, cursor, 100)
		if err != nil {
			return err
		}

		for _, key := range keys {
			targetType, targetID, ok := likerepo.ParseLikeKey(key)
			if !ok {
				logx.Errorf("invalid like key: %s", key)
				continue
			}

			// 同步 Redis → MySQL
			if err := svcCtx.Repos.Like.SyncTarget(ctx, targetType, targetID); err != nil {
				logx.Errorf("sync target failed (type=%d id=%d): %v", targetType, targetID, err)
				continue
			}

			// 同步后更新目标的 LikeNum + HotScore
			if err := updateTargetLikeNum(ctx, svcCtx, targetType, targetID); err != nil {
				logx.Errorf("update target like num failed (type=%d id=%d): %v", targetType, targetID, err)
			}

			logx.Infof("synced like for target type=%d id=%d", targetType, targetID)
		}

		if nextCursor == 0 {
			break
		}
		cursor = nextCursor
	}

	return nil
}

func updateTargetLikeNum(ctx context.Context, svcCtx *svc.ServiceContext, targetType uint8, targetID uint64) error {
	switch targetType {
	case constvar.TargetTypeMoment:
		count, err := svcCtx.Repos.Like.Count(ctx, targetType, targetID)
		if err != nil {
			return err
		}
		return svcCtx.Repos.Moment.SetLikeNum(ctx, targetID, count)
	}
	return nil
}
