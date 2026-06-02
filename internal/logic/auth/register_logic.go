// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/luyb177/silent-sign-backend/common/errorx"
	"github.com/luyb177/silent-sign-backend/internal/constvar"
	"github.com/luyb177/silent-sign-backend/internal/pkg/email"
	"github.com/luyb177/silent-sign-backend/internal/pkg/password"
	"github.com/luyb177/silent-sign-backend/internal/repo/user"
	"github.com/luyb177/silent-sign-backend/internal/repo/verify"
	"github.com/luyb177/silent-sign-backend/internal/svc"
	"github.com/luyb177/silent-sign-backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewRegisterLogic 注册账号
func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterLogic) Register(req *types.RegisterReq) (resp *types.Response, err error) {
	l.Info("Auth Register")

	// 1. 参数校验
	if errResp, err := l.validReq(req); err != nil {
		return errResp, err
	}

	target := email.CanonicalEmail(req.Target)

	// 2. 根据渠道分发（当前仅支持邮箱）
	switch req.Channel {
	case constvar.ChannelEmail:
		return l.registerByEmail(target, req)
	default:
		return nil, errorx.WrapBadRequest("暂仅支持邮箱注册", nil)
	}
}

func (l *RegisterLogic) registerByEmail(target string, req *types.RegisterReq) (*types.Response, error) {
	// 1. 校验验证码
	meta := &verify.Meta{
		Target:  target,
		Channel: constvar.ChannelEmail,
		Purpose: constvar.PurposeRegistration,
	}
	ok, err := l.svcCtx.Repos.Verify.VerifyCode(l.ctx, meta, req.Code)
	if err != nil {
		l.Errorf("verify code failed: %v", err)
		return nil, errorx.WrapInternal("验证码校验失败", err)
	}
	if !ok {
		return nil, errorx.WrapBadRequest("验证码错误或已过期", nil)
	}

	// 2. 密码哈希存储
	hashedPassword, err := password.Hash(req.Password)
	if err != nil {
		l.Errorf("hash password failed: %v", err)
		return nil, errorx.WrapInternal("密码加密失败", err)
	}

	// 3. 创建用户（Email 唯一索引兜底，不额外查询）
	u := &user.User{
		Name:     randomName(),
		Email:    target,
		Password: hashedPassword,
	}
	if err := l.svcCtx.Repos.User.Create(l.ctx, u); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, errorx.WrapBadRequest("该邮箱已被注册", nil)
		}
		l.Errorf("create user failed: %v", err)
		return nil, errorx.WrapDBInsert("创建用户失败", err)
	}

	// 4. 异步发送欢迎邮件
	go func() {
		if err := l.svcCtx.EmailSender.SendWelcomeEmail(l.ctx, target, u.Name); err != nil {
			l.Errorf("send welcome email failed: %v", err)
		}
	}()

	return &types.Response{}, nil
}

func (l *RegisterLogic) validReq(req *types.RegisterReq) (*types.Response, error) {
	target := strings.TrimSpace(req.Target)
	code := strings.TrimSpace(req.Code)
	pwd := req.Password

	switch {
	case target == "":
		return nil, errorx.WrapBadRequest("邮箱不能为空", nil)
	case len(target) > 254:
		return nil, errorx.WrapBadRequest("邮箱地址过长", nil)
	case !email.IsValidEmail(email.CanonicalEmail(target)):
		return nil, errorx.WrapBadRequest("无效的邮箱格式", nil)
	case code == "":
		return nil, errorx.WrapBadRequest("验证码不能为空", nil)
	case len(code) != 6:
		return nil, errorx.WrapBadRequest("验证码长度不正确", nil)
	case pwd == "":
		return nil, errorx.WrapBadRequest("密码不能为空", nil)
	case len(pwd) < 6:
		return nil, errorx.WrapBadRequest("密码长度不能少于6位", nil)
	case len(pwd) > 128:
		return nil, errorx.WrapBadRequest("密码长度不能超过128位", nil)
	}

	return nil, nil
}

// randomName 生成随机用户名，格式：用户_xxxx（8位 hex）
func randomName() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return "用户_" + hex.EncodeToString(b)
}
