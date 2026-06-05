// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package message

import (
	"context"

	"github.com/luyb177/silent-sign-backend/common/errorx"
	"github.com/luyb177/silent-sign-backend/internal/middleware"
	"github.com/luyb177/silent-sign-backend/internal/svc"
	"github.com/luyb177/silent-sign-backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UnreadCountLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// UnreadCountLogic 未读消息数
func NewUnreadCountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnreadCountLogic {
	return &UnreadCountLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UnreadCountLogic) UnreadCount() (resp *types.UnreadCountResp, err error) {
	// 1. 获取当前用户
	authUser := middleware.GetAuthUser(l.ctx)
	if authUser == nil {
		return nil, errorx.ErrUnauthorized
	}

	// 2. 查询未读数
	count, err := l.svcCtx.Repos.Message.UnreadCount(l.ctx, authUser.UserID)
	if err != nil {
		l.Errorf("unread count failed: %v", err)
		return nil, errorx.WrapDBQuery("查询未读消息数失败", err)
	}

	return &types.UnreadCountResp{Count: count}, nil
}
