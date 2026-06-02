// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package user

import (
	"context"
	"strings"

	"github.com/luyb177/silent-sign-backend/common/errorx"
	"github.com/luyb177/silent-sign-backend/internal/constvar"
	"github.com/luyb177/silent-sign-backend/internal/middleware"
	"github.com/luyb177/silent-sign-backend/internal/pkg/password"
	"github.com/luyb177/silent-sign-backend/internal/repo/verify"
	"github.com/luyb177/silent-sign-backend/internal/svc"
	"github.com/luyb177/silent-sign-backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChangePasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewChangePasswordLogic 修改密码
func NewChangePasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChangePasswordLogic {
	return &ChangePasswordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChangePasswordLogic) ChangePassword(req *types.ChangePasswordReq) (resp *types.Response, err error) {
	// 1. 参数校验
	if err := l.validReq(req); err != nil {
		return nil, err
	}

	// 2. 从 token 获取当前用户
	authUser := middleware.GetAuthUser(l.ctx)
	if authUser == nil {
		return nil, errorx.ErrUnauthorized
	}

	// 3. 查询用户获取邮箱
	u, err := l.svcCtx.Repos.User.FindByID(l.ctx, authUser.UserID)
	if err != nil {
		l.Errorf("find user by id failed: %v", err)
		return nil, errorx.WrapDBQuery("查询用户失败", err)
	}
	if u == nil {
		return nil, errorx.ErrNotFound
	}

	// 4. 校验验证码
	meta := &verify.Meta{
		Target:  u.Email,
		Channel: constvar.ChannelEmail,
		Purpose: constvar.PurposePasswordReset,
	}
	ok, err := l.svcCtx.Repos.Verify.VerifyCode(l.ctx, meta, req.Code)
	if err != nil {
		l.Errorf("verify code failed: %v", err)
		return nil, errorx.WrapInternal("验证码校验失败", err)
	}
	if !ok {
		return nil, errorx.WrapBadRequest("验证码错误或已过期", nil)
	}

	// 5. 新密码哈希
	hashed, err := password.Hash(req.NewPassword)
	if err != nil {
		l.Errorf("hash password failed: %v", err)
		return nil, errorx.WrapInternal("密码加密失败", err)
	}

	// 6. 更新密码
	updates := map[string]interface{}{"password": hashed}
	if err := l.svcCtx.Repos.User.Update(l.ctx, authUser.UserID, updates); err != nil {
		l.Errorf("update password failed: %v", err)
		return nil, errorx.WrapDBUpdate("修改密码失败", err)
	}

	return &types.Response{}, nil
}

func (l *ChangePasswordLogic) validReq(req *types.ChangePasswordReq) error {
	switch {
	case strings.TrimSpace(req.Code) == "":
		return errorx.WrapBadRequest("验证码不能为空", nil)
	case len(req.Code) != 6:
		return errorx.WrapBadRequest("验证码长度不正确", nil)
	case req.NewPassword == "":
		return errorx.WrapBadRequest("新密码不能为空", nil)
	case len(req.NewPassword) < 6 || len(req.NewPassword) > 128:
		return errorx.WrapBadRequest("密码长度应在6-128位之间", nil)
	}
	return nil
}
