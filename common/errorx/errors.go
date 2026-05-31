package errorx

import (
	"errors"
	"fmt"
	"net/http"
)

const (
	CodeOK = 0

	CodeBadRequest      = 400
	CodeUnauthorized    = 401
	CodeForbidden       = 403
	CodeNotFound        = 404
	CodeConflict        = 409
	CodeUnprocessable   = 422
	CodeTooManyRequests = 429

	// 10 - 通用系统错误
	CodeInternalError  = 100001 // 系统内部错误
	CodeNotImplemented = 100002 // 功能暂未实现
	CodeUnavailable    = 100003 // 服务暂不可用

	// 11 - 数据库错误
	CodeDatabaseQueryFailed  = 110101 // 数据查询失败
	CodeDatabaseInsertFailed = 110102 // 数据插入失败
	CodeDatabaseUpdateFailed = 110103 // 数据更新失败
	CodeDatabaseDeleteFailed = 110104 // 数据删除失败
	CodeDatabaseTxFailed     = 110105 // 数据事务失败

	// 12 - Redis错误
	CodeRedisGetFailed = 120101 // Redis读取失败
	CodeRedisSetFailed = 120102 // Redis写入失败

	// 18 - 权限认证错误
	CodeTokenInvalid = 180101 // Token无效
	CodeTokenExpired = 180102 // Token过期
)

var codeToHTTPStatus = map[int]int{
	CodeOK: http.StatusOK,

	CodeBadRequest:      http.StatusBadRequest,
	CodeUnauthorized:    http.StatusUnauthorized,
	CodeForbidden:       http.StatusForbidden,
	CodeNotFound:        http.StatusNotFound,
	CodeConflict:        http.StatusConflict,
	CodeUnprocessable:   http.StatusUnprocessableEntity,
	CodeTooManyRequests: http.StatusTooManyRequests,

	CodeInternalError:        http.StatusInternalServerError,
	CodeNotImplemented:       http.StatusNotImplemented,
	CodeUnavailable:          http.StatusServiceUnavailable,
	CodeDatabaseQueryFailed:  http.StatusInternalServerError,
	CodeDatabaseInsertFailed: http.StatusInternalServerError,
	CodeDatabaseUpdateFailed: http.StatusInternalServerError,
	CodeDatabaseDeleteFailed: http.StatusInternalServerError,
	CodeDatabaseTxFailed:     http.StatusInternalServerError,
	CodeRedisGetFailed:       http.StatusInternalServerError,
	CodeRedisSetFailed:       http.StatusInternalServerError,
	CodeTokenInvalid:         http.StatusUnauthorized,
	CodeTokenExpired:         http.StatusUnauthorized,
}

type AppError struct {
	Code  int
	Msg   string
	cause error
}

func (e *AppError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("code=%d msg=%s cause=%v", e.Code, e.Msg, e.cause)
	}
	return fmt.Sprintf("code=%d msg=%s", e.Code, e.Msg)
}

func (e *AppError) Unwrap() error {
	return e.cause
}

func (e *AppError) HTTPStatus() int {
	if s, ok := codeToHTTPStatus[e.Code]; ok {
		return s
	}
	return http.StatusInternalServerError
}

func New(code int, msg string) *AppError {
	return &AppError{Code: code, Msg: msg}
}

func Wrap(code int, msg string, cause error) *AppError {
	return &AppError{Code: code, Msg: msg, cause: cause}
}

var (
	ErrBadRequest     = New(CodeBadRequest, "请求参数错误")
	ErrUnauthorized   = New(CodeUnauthorized, "请先登录")
	ErrForbidden      = New(CodeForbidden, "无权限访问")
	ErrNotFound       = New(CodeNotFound, "资源不存在")
	ErrInternalServer = New(CodeInternalError, "服务器内部错误")
	ErrUnavailable    = New(CodeUnavailable, "服务暂不可用")
	ErrNotImplemented = New(CodeNotImplemented, "功能暂未实现")

	ErrTokenInvalid = New(CodeTokenInvalid, "令牌无效")
	ErrTokenExpired = New(CodeTokenExpired, "令牌已过期")

	ErrDatabaseQuery  = New(CodeDatabaseQueryFailed, "数据库查询失败")
	ErrDatabaseInsert = New(CodeDatabaseInsertFailed, "数据库新增失败")
	ErrDatabaseUpdate = New(CodeDatabaseUpdateFailed, "数据库更新失败")
	ErrDatabaseDelete = New(CodeDatabaseDeleteFailed, "数据库删除失败")
	ErrDatabaseTx     = New(CodeDatabaseTxFailed, "数据库事务失败")

	ErrRedisGet = New(CodeRedisGetFailed, "缓存读取失败")
	ErrRedisSet = New(CodeRedisSetFailed, "缓存写入失败")
)

// ──────────────────────────────────────────────
// Helpers
// ──────────────────────────────────────────────

func AsAppError(err error) *AppError {
	if err == nil {
		return nil
	}
	if ae, ok := errors.AsType[*AppError](err); ok {
		return ae
	}
	return ErrInternalServer
}

func WrapBadRequest(msg string, err error) *AppError {
	return Wrap(CodeBadRequest, msg, err)
}

func WrapInternal(msg string, err error) *AppError {
	return Wrap(CodeInternalError, msg, err)
}

func WrapDBQuery(msg string, err error) *AppError {
	return Wrap(CodeDatabaseQueryFailed, msg, err)
}

func WrapDBInsert(msg string, err error) *AppError {
	return Wrap(CodeDatabaseInsertFailed, msg, err)
}

func WrapDBUpdate(msg string, err error) *AppError {
	return Wrap(CodeDatabaseUpdateFailed, msg, err)
}

func WrapDBDelete(msg string, err error) *AppError {
	return Wrap(CodeDatabaseDeleteFailed, msg, err)
}

func WrapDBTx(msg string, err error) *AppError {
	return Wrap(CodeDatabaseTxFailed, msg, err)
}

func WrapRedisGet(msg string, err error) *AppError {
	return Wrap(CodeRedisGetFailed, msg, err)
}

func WrapRedisSet(msg string, err error) *AppError {
	return Wrap(CodeRedisSetFailed, msg, err)
}
