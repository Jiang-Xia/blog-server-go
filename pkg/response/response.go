// Package response 统一 HTTP JSON 响应体，HTTP 状态码恒为 200，业务语义由 code 表达。
package response

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/cloudwego/hertz/pkg/app"
)

// Body 与 Nest TransformInterceptor 对齐：code/bizCode=200 表示成功。
type Body struct {
	Code    int         `json:"code"`
	BizCode int         `json:"bizCode"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Success 返回 code=200、bizCode=200 的成功响应。
func Success(ctx context.Context, c *app.RequestContext, data interface{}) {
	c.JSON(http.StatusOK, Body{Code: 200, BizCode: 200, Message: "success", Data: data})
}

// SuccessWithMessage 返回成功响应并自定义 message（如 refresh token）。
func SuccessWithMessage(ctx context.Context, c *app.RequestContext, message string, data interface{}) {
	c.JSON(http.StatusOK, Body{Code: 200, BizCode: 200, Message: message, Data: data})
}

// Error 返回业务错误响应；bizCode 与 code 一致。
func Error(ctx context.Context, c *app.RequestContext, ec errcode.ErrCode, args ...any) {
	code := ec.Code()
	c.JSON(http.StatusOK, Body{Code: code, BizCode: code, Message: ec.Message(args...)})
}

// FromError 将 errcode 包装的错误映射为响应。
func FromError(ctx context.Context, c *app.RequestContext, err error) {
	if ec, ok := err.(errcode.ErrCode); ok {
		Error(ctx, c, ec)
		return
	}
	Error(ctx, c, errcode.InternalError)
}

// WriteHTTPError 向 net/http 写入 JSON 错误体（gateway 代理上游不可用时使用）。
func WriteHTTPError(w http.ResponseWriter, httpStatus, bizCode int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(httpStatus)
	_ = json.NewEncoder(w).Encode(Body{Code: bizCode, BizCode: bizCode, Message: message})
}
