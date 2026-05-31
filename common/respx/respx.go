package respx

import (
	"context"
	"net/http"

	"github.com/luyb177/silent-sign-backend/common/errorx"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// Body is the unified API response envelope: {code, msg, data}.
type Body struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type EmptyData struct{}

// OkCtx writes a successful response with code=0, msg="ok".
func OkCtx(ctx context.Context, w http.ResponseWriter, data interface{}) {
	httpx.OkJsonCtx(ctx, w, &Body{Code: 0, Msg: "ok", Data: data})
}

// ErrorCtx writes an error response. If err is an *AppError, its Code and Msg
// are extracted; otherwise it falls back to CodeInternalError.
// The HTTP status code is derived from the business code.
func ErrorCtx(ctx context.Context, w http.ResponseWriter, err error) {
	ae := errorx.AsAppError(err)
	httpx.WriteJson(w, ae.HTTPStatus(), &Body{Code: ae.Code, Msg: ae.Msg, Data: &EmptyData{}})
}
