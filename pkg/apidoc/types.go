package apidoc

import "github.com/Jiang-Xia/blog-server-go/pkg/response"

// SuccessBody swag @Success 引用：Nest 对齐的统一成功响应。
type SuccessBody struct {
	Code    int         `json:"code" example:"200"`
	BizCode int         `json:"bizCode" example:"200"`
	Message string      `json:"message" example:"success"`
	Data    interface{} `json:"data"`
}

// ErrorBody swag @Failure 引用：业务错误响应。
type ErrorBody struct {
	Code    int    `json:"code" example:"40001"`
	BizCode int    `json:"bizCode" example:"40001"`
	Message string `json:"message" example:"参数错误"`
}

// EnsureResponseRef 编译期引用 pkg/response，避免 apidoc 与 response 包循环依赖。
var _ = response.Body{}
