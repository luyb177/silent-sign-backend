// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package friend

import (
	"context"

	"github.com/luyb177/silent-sign-backend/common/errorx"
	"github.com/luyb177/silent-sign-backend/internal/middleware"
	"github.com/luyb177/silent-sign-backend/internal/svc"
	"github.com/luyb177/silent-sign-backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type DeleteFriendLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// DeleteFriendLogic 删除好友
func NewDeleteFriendLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteFriendLogic {
	return &DeleteFriendLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteFriendLogic) DeleteFriend(req *types.DeleteFriendReq) (resp *types.Response, err error) {
	// 1. 获取当前用户
	authUser := middleware.GetAuthUser(l.ctx)
	if authUser == nil {
		return nil, errorx.ErrUnauthorized
	}

	// 2. 事务内双向删除好友关系
	err = l.svcCtx.Repos.Transaction(func(tx *gorm.DB) error {
		if err := l.svcCtx.Repos.Friend.Delete(l.ctx, authUser.UserID, req.FriendID, tx); err != nil {
			return err
		}
		return l.svcCtx.Repos.Friend.Delete(l.ctx, req.FriendID, authUser.UserID, tx)
	})
	if err != nil {
		l.Errorf("delete friend failed: %v", err)
		return nil, errorx.WrapDBDelete("删除好友失败", err)
	}

	return &types.Response{}, nil
}
