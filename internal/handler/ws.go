package handler

import (
	"net/http"

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
func ChatHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if token == "" {
			http.Error(w, "missing token", http.StatusUnauthorized)
			return
		}

		claims, err := svcCtx.JWTHandler.ParseJWTToken(token)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logx.Errorf("ws upgrade failed: %v", err)
			return
		}

		client := ws.NewClient(claims.UserID, svcCtx.WSHub, svcCtx.Repos, conn)
		svcCtx.WSHub.Register(client)

		go client.WritePump()
		go client.ReadPump()
	}
}
