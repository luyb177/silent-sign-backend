// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package friend

import (
	"context"

	"github.com/luyb177/silent-sign-backend/common/errorx"
	"github.com/luyb177/silent-sign-backend/internal/constvar"
	"github.com/luyb177/silent-sign-backend/internal/middleware"
	"github.com/luyb177/silent-sign-backend/internal/svc"
	"github.com/luyb177/silent-sign-backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RejectFriendRequestLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// RejectFriendRequestLogic 拒绝好友申请
func NewRejectFriendRequestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RejectFriendRequestLogic {
	return &RejectFriendRequestLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RejectFriendRequestLogic) RejectFriendRequest(req *types.HandleFriendRequestReq) (resp *types.Response, err error) {
	// 1. 获取当前用户
	authUser := middleware.GetAuthUser(l.ctx)
	if authUser == nil {
		return nil, errorx.ErrUnauthorized
	}

	// 2. 查找申请
	fr, err := l.svcCtx.Repos.FriendRequest.FindByID(l.ctx, req.RequestID)
	if err != nil {
		return nil, errorx.WrapDBQuery("查询申请失败", err)
	}
	if fr == nil {
		return nil, errorx.WrapBadRequest("申请不存在", nil)
	}
	if fr.ToUserID != authUser.UserID {
		return nil, errorx.ErrForbidden
	}
	if fr.Status != constvar.FriendRequestStatusPending {
		return nil, errorx.WrapBadRequest("申请已处理", nil)
	}

	// 3. 更新状态为已拒绝
	if err := l.svcCtx.Repos.FriendRequest.UpdateStatus(l.ctx, fr.ID, constvar.FriendRequestStatusRejected); err != nil {
		l.Errorf("reject friend request failed: %v", err)
		return nil, errorx.WrapDBUpdate("拒绝好友申请失败", err)
	}

	return &types.Response{}, nil
}
