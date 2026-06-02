package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/luyb177/silent-sign-backend/common/errorx"
	"github.com/luyb177/silent-sign-backend/common/jwtx"
	"github.com/luyb177/silent-sign-backend/common/respx"
	"github.com/luyb177/silent-sign-backend/internal/constvar"
	"github.com/luyb177/silent-sign-backend/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type AuthMiddleware struct {
	handler jwtx.Handler
	logx.Logger
}

func NewAuthMiddleware(handler jwtx.Handler) *AuthMiddleware {
	return &AuthMiddleware{
		handler: handler,
		Logger:  logx.WithContext(nil),
	}
}

func (m *AuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := parseAuthorizationToken(r.Header)
		if token == "" {
			respx.ErrorCtx(r.Context(), w, errorx.ErrUnauthorized)
			return
		}

		claims, err := m.handler.ParseJWTToken(token)
		if err != nil {
			m.Errorf("ParseJWTToken 解析token失败: %v", err)
			respx.ErrorCtx(r.Context(), w, errorx.Wrap(errorx.CodeUnauthorized, "token令牌无效", err))
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
	s := strings.TrimSpace(h.Get("Authorization"))
	if s == "" {
		return ""
	}
	const prefix = "Bearer "
	if len(s) > len(prefix) && strings.EqualFold(s[:len(prefix)], prefix) {
		return strings.TrimSpace(s[len(prefix):])
	}
	return s
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
