// wechat 微信小程序 code 登录与自动建号。
package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/repo"
)

// WechatMiniProgramLoginInput 微信小程序 code 登录参数。
type WechatMiniProgramLoginInput struct {
	Code string
}

// WechatSession 微信 jscode2session 响应。
type WechatSession struct {
	OpenID     string `json:"openid"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
	SessionKey string `json:"session_key"`
}

// WechatMiniProgramLogin 微信小程序 code 登录：jscode2session → 查找/创建用户 → 签发 token。
func (s *AuthService) WechatMiniProgramLogin(ctx context.Context, in WechatMiniProgramLoginInput) (*LoginResult, error) {
	appID := s.cfg.Wechat.AppID
	secret := s.cfg.Wechat.Secret
	if appID == "" || secret == "" {
		return nil, errcode.WithMessage(errcode.InternalError, "微信小程序登录未配置 pay_wechatAppId / pay_wechatSecret")
	}
	session, err := s.wechatCode2Session(ctx, appID, secret, in.Code)
	if err != nil {
		return nil, err
	}
	if session.OpenID == "" {
		msg := session.ErrMsg
		if msg == "" {
			msg = "微信登录失败"
		}
		return nil, errcode.WithMessage(errcode.InvalidParam, "%s", msg)
	}
	u, err := s.users.FindByWechatOpenID(ctx, session.OpenID)
	if err != nil {
		if !ent.IsNotFound(err) {
			return nil, err
		}
		u, err = s.createWechatUser(ctx, session.OpenID)
		if err != nil {
			return nil, err
		}
	}
	return s.LoginAfterOAuth(ctx, u)
}

func (s *AuthService) wechatCode2Session(ctx context.Context, appID, secret, code string) (*WechatSession, error) {
	q := url.Values{}
	q.Set("appid", appID)
	q.Set("secret", secret)
	q.Set("js_code", code)
	q.Set("grant_type", "authorization_code")
	apiURL := "https://api.weixin.qq.com/sns/jscode2session?" + q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var session WechatSession
	if err := json.Unmarshal(body, &session); err != nil {
		return nil, err
	}
	if session.ErrCode != 0 && session.OpenID == "" {
		return &session, nil
	}
	return &session, nil
}

func (s *AuthService) createWechatUser(ctx context.Context, openID string) (*ent.User, error) {
	suffix := openID
	if len(suffix) > 8 {
		suffix = suffix[len(suffix)-8:]
	}
	username := fmt.Sprintf("wx_%s_%s", suffix, fmt.Sprintf("%x", time.Now().UnixNano())[:6])
	in := repo.CreateUserInput{
		WechatID: openID,
		Nickname: "微信用户" + suffix,
		Username: username,
		Password: RandomPassword(),
	}
	return s.CreateUserWithDefaultRole(ctx, in)
}
