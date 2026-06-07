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
//  2. Sec-WebSocket-Protocol 头: 客户端通过子协议传 token，服务端回显该值完成握手
func ChatHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. 提取 token：优先 query 参数，兜底 Sec-WebSocket-Protocol 子协议头
		token := r.URL.Query().Get("token")
		subProto := ""
		if token == "" {
			subProto = r.Header.Get("Sec-WebSocket-Protocol")
			if subProto != "" {
				token = subProto
			}
		}
		if token == "" {
			logx.Errorf("ws: missing token from %s", r.RemoteAddr)
			http.Error(w, "missing token", http.StatusUnauthorized)
			return
		}

		// 2. 校验 JWT
		claims, err := svcCtx.JWTHandler.ParseJWTToken(token)
		if err != nil {
			logx.Errorf("ws: invalid token from %s: %v", r.RemoteAddr, err)
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		// 3. 如果通过子协议传 token，需要在响应中回显子协议值完成握手
		var respHeader http.Header
		if subProto != "" {
			respHeader = http.Header{}
			respHeader.Set("Sec-WebSocket-Protocol", subProto)
		}

		conn, err := upgrader.Upgrade(w, r, respHeader)
		if err != nil {
			logx.Errorf("ws: upgrade failed for user %d: %v", claims.UserID, err)
			return
		}

		logx.Infof("ws: user %d connected", claims.UserID)

		client := ws.NewClient(claims.UserID, svcCtx.WSHub, svcCtx.Repos, conn)
		svcCtx.WSHub.Register(client)

		go client.WritePump()
		go client.ReadPump()
	}
}

// extractBearerToken 从 Authorization 头提取 Bearer token（备用工具函数）
func extractBearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return auth[7:]
	}
	return ""
}
