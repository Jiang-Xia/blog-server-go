// Package response 统一 HTTP JSON 响应体，HTTP 状态码恒为 200，业务语义由 code 表达。
package response

import (
	"context"
	"net/http"

	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/cloudwego/hertz/pkg/app"
)

// Body 与 Nest TransformInterceptor 对齐的成功/失败包装（Go 侧成功 code=0）。
type Body struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Success 返回 code=0 的成功响应。
func Success(ctx context.Context, c *app.RequestContext, data interface{}) {
	c.JSON(http.StatusOK, Body{Code: 0, Message: "success", Data: data})
}

// Error 返回业务错误响应。
func Error(ctx context.Context, c *app.RequestContext, ec errcode.ErrCode, args ...any) {
	c.JSON(http.StatusOK, Body{Code: ec.Code(), Message: ec.Message(args...)})
}

// FromError 将 errcode 包装的错误映射为响应。
func FromError(ctx context.Context, c *app.RequestContext, err error) {
	if ec, ok := err.(errcode.ErrCode); ok {
		Error(ctx, c, ec)
		return
	}
	Error(ctx, c, errcode.InternalError)
}
