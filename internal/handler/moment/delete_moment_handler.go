// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package moment

import (
	"net/http"

	"github.com/luyb177/silent-sign-backend/common/errorx"
	"github.com/luyb177/silent-sign-backend/common/respx"
	"github.com/luyb177/silent-sign-backend/internal/logic/moment"
	"github.com/luyb177/silent-sign-backend/internal/svc"
	"github.com/luyb177/silent-sign-backend/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// DeleteMomentHandler 删除动态
func DeleteMomentHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.DeleteMomentReq
		if err := httpx.Parse(r, &req); err != nil {
			respx.ErrorCtx(r.Context(), w, errorx.WrapBadRequest("请求参数解析失败", err))
			return
		}

		l := moment.NewDeleteMomentLogic(r.Context(), svcCtx)
		resp, err := l.DeleteMoment(&req)
		if err != nil {
			respx.ErrorCtx(r.Context(), w, err)
			return
		}
		respx.OkCtx(r.Context(), w, resp)
	}
}
