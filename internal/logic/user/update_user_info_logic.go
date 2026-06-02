// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package user

import (
	"context"

	"github.com/luyb177/silent-sign-backend/common/errorx"
	"github.com/luyb177/silent-sign-backend/internal/middleware"
	"github.com/luyb177/silent-sign-backend/internal/svc"
	"github.com/luyb177/silent-sign-backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateUserInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewUpdateUserInfoLogic 修改用户信息
func NewUpdateUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserInfoLogic {
	return &UpdateUserInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateUserInfoLogic) UpdateUserInfo(req *types.UpdateUserInfoReq) (resp *types.Response, err error) {
	// 1. 从 token 中获取当前登录用户 ID
	authUser := middleware.GetAuthUser(l.ctx)
	if authUser == nil {
		return nil, errorx.ErrUnauthorized
	}

	// 2. 仅更新非零值字段（增量更新）
	updates := make(map[string]interface{})
	if req.Username != "" {
		updates["name"] = req.Username
	}
	if req.Avatar != "" {
		updates["avatar"] = req.Avatar
	}
	if len(updates) == 0 {
		return nil, errorx.WrapBadRequest("没有可更新的字段", nil)
	}

	// 3. 执行更新
	if err = l.svcCtx.Repos.User.Update(l.ctx, authUser.UserID, updates); err != nil {
		l.Errorf("update user failed: %v", err)
		return nil, errorx.WrapDBUpdate("更新用户信息失败", err)
	}

	return &types.Response{}, nil
}
