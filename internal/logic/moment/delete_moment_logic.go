// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package moment

import (
	"context"

	"github.com/luyb177/silent-sign-backend/common/errorx"
	"github.com/luyb177/silent-sign-backend/internal/middleware"
	"github.com/luyb177/silent-sign-backend/internal/svc"
	"github.com/luyb177/silent-sign-backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteMomentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewDeleteMomentLogic 删除动态
func NewDeleteMomentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteMomentLogic {
	return &DeleteMomentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteMomentLogic) DeleteMoment(req *types.DeleteMomentReq) (resp *types.Response, err error) {
	// 1. 从 token 获取当前用户
	authUser := middleware.GetAuthUser(l.ctx)
	if authUser == nil {
		return nil, errorx.ErrUnauthorized
	}

	// 2. 联合用户ID和momentID删除动态，确保只能删除自己的
	if err = l.svcCtx.Repos.Moment.Delete(l.ctx, authUser.UserID, req.MomentID); err != nil {
		l.Errorf("delete moment failed: %v", err)
		return nil, errorx.WrapDBDelete("删除动态失败", err)
	}

	return &types.Response{}, nil
}
