// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package friend

import (
	"net/http"

	"github.com/luyb177/silent-sign-backend/common/errorx"
	"github.com/luyb177/silent-sign-backend/common/respx"
	"github.com/luyb177/silent-sign-backend/internal/logic/friend"
	"github.com/luyb177/silent-sign-backend/internal/svc"
	"github.com/luyb177/silent-sign-backend/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// ListFriendRequestsHandler 好友申请列表（支持按状态筛选）
func ListFriendRequestsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListFriendRequestsReq
		if err := httpx.Parse(r, &req); err != nil {
			respx.ErrorCtx(r.Context(), w, errorx.WrapBadRequest("请求参数解析失败", err))
			return
		}

		l := friend.NewListFriendRequestsLogic(r.Context(), svcCtx)
		resp, err := l.ListFriendRequests(&req)
		if err != nil {
			respx.ErrorCtx(r.Context(), w, err)
			return
		}
		respx.OkCtx(r.Context(), w, resp)
	}
}
