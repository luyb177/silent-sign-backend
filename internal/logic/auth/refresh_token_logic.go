// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package auth

import (
	"context"

	"github.com/luyb177/silent-sign-backend/common/errorx"
	"github.com/luyb177/silent-sign-backend/common/jwtx"
	"github.com/luyb177/silent-sign-backend/internal/svc"
	"github.com/luyb177/silent-sign-backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RefreshTokenLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewRefreshTokenLogic 刷新令牌
func NewRefreshTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshTokenLogic {
	return &RefreshTokenLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RefreshTokenLogic) RefreshToken(req *types.RefreshTokenReq) (resp *types.RefreshTokenResp, err error) {
	// 1. 解析 refresh token
	claims, err := l.svcCtx.JWTHandler.ParseJWTToken(req.RefreshToken)
	if err != nil {
		l.Errorf("parse refresh token failed: %v", err)
		return nil, errorx.ErrTokenInvalid
	}

	// 2. 必须是 refresh 类型的 token
	if claims.TokenType != jwtx.TokenTypeRefresh {
		return nil, errorx.ErrTokenInvalid
	}

	// 3. 签发新的 access token + 轮换 refresh token
	accessToken, err := l.svcCtx.JWTHandler.SetAccessToken(jwtx.ClaimsParams{
		UserID: claims.UserID,
	})
	if err != nil {
		l.Errorf("generate access token failed: %v", err)
		return nil, errorx.WrapInternal("生成令牌失败", err)
	}

	refreshToken, err := l.svcCtx.JWTHandler.SetRefreshToken(jwtx.ClaimsParams{
		UserID: claims.UserID,
	})
	if err != nil {
		l.Errorf("generate refresh token failed: %v", err)
		return nil, errorx.WrapInternal("生成令牌失败", err)
	}

	return &types.RefreshTokenResp{
		Token:        accessToken,
		RefreshToken: refreshToken,
	}, nil
}
