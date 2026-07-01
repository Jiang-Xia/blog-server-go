// Package handler 用户模块 HTTP 端点，路径对齐 Nest UserController。
package handler

import (
	"context"
	"strconv"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/pkg/ctxutil"
	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/pkg/response"
	"github.com/Jiang-Xia/blog-server-go/services/user/internal/user/captcha"
	"github.com/cloudwego/hertz/pkg/app"
)

// UserAppService 用户域应用服务，由 profile 包实现并经 wire 注入。
type UserAppService interface {
	Register(ctx context.Context, req *RegisterReq) (interface{}, error)
	Login(ctx context.Context, req *LoginReq, ip string) (interface{}, error)
	Refresh(ctx context.Context, token string) (data interface{}, message string, err error)
	ResolveProfileQueryUID(ctx context.Context, operatorUID, requestedUID int) (int, error)
	GetUserRolePrivilegeInfo(ctx context.Context, uid int) (interface{}, error)
	FindAll(ctx context.Context, req *UserListReq) (interface{}, error)
	UpdateField(ctx context.Context, req map[string]interface{}) (interface{}, error)
	AssertSuperAdmin(ctx context.Context, uid int) error
	AssertSelfOrSuperAdmin(ctx context.Context, operatorUID, targetUID int) error
	UpdatePassword(ctx context.Context, uid int, req *PasswordReq) (interface{}, error)
	ResetPassword(ctx context.Context, req *ResetPasswordReq) (interface{}, error)
	DeleteByID(ctx context.Context, id int) (interface{}, error)
	SendEmailCode(ctx context.Context, req *SendEmailCodeReq) (interface{}, error)
	EmailRegister(ctx context.Context, req *EmailRegisterReq) (interface{}, error)
	EmailLogin(ctx context.Context, req *EmailLoginReq) (interface{}, error)
	GithubAuthRedirectURL(ctx context.Context) (string, error)
	GithubAuthCallback(ctx context.Context, profile interface{}) (redirectURL string, err error)
	ExchangeOAuthTicket(ctx context.Context, ticket string) (interface{}, error)
	WechatMiniProgramLogin(ctx context.Context, req *WechatMiniLoginReq) (interface{}, error)
	AdminCreateUser(ctx context.Context, req map[string]interface{}) (interface{}, error)
	AdminUpdateUser(ctx context.Context, id int, req map[string]interface{}) (interface{}, error)
}

// UserHandlerDeps 用户 handler 依赖。
type UserHandlerDeps struct {
	Cfg     *config.Config
	Svc     UserAppService
	Captcha *captcha.Service
}

// UserHandler 用户模块 HTTP handler。
type UserHandler struct {
	cfg     *config.Config
	svc     UserAppService
	captcha *captcha.Service
}

// NewUserHandler 构造 UserHandler。
func NewUserHandler(deps UserHandlerDeps) *UserHandler {
	return &UserHandler{cfg: deps.Cfg, svc: deps.Svc, captcha: deps.Captcha}
}

// RegisterReq 账号注册请求体。
type RegisterReq struct {
	Username       string `json:"username"`
	Nickname       string `json:"nickname"`
	Password       string `json:"password"`
	PasswordRepeat string `json:"passwordRepeat"`
	AuthCode       string `json:"authCode"`
	CaptchaID      string `json:"captchaId"`
	Avatar         string `json:"avatar"`
}

// LoginReq 账号登录请求体。
type LoginReq struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	AuthCode  string `json:"authCode"`
	CaptchaID string `json:"captchaId"`
	Admin     bool   `json:"admin"`
}

// UserListReq 用户列表分页请求。
type UserListReq struct {
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Status   string `json:"status"`
}

// PasswordReq 修改密码请求。
type PasswordReq struct {
	Password       string `json:"password"`
	PasswordRepeat string `json:"passwordRepeat"`
	PasswordOld    string `json:"passwordOld"`
}

// ResetPasswordReq 重置密码请求。
type ResetPasswordReq struct {
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	AuthCode  string `json:"authCode"`
	CaptchaID string `json:"captchaId"`
}

// SendEmailCodeReq 发送邮箱验证码。
type SendEmailCodeReq struct {
	Email string `json:"email"`
	Type  string `json:"type"`
}

