// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/luyb177/silent-sign-backend/common/errorx"
	"github.com/luyb177/silent-sign-backend/common/jwtx"
	"github.com/luyb177/silent-sign-backend/common/respx"
	"github.com/luyb177/silent-sign-backend/internal/constvar"
	"github.com/luyb177/silent-sign-backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type JWTMiddleware struct {
	handler jwtx.Handler
	logx.Logger
}

func NewJWTMiddleware(handler jwtx.Handler) *JWTMiddleware {
	return &JWTMiddleware{
		handler: handler,
		Logger:  logx.WithContext(context.Background()),
	}
}

func (m *JWTMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := parseAuthorizationToken(r.Header)
		if token == "" {
			respx.ErrorCtx(r.Context(), w, errorx.ErrUnauthorized)
			return
		}

		claims, err := m.handler.ParseJWTToken(token)
		if err != nil {
			m.Errorf("ParseJWTToken 解析token失败: %v", err)
			if errors.Is(err, jwtv5.ErrTokenExpired) {
				respx.ErrorCtx(r.Context(), w, errorx.ErrTokenExpired)
			} else {
				respx.ErrorCtx(r.Context(), w, errorx.ErrTokenInvalid)
			}
			return
		}

		u := &types.AuthUser{
			UserID: claims.UserID,
		}

		ctx := context.WithValue(r.Context(), constvar.AuthUserKey, u)

		next(w, r.WithContext(ctx))
	}
}

// 裸 token / Bearer token
func parseAuthorizationToken(h http.Header) string {
	auth := strings.TrimSpace(h.Get("Authorization"))
	if auth == "" {
		return ""
	}

	parts := strings.Fields(auth)

	switch len(parts) {
	case 1:
		return parts[0]

	case 2:
		if strings.EqualFold(parts[0], "Bearer") {
			return parts[1]
		}
	}

	return ""
}

// GetAuthUser 从 context 中获取 AuthUser
func GetAuthUser(ctx context.Context) *types.AuthUser {
	v := ctx.Value(constvar.AuthUserKey)
	if v == nil {
		return nil
	}
	u, ok := v.(*types.AuthUser)
	if !ok {
		return nil
	}
	return u
}
