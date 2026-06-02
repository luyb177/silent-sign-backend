// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package auth

import (
	"context"
	"strings"

	"github.com/luyb177/silent-sign-backend/common/errorx"
	"github.com/luyb177/silent-sign-backend/internal/constvar"
	"github.com/luyb177/silent-sign-backend/internal/pkg/code"
	"github.com/luyb177/silent-sign-backend/internal/pkg/email"
	"github.com/luyb177/silent-sign-backend/internal/repo/verify"
	"github.com/luyb177/silent-sign-backend/internal/svc"
	"github.com/luyb177/silent-sign-backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendVerificationCodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewSendVerificationCodeLogic 发送验证码
func NewSendVerificationCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendVerificationCodeLogic {
	return &SendVerificationCodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SendVerificationCodeLogic) SendVerificationCode(req *types.SendVerificationCodeReq) (resp *types.Response, err error) {
	l.Info("Auth SendVerificationCode")

	if !validVerifyPurpose(req.Purpose) {
		return nil, errorx.WrapBadRequest("无效的验证码用途", nil)
	}

	switch req.Channel {
	case constvar.ChannelEmail:
		return l.sendVerificationEmailCode(req.Target, req.Purpose)
	default:
		return nil, errorx.WrapBadRequest("无效的验证码渠道", nil)
	}
}

func (l *SendVerificationCodeLogic) sendVerificationEmailCode(target string, purpose int32) (*types.Response, error) {
	if strings.TrimSpace(target) == "" {
		return nil, errorx.WrapBadRequest("邮箱地址不能为空", nil)
	}

	if len(target) > 254 {
		return nil, errorx.WrapBadRequest("邮箱地址过长", nil)
	}

	target = email.CanonicalEmail(target)

	if !email.IsValidEmail(target) {
		return nil, errorx.WrapBadRequest("无效的邮箱地址", nil)
	}

	// 生成 6 位验证码
	emailCode := code.EmailCode()

	// 存入 Redis
	meta := &verify.Meta{
		Target:  target,
		Channel: constvar.ChannelEmail,
		Purpose: purpose,
	}
	if err := l.svcCtx.Repos.Verify.SetCode(l.ctx, meta, emailCode, constvar.VerifyCodeExpire); err != nil {
		l.Errorf("set verify code failed: %v", err)
		return nil, errorx.WrapInternal("验证码存储失败", err)
	}

	// todo mq 发送邮件
	go func() {
		if err := l.svcCtx.EmailSender.SendVerifyCode(l.ctx, target, emailCode, int(constvar.VerifyCodeExpire.Minutes())); err != nil {
			l.Errorf("send verify code email failed: %v", err)
		}
	}()

	return &types.Response{}, nil
}

func validVerifyPurpose(purpose int32) bool {
	switch purpose {
	case constvar.PurposeRegistration,
		constvar.PurposePasswordReset:
		return true
	default:
		return false
	}
}