// EmailRegisterReq 邮箱注册。
type EmailRegisterReq struct {
	Email    string `json:"email"`
	Code     string `json:"code"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
}

// EmailLoginReq 邮箱验证码登录。
type EmailLoginReq struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

// OAuthTicketExchangeReq OAuth ticket 兑换 token。
type OAuthTicketExchangeReq struct {
	Ticket string `json:"ticket"`
}

// WechatMiniLoginReq 微信小程序登录。
type WechatMiniLoginReq struct {
	Code string `json:"code"`
}

// AuthCode GET /user/authCode — 生成图形验证码（base64），写入 captcha_id Cookie。
func (h *UserHandler) AuthCode(ctx context.Context, c *app.RequestContext) {
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

// Register POST /user/register — 账号注册（需图形验证码）。
func (h *UserHandler) Register(ctx context.Context, c *app.RequestContext) {
	var req RegisterReq
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	captchaID := resolveCaptchaID(c, req.CaptchaID)
	if err := h.captcha.Verify(ctx, captchaID, req.AuthCode); err != nil {
		response.FromError(ctx, c, err)
		return
	}
	out, err := h.svc.Register(ctx, &req)
	if err != nil {
		response.FromError(ctx, c, err)
		return
	}
	response.Success(ctx, c, out)
}

// Login POST /user/login — 账号登录。
func (h *UserHandler) Login(ctx context.Context, c *app.RequestContext) {
	var req LoginReq
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	captchaID := resolveCaptchaID(c, req.CaptchaID)
	if err := h.captcha.Verify(ctx, captchaID, req.AuthCode); err != nil {
		response.FromError(ctx, c, err)
		return
	}
	out, err := h.svc.Login(ctx, &req, clientIP(c))
	if err != nil {
		response.FromError(ctx, c, err)
		return
	}
	response.Success(ctx, c, out)
}

// Refresh GET /user/refresh — 用 refreshToken 刷新 accessToken。
func (h *UserHandler) Refresh(ctx context.Context, c *app.RequestContext) {
	token := string(c.Query("token"))
	if token == "" {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	data, message, err := h.svc.Refresh(ctx, token)
	if err != nil {
		response.FromError(ctx, c, err)
		return
	}
	if message != "" {
		response.SuccessWithMessage(ctx, c, message, data)
		return
	}
	response.Success(ctx, c, data)
}

// Info GET /user/info — 获取用户信息（需 JWT；?id= 仅超管可查他人）。
func (h *UserHandler) Info(ctx context.Context, c *app.RequestContext) {
	tokenUID := ctxutil.UserID(ctx)
	requestedID, _ := strconv.Atoi(string(c.Query("id")))
	targetID, err := h.svc.ResolveProfileQueryUID(ctx, tokenUID, requestedID)
	if err != nil {
		response.FromError(ctx, c, err)
		return
	}
	out, err := h.svc.GetUserRolePrivilegeInfo(ctx, targetID)
	if err != nil {
		response.FromError(ctx, c, err)
		return
	}
	response.Success(ctx, c, out)
}

// List POST /user/list — 分页用户列表。
func (h *UserHandler) List(ctx context.Context, c *app.RequestContext) {
	var req UserListReq
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	out, err := h.svc.FindAll(ctx, &req)
	if err != nil {
		response.FromError(ctx, c, err)
		return
	}
	response.Success(ctx, c, out)
}

// UpdateStatus PATCH /user/status — 锁定/解锁用户（超管）。
func (h *UserHandler) UpdateStatus(ctx context.Context, c *app.RequestContext) {
	tokenUID := ctxutil.UserID(ctx)
	if err := h.svc.AssertSuperAdmin(ctx, tokenUID); err != nil {
		response.FromError(ctx, c, err)
		return
	}
	var body map[string]interface{}
	if err := c.BindAndValidate(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	out, err := h.svc.UpdateField(ctx, body)
	if err != nil {
		response.FromError(ctx, c, err)
		return
	}
	response.Success(ctx, c, out)
}

// Edit PATCH /user/edit — 修改用户资料（本人或超管）。
func (h *UserHandler) Edit(ctx context.Context, c *app.RequestContext) {
	tokenUID := ctxutil.UserID(ctx)
	var body map[string]interface{}
	if err := c.BindAndValidate(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	targetID := intFromMap(body, "id")
	if targetID == 0 {
		response.Error(ctx, c, errcode.WithMessage(errcode.InvalidParam, "缺少用户 id"))
		return
	}
	if err := h.svc.AssertSelfOrSuperAdmin(ctx, tokenUID, targetID); err != nil {
		response.FromError(ctx, c, err)
		return
	}
	out, err := h.svc.UpdateField(ctx, body)
	if err != nil {
		response.FromError(ctx, c, err)
		return
	}
	response.Success(ctx, c, out)
}

// Password PATCH /user/password — 修改密码（当前登录用户）。
func (h *UserHandler) Password(ctx context.Context, c *app.RequestContext) {
	tokenUID := ctxutil.UserID(ctx)
	var req PasswordReq
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	out, err := h.svc.UpdatePassword(ctx, tokenUID, &req)
	if err != nil {
		response.FromError(ctx, c, err)
		return
	}
	response.Success(ctx, c, out)
}

// ResetPassword POST /user/resetPassword — 通过验证码重置密码。
func (h *UserHandler) ResetPassword(ctx context.Context, c *app.RequestContext) {
	var req ResetPasswordReq
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	out, err := h.svc.ResetPassword(ctx, &req)
	if err != nil {
		response.FromError(ctx, c, err)
		return
	}
	response.Success(ctx, c, out)
}

// Delete DELETE /user — 删除用户（超管）。
func (h *UserHandler) Delete(ctx context.Context, c *app.RequestContext) {
	tokenUID := ctxutil.UserID(ctx)
	if err := h.svc.AssertSuperAdmin(ctx, tokenUID); err != nil {
		response.FromError(ctx, c, err)
		return
	}
	id, _ := strconv.Atoi(string(c.Query("id")))
	if id == 0 {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	out, err := h.svc.DeleteByID(ctx, id)
	if err != nil {
		response.FromError(ctx, c, err)
		return
	}
	response.Success(ctx, c, out)
}

// SendEmailCode POST /user/email/sendCode — 发送邮箱验证码。
func (h *UserHandler) SendEmailCode(ctx context.Context, c *app.RequestContext) {
	var req SendEmailCodeReq
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	out, err := h.svc.SendEmailCode(ctx, &req)
	if err != nil {
		response.FromError(ctx, c, err)
		return
	}
	response.Success(ctx, c, out)
}

// EmailRegister POST /user/email/register — 邮箱注册。
func (h *UserHandler) EmailRegister(ctx context.Context, c *app.RequestContext) {
	var req EmailRegisterReq
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	out, err := h.svc.EmailRegister(ctx, &req)
	if err != nil {
		response.FromError(ctx, c, err)
		return
	}
	response.Success(ctx, c, out)
}

// EmailLogin POST /user/email/login — 邮箱验证码登录。
func (h *UserHandler) EmailLogin(ctx context.Context, c *app.RequestContext) {
	var req EmailLoginReq
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	out, err := h.svc.EmailLogin(ctx, &req)
	if err != nil {
		response.FromError(ctx, c, err)
		return
	}
	response.Success(ctx, c, out)
}

// GithubAuth GET /user/auth/github — 跳转 GitHub OAuth 授权页。
func (h *UserHandler) GithubAuth(ctx context.Context, c *app.RequestContext) {
	url, err := h.svc.GithubAuthRedirectURL(ctx)
	if err != nil {
		response.FromError(ctx, c, err)
		return
	}
	c.Redirect(302, []byte(url))
}

// GithubCallback GET /user/auth/github/callback — GitHub OAuth 回调，重定向前端带 ticket。
func (h *UserHandler) GithubCallback(ctx context.Context, c *app.RequestContext) {
	code := string(c.Query("code"))
	adapter, ok := h.svc.(*UserAppAdapter)
	if !ok {
		response.Error(ctx, c, errcode.InternalError)
		return
	}
	redirectURL, err := adapter.GithubAuthCallbackWithCode(ctx, code)
	if err != nil {
		response.FromError(ctx, c, err)
		return
	}
	c.Redirect(302, []byte(redirectURL))
}

// ExchangeOAuthTicket POST /user/auth/ticket/exchange — OAuth ticket 换 token。
func (h *UserHandler) ExchangeOAuthTicket(ctx context.Context, c *app.RequestContext) {
	var req OAuthTicketExchangeReq
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	out, err := h.svc.ExchangeOAuthTicket(ctx, req.Ticket)
	if err != nil {
		response.FromError(ctx, c, err)
		return
	}
	response.Success(ctx, c, map[string]interface{}{"info": out})
}

// WechatMiniProgramLogin POST /user/auth/wechat/miniprogram — 微信小程序登录。
func (h *UserHandler) WechatMiniProgramLogin(ctx context.Context, c *app.RequestContext) {
	var req WechatMiniLoginReq
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	out, err := h.svc.WechatMiniProgramLogin(ctx, &req)
	if err != nil {
		response.FromError(ctx, c, err)
		return
	}
	response.Success(ctx, c, out)
}

// AdminCreate POST /user/admin/create — 超管创建用户。
func (h *UserHandler) AdminCreate(ctx context.Context, c *app.RequestContext) {
	var body map[string]interface{}
	if err := c.BindAndValidate(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	out, err := h.svc.AdminCreateUser(ctx, body)
	if err != nil {
		response.FromError(ctx, c, err)
		return
	}
	response.Success(ctx, c, out)
}

// AdminUpdate POST /user/admin/update/:id — 超管更新用户。
func (h *UserHandler) AdminUpdate(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(c.Param("id"))
	if id == 0 {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	var body map[string]interface{}
	if err := c.BindAndValidate(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	out, err := h.svc.AdminUpdateUser(ctx, id, body)
	if err != nil {
		response.FromError(ctx, c, err)
		return
	}
	response.Success(ctx, c, out)
}

func intFromMap(m map[string]interface{}, key string) int {
	v, ok := m[key]
	if !ok || v == nil {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case int64:
		return int(n)
	case string:
		i, _ := strconv.Atoi(n)
		return i
	default:
		return 0
	}
}
