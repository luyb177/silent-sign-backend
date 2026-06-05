// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package sse

import (
	"fmt"
	"net/http"

	"github.com/luyb177/silent-sign-backend/internal/middleware"
	"github.com/luyb177/silent-sign-backend/internal/svc"
)

// EventsHandler SSE 事件流
func EventsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authUser := middleware.GetAuthUser(r.Context())
		if authUser == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		ch := svcCtx.SSEHub.Register(authUser.UserID)
		defer svcCtx.SSEHub.Unregister(authUser.UserID)

		for {
			select {
			case data, ok := <-ch:
				if !ok {
					return
				}
				if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
					return
				}
				if flusher, ok := w.(http.Flusher); ok {
					flusher.Flush()
				}
			case <-r.Context().Done():
				return
			}
		}
	}
}
