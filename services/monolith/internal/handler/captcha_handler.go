// Package handler 验证码 HTTP 端点，对齐 Nest CaptchaController。
package handler

import (
	"context"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/pkg/response"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/captcha"
	"github.com/cloudwego/hertz/pkg/app"
)

// CaptchaHandlerDeps 验证码 handler 依赖。
type CaptchaHandlerDeps struct {
	Cfg     *config.Config
	Captcha *captcha.Service
}

// CaptchaHandler 图形验证码接口。
type CaptchaHandler struct {
	cfg     *config.Config
	captcha *captcha.Service
}

// NewCaptchaHandler 构造 CaptchaHandler。
func NewCaptchaHandler(deps CaptchaHandlerDeps) *CaptchaHandler {
	return &CaptchaHandler{cfg: deps.Cfg, captcha: deps.Captcha}
}

// Get GET /captcha — 生成图形验证码，id 写入 captcha_id Cookie。
func (h *CaptchaHandler) Get(ctx context.Context, c *app.RequestContext) {
	ip := clientIP(c)
	browserID := ensureBrowserIDCookie(c, h.cfg)
	identity := captchaIdentity(ip, browserID)
	if err := h.captcha.AssertRateLimit(ctx, identity); err != nil {
		response.FromError(ctx, c, err)
		return
	}
	result, err := h.captcha.Create(ctx)
	if err != nil {
		response.FromError(ctx, c, err)
		return
	}
	setCaptchaIDCookie(c, h.cfg, result.ID)
	response.Success(ctx, c, captchaPayload(result))
}

type captchaVerifyReq struct {
	ID     string `json:"id"`
	Answer string `json:"answer"`
}

// Verify POST /captcha/verify — 校验验证码答案。
func (h *CaptchaHandler) Verify(ctx context.Context, c *app.RequestContext) {
	var req captchaVerifyReq
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	if err := h.captcha.Verify(ctx, req.ID, req.Answer); err != nil {
		response.FromError(ctx, c, err)
		return
	}
	response.Success(ctx, c, map[string]bool{"ok": true})
}
