// Package handler 用户域应用服务，聚合 auth/profile/oauth 供 UserHandler 调用。
package handler

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/pkg/crypto"
	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/services/user/internal/user/auth"
	"github.com/Jiang-Xia/blog-server-go/services/user/internal/user/profile"
	"github.com/Jiang-Xia/blog-server-go/services/user/internal/user/repo"
)

// UserAppAdapter 实现 UserAppService，委托 auth 与 profile。
type UserAppAdapter struct {
	cfg     *config.Config
	auth    *auth.AuthService
	profile *profile.Service
	github  *auth.GitHubOAuth
}

// NewUserAppAdapter 构造 UserAppAdapter。
func NewUserAppAdapter(cfg *config.Config, authSvc *auth.AuthService, profileSvc *profile.Service, github *auth.GitHubOAuth) *UserAppAdapter {
	return &UserAppAdapter{cfg: cfg, auth: authSvc, profile: profileSvc, github: github}
}

func (a *UserAppAdapter) Register(ctx context.Context, req *RegisterReq) (interface{}, error) {
	out, err := a.auth.Register(ctx, auth.RegisterInput{
		Username: req.Username, Nickname: req.Nickname,
		Password: req.Password, PasswordRepeat: req.PasswordRepeat, Avatar: req.Avatar,
	}, false)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (a *UserAppAdapter) Login(ctx context.Context, req *LoginReq, ip string) (interface{}, error) {
	res, err := a.auth.Login(ctx, auth.LoginInput{Username: req.Username, Password: req.Password}, ip)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"info": res}, nil
}

func (a *UserAppAdapter) Refresh(ctx context.Context, token string) (interface{}, string, error) {
	m, err := a.auth.Refresh(ctx, token)
	if err != nil {
		return nil, "", err
	}
	msg, _ := m["message"].(string)
	delete(m, "message")
	return m, msg, nil
}

func (a *UserAppAdapter) ResolveProfileQueryUID(ctx context.Context, operatorUID, requestedUID int) (int, error) {
	var ptr *int
	if requestedUID > 0 {
		ptr = &requestedUID
	}
	return a.profile.ResolveProfileUID(ctx, operatorUID, ptr)
}

func (a *UserAppAdapter) GetUserRolePrivilegeInfo(ctx context.Context, uid int) (interface{}, error) {
	return a.profile.GetUserRolePrivilegeInfo(ctx, uid)
}

func (a *UserAppAdapter) FindAll(ctx context.Context, req *UserListReq) (interface{}, error) {
	return a.profile.ListUsers(ctx, repo.UserListQuery{
		Page: req.Page, PageSize: req.PageSize, Username: req.Username, Nickname: req.Nickname,
	})
}

func (a *UserAppAdapter) UpdateField(ctx context.Context, req map[string]interface{}) (interface{}, error) {
	return a.profile.UpdateField(ctx, req)
}

func (a *UserAppAdapter) AssertSuperAdmin(ctx context.Context, uid int) error {
	return a.profile.AssertSuperAdmin(ctx, uid)
}

func (a *UserAppAdapter) AssertSelfOrSuperAdmin(ctx context.Context, operatorUID, targetUID int) error {
	return a.profile.AssertSelfOrSuperAdmin(ctx, operatorUID, targetUID)
}

func (a *UserAppAdapter) UpdatePassword(ctx context.Context, uid int, req *PasswordReq) (interface{}, error) {
	old := crypto.RSADecrypt(req.PasswordOld, a.cfg.Crypto.RSAPrivateKeyOrDefault())
	newPwd := crypto.RSADecrypt(req.Password, a.cfg.Crypto.RSAPrivateKeyOrDefault())
	if err := a.auth.UpdatePassword(ctx, auth.UpdatePasswordInput{
		ID: uid, PasswordOld: old, Password: newPwd, PasswordRepeat: req.PasswordRepeat,
	}); err != nil {
		return nil, err
	}
	return true, nil
}

func (a *UserAppAdapter) ResetPassword(ctx context.Context, req *ResetPasswordReq) (interface{}, error) {
	return a.auth.ResetPassword(ctx, auth.ResetPasswordInput{Username: req.Username, Nickname: req.Nickname})
}

func (a *UserAppAdapter) DeleteByID(ctx context.Context, id int) (interface{}, error) {
	if err := a.profile.DeleteUser(ctx, id); err != nil {
		return nil, err
	}
	return true, nil
}

