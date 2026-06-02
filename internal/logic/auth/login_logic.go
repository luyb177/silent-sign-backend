// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package auth

import (
	"context"
	"strings"

	"github.com/luyb177/silent-sign-backend/common/errorx"
	"github.com/luyb177/silent-sign-backend/common/jwtx"
	"github.com/luyb177/silent-sign-backend/internal/constvar"
	"github.com/luyb177/silent-sign-backend/internal/middleware"
	"github.com/luyb177/silent-sign-backend/internal/pkg/email"
	"github.com/luyb177/silent-sign-backend/internal/pkg/password"
	"github.com/luyb177/silent-sign-backend/internal/repo/user"
	"github.com/luyb177/silent-sign-backend/internal/svc"
	"github.com/luyb177/silent-sign-backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewLoginLogic 登录账号
func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginReq) (resp *types.LoginResp, err error) {
	l.Info("Auth Login")

	// 1. 参数校验
	if errResp, err := l.validReq(req); err != nil {
		return errResp, err
	}

	target := email.CanonicalEmail(req.Target)

	// 2. 根据渠道分发（当前仅支持邮箱）
	switch req.Channel {
	case constvar.ChannelEmail:
		return l.loginByEmail(target, req)
	default:
		return nil, errorx.WrapBadRequest("暂仅支持邮箱登录", nil)
	}
}

func (l *LoginLogic) loginByEmail(target string, req *types.LoginReq) (*types.LoginResp, error) {
	// 1. 查找用户
	u, err := l.svcCtx.Repos.User.FindByEmail(l.ctx, target)
	if err != nil {
		l.Errorf("find user by email failed: %v", err)
		return nil, errorx.WrapDBQuery("查询用户失败", err)
	}
	if u == nil {
		return nil, errorx.WrapBadRequest("邮箱未注册", nil)
	}

	// 2. 校验密码
	if !password.Compare(req.Password, u.Password) {
		return nil, errorx.WrapBadRequest("密码错误", nil)
	}

	// 3. 生成 JWT
	token, err := l.svcCtx.JWTHandler.SetJWTToken(jwtx.ClaimsParams{
		UserID: u.ID,
	})
	if err != nil {
		l.Errorf("generate jwt failed: %v", err)
		return nil, errorx.WrapInternal("生成令牌失败", err)
	}

	// todo mq 4. 异步发送登录通知邮件
	go func(to, username string) {
		ip, location := l.getLoginLocation()
		if err := l.svcCtx.EmailSender.SendLoginNotification(l.ctx, to, username, ip, location); err != nil {
			l.Errorf("send login notification email failed: %v", err)
		}
	}(target, u.Name)

	return &types.LoginResp{
		Token:    token,
		UserInfo: l.buildUserInfo(u),
	}, nil
}

func (l *LoginLogic) getLoginLocation() (ip, location string) {
	loc := middleware.GetIPLocation(l.ctx)
	if loc == nil {
		return "未知", "未知"
	}
	return loc.IP, middleware.FullLocation(loc)
}

func (l *LoginLogic) buildUserInfo(u *user.User) types.UserInfo {
	return types.UserInfo{
		UserID:    u.ID,
		Username:  u.Name,
		Email:     u.Email,
		Avatar:    u.Avatar,
		CreatedAt: u.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (l *LoginLogic) validReq(req *types.LoginReq) (*types.LoginResp, error) {
	target := strings.TrimSpace(req.Target)
	pwd := req.Password

	switch {
	case target == "":
		return nil, errorx.WrapBadRequest("邮箱不能为空", nil)
	case len(target) > 254:
		return nil, errorx.WrapBadRequest("邮箱地址过长", nil)
	case !email.IsValidEmail(email.CanonicalEmail(target)):
		return nil, errorx.WrapBadRequest("无效的邮箱格式", nil)
	case pwd == "":
		return nil, errorx.WrapBadRequest("密码不能为空", nil)
	case len(pwd) < 6 || len(pwd) > 128:
		return nil, errorx.WrapBadRequest("密码长度应在6-128位之间", nil)
	}

	return nil, nil
}
