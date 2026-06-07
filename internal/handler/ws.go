package handler

import (
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/luyb177/silent-sign-backend/internal/svc"
	ws "github.com/luyb177/silent-sign-backend/internal/ws"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// RegisterWSRoute 注册 WebSocket 聊天路由
func RegisterWSRoute(server *rest.Server, svcCtx *svc.ServiceContext) {
	server.AddRoute(rest.Route{
		Method:  http.MethodGet,
		Path:    "/api/v1/ws/chat",
		Handler: ChatHandler(svcCtx),
	})
}

// ChatHandler WebSocket 聊天入口
// 鉴权方式（按优先级）：
//  1. query 参数: ws://host/api/v1/ws/chat?token=xxx（推荐）
//  2. Sec-WebSocket-Protocol 头: 仅在值为合法 JWT 格式时使用
func ChatHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logx.Infof("ws: new connection attempt from %s, query=%s, sec-websocket-protocol=%s",
			r.RemoteAddr, r.URL.RawQuery, r.Header.Get("Sec-WebSocket-Protocol"))

		// 1. 提取 token：优先 query 参数
		token := r.URL.Query().Get("token")
		tokenSrc := "query"
		if token != "" {
			logx.Infof("ws: got token from query, first 20 chars: %s", maskToken(token))
		}

		// 2. 兜底：从 Sec-WebSocket-Protocol 子协议头取（仅当值是合法 JWT 时）
		subProto := ""
		subProtoRaw := ""
		if token == "" {
			subProtoRaw = r.Header.Get("Sec-WebSocket-Protocol")
			if subProtoRaw != "" {
				// 去掉可能的 "Bearer " 前缀（Apifox 等工具会自动加）
				subProto = strings.TrimPrefix(subProtoRaw, "Bearer ")
				logx.Infof("ws: query token empty, trying Sec-WebSocket-Protocol: %s", maskToken(subProto))
				if isJWTFormat(subProto) {
					token = subProto
					tokenSrc = "subproto"
				} else {
					logx.Infof("ws: Sec-WebSocket-Protocol value is not a JWT, ignoring")
				}
			}
		}

		if token == "" {
			logx.Errorf("ws: missing token from %s (query=%s, subproto=%s)",
				r.RemoteAddr, r.URL.Query().Get("token"), subProto)
			http.Error(w, "missing token", http.StatusUnauthorized)
			return
		}

		// 3. 校验 JWT
		claims, err := svcCtx.JWTHandler.ParseJWTToken(token)
		if err != nil {
			logx.Errorf("ws: invalid token from %s (src=%s, token_prefix=%s): %v",
				r.RemoteAddr, tokenSrc, maskToken(token), err)
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		// 4. 如果通过子协议传 token，需要在响应中回显原始子协议值完成握手
		var respHeader http.Header
		if subProtoRaw != "" && tokenSrc == "subproto" {
			respHeader = http.Header{}
			respHeader.Set("Sec-WebSocket-Protocol", subProtoRaw)
		}

		conn, err := upgrader.Upgrade(w, r, respHeader)
		if err != nil {
			logx.Errorf("ws: upgrade failed for user %d: %v", claims.UserID, err)
			return
		}

		logx.Infof("ws: user %d connected successfully", claims.UserID)

		client := ws.NewClient(claims.UserID, svcCtx.WSHub, svcCtx.Repos, conn)
		svcCtx.WSHub.Register(client)

		go client.WritePump()
		go client.ReadPump()
	}
}

// isJWTFormat 判断字符串是否符合 JWT 格式（三段 base64url，用 . 分隔）
func isJWTFormat(s string) bool {
	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return false
	}
	for _, p := range parts {
		if len(p) == 0 {
			return false
		}
	}
	return true
}

// maskToken 脱敏打印 token（仅显示前 20 字符）
func maskToken(token string) string {
	if len(token) <= 20 {
		return token
	}
	return token[:20] + "..."
}
