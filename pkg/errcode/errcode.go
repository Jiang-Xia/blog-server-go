// Package errcode 定义业务错误码，与 Nest HTTP 状态及 bizCode 语义对齐。
package errcode

import "fmt"

// ErrCode 业务错误接口。
type ErrCode interface {
	error
	Code() int
	Message(args ...any) string
}

type bizError struct {
	code int
	msg  string
}

func (e bizError) Error() string             { return e.msg }
func (e bizError) Code() int                 { return e.code }
func (e bizError) Message(_ ...any) string   { return e.msg }

func newCode(code int, msg string) ErrCode {
	return bizError{code: code, msg: msg}
}

// 通用错误码（HTTP 语义映射为业务 code）。
var (
	Success        = newCode(0, "success")
	InvalidParam   = newCode(400, "参数错误")
	Unauthorized   = newCode(401, "未授权")
	Forbidden      = newCode(403, "无权限")
	NotFound       = newCode(404, "资源不存在")
	Conflict       = newCode(409, "资源冲突")
	InternalError  = newCode(500, "服务器内部错误")
	TokenExpired   = newCode(401, "登录已过期")
	CaptchaRefresh = newCode(10001, "验证码需要刷新")
)

// WithMessage 基于模板错误码生成带格式化文案的错误。
func WithMessage(base ErrCode, format string, args ...any) ErrCode {
	return bizError{code: base.Code(), msg: fmt.Sprintf(format, args...)}
}
