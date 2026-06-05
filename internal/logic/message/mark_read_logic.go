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

type MarkReadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// MarkReadLogic 标记已读
func NewMarkReadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MarkReadLogic {
	return &MarkReadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MarkReadLogic) MarkRead(req *types.MarkReadReq) (resp *types.Response, err error) {
	// 1. 获取当前用户
	authUser := middleware.GetAuthUser(l.ctx)
	if authUser == nil {
		return nil, errorx.ErrUnauthorized
	}

	// 2. 参数校验
	if len(req.MessageIDs) == 0 {
		return nil, errorx.WrapBadRequest("消息ID不能为空", nil)
	}

	// 3. 标记已读
	if err := l.svcCtx.Repos.Message.MarkRead(l.ctx, authUser.UserID, req.MessageIDs); err != nil {
		l.Errorf("mark read failed: %v", err)
		return nil, errorx.WrapDBUpdate("标记已读失败", err)
	}

	return &types.Response{}, nil
}