func (a *UserAppAdapter) SendEmailCode(ctx context.Context, req *SendEmailCodeReq) (interface{}, error) {
	return a.auth.SendEmailCode(ctx, auth.SendEmailCodeInput{Email: req.Email, Type: req.Type})
}

func (a *UserAppAdapter) EmailRegister(ctx context.Context, req *EmailRegisterReq) (interface{}, error) {
	return a.auth.EmailRegister(ctx, auth.EmailRegisterInput{
		Email: req.Email, Nickname: req.Nickname, Password: req.Password,
		PasswordRepeat: req.Password, VerificationCode: req.Code,
	})
}

func (a *UserAppAdapter) EmailLogin(ctx context.Context, req *EmailLoginReq) (interface{}, error) {
	res, err := a.auth.EmailLogin(ctx, auth.EmailLoginInput{Email: req.Email, VerificationCode: req.Code})
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"info": res}, nil
}

func (a *UserAppAdapter) GithubAuthRedirectURL(ctx context.Context) (string, error) {
	if a.cfg.OAuth.GithubClientID == "" {
		return "", errcode.WithMessage(errcode.InternalError, "GitHub OAuth 未配置")
	}
	return a.github.AuthCodeURL("state"), nil
}

func (a *UserAppAdapter) GithubAuthCallback(ctx context.Context, _ interface{}) (string, error) {
	return "", errcode.InvalidParam
}

// GithubAuthCallbackWithCode 用 authorization code 完成 OAuth 并重定向前端。
func (a *UserAppAdapter) GithubAuthCallbackWithCode(ctx context.Context, code string) (string, error) {
	res, err := a.github.HandleCallback(ctx, code)
	if err != nil {
		return "", err
	}
	blogHome := a.cfg.App.BlogHome
	if blogHome == "" {
		blogHome = "http://localhost:5050"
	}
	return fmt.Sprintf("%s/login?ticket=%s", blogHome, url.QueryEscape(res.Ticket)), nil
}

func (a *UserAppAdapter) ExchangeOAuthTicket(ctx context.Context, ticket string) (interface{}, error) {
	tokens, err := a.auth.ExchangeOAuthTicket(ctx, ticket)
	if err != nil {
		return nil, err
	}
	return tokens, nil
}

func (a *UserAppAdapter) WechatMiniProgramLogin(ctx context.Context, req *WechatMiniLoginReq) (interface{}, error) {
	res, err := a.auth.WechatMiniProgramLogin(ctx, auth.WechatMiniProgramLoginInput{Code: req.Code})
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"info": res}, nil
}

func (a *UserAppAdapter) AdminCreateUser(ctx context.Context, body map[string]interface{}) (interface{}, error) {
	in := profile.AdminCreateInput{
		Username: strField(body, "username"),
		Nickname: strField(body, "nickname"),
		Password: strField(body, "password"),
		Intro:    strField(body, "intro"),
		Avatar:   strField(body, "avatar"),
		DeptID:   intField(body, "deptId"),
	}
	if ids, ok := body["roleIds"].([]interface{}); ok {
		for _, v := range ids {
			if id, ok := toInt(v); ok {
				in.RoleIDs = append(in.RoleIDs, id)
			}
		}
	}
	return a.profile.AdminCreate(ctx, in)
}

func (a *UserAppAdapter) AdminUpdateUser(ctx context.Context, id int, body map[string]interface{}) (interface{}, error) {
	in := profile.AdminUpdateInput{
		Nickname: strField(body, "nickname"),
		Intro:    strField(body, "intro"),
		Avatar:   strField(body, "avatar"),
		DeptID:   intField(body, "deptId"),
	}
	if ids, ok := body["roleIds"].([]interface{}); ok {
		in.RoleIDs = make([]int, 0, len(ids))
		for _, v := range ids {
			if rid, ok := toInt(v); ok {
				in.RoleIDs = append(in.RoleIDs, rid)
			}
		}
	}
	return a.profile.AdminUpdate(ctx, id, in)
}

func strField(m map[string]interface{}, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	return fmt.Sprint(v)
}

func intField(m map[string]interface{}, key string) int {
	v, ok := m[key]
	if !ok {
		return 0
	}
	i, _ := toInt(v)
	return i
}

func toInt(v interface{}) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case int64:
		return int(n), true
	case float64:
		return int(n), true
	case string:
		i, err := strconv.Atoi(n)
		return i, err == nil
	default:
		return 0, false
	}
}
