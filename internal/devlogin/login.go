// Package devlogin 提供本地开发登录（captcha + RSA），供 dev_login 脚本与 test 复用。
package devlogin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/pkg/crypto"
	"github.com/redis/rueidis"
)

const (
	DefaultUsername = "18888888888"
	DefaultPassword = "super"
)

// Tokens 与 Nest 登录响应三 token 对齐。
type Tokens struct {
	Token        string `json:"token"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type apiResp struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type loginData struct {
	Info struct {
		Token        string `json:"token"`
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
	} `json:"info"`
}

type authCodeData struct {
	CaptchaID string `json:"captchaId"`
}

// Login 完整登录：清 captcha 频控 → authCode → Redis 读答案 → RSA 密码 → POST login。
func Login(ctx context.Context, cfg *config.Config, apiBase, username, password string) (*Tokens, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Jar: jar, Timeout: 15 * time.Second}

	if err := clearCaptchaRate(ctx, cfg); err != nil {
		return nil, fmt.Errorf("clear captcha rate: %w", err)
	}

	captchaID, captchaAnswer, err := fetchCaptcha(ctx, client, apiBase)
	if err != nil {
		return nil, err
	}
	if captchaAnswer == "" {
		return nil, fmt.Errorf("无法从 Redis 读取 captcha:%s，请确认 Redis db=%d 已启动", captchaID, cfg.Redis.DB)
	}

	encrypted, err := crypto.RSAEncrypt(password, config.RSAPublicKeyOrDefault())
	if err != nil {
		return nil, fmt.Errorf("rsa encrypt password: %w", err)
	}

	body, _ := json.Marshal(map[string]string{
		"username":  username,
		"password":  encrypted,
		"authCode":  captchaAnswer,
		"captchaId": captchaID,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiBase+"/user/login", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("login request: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	var envelope apiResp
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return nil, fmt.Errorf("parse login response: %w", err)
	}
	if envelope.Code != 200 {
		return nil, fmt.Errorf("login failed code=%d message=%s", envelope.Code, envelope.Message)
	}

	var data loginData
	if err := json.Unmarshal(envelope.Data, &data); err != nil {
		return nil, fmt.Errorf("parse login data: %w", err)
	}
	if data.Info.AccessToken == "" && data.Info.Token == "" {
		return nil, fmt.Errorf("login ok but token empty")
	}
	return &Tokens{
		Token:        firstNonEmpty(data.Info.Token, data.Info.AccessToken),
		AccessToken:  firstNonEmpty(data.Info.AccessToken, data.Info.Token),
		RefreshToken: data.Info.RefreshToken,
	}, nil
}

func fetchCaptcha(ctx context.Context, client *http.Client, apiBase string) (id, answer string, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiBase+"/user/authCode", nil)
	if err != nil {
		return "", "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("authCode: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	var envelope apiResp
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return "", "", err
	}
	if envelope.Code != 200 {
		return "", "", fmt.Errorf("authCode failed code=%d message=%s", envelope.Code, envelope.Message)
	}
	var data authCodeData
	if err := json.Unmarshal(envelope.Data, &data); err != nil {
		return "", "", err
	}
	if data.CaptchaID == "" {
		return "", "", fmt.Errorf("authCode missing captchaId")
	}

	answer, err = readCaptchaAnswer(ctx, data.CaptchaID)
	return data.CaptchaID, answer, err
}

func readCaptchaAnswer(ctx context.Context, captchaID string) (string, error) {
	cfg, err := config.MustLoad("")
	if err != nil {
		return "", err
	}
	client, err := rueidis.NewClient(rueidis.ClientOption{
		InitAddress:  []string{cfg.Redis.Addr},
		SelectDB:     cfg.Redis.DB,
		DisableCache: true,
	})
	if err != nil {
		return "", err
	}
	defer client.Close()

	resp := client.Do(ctx, client.B().Get().Key("captcha:"+captchaID).Build())
	if rueidis.IsRedisNil(resp.Error()) {
		return "", nil
	}
	return resp.ToString()
}

func clearCaptchaRate(ctx context.Context, cfg *config.Config) error {
	client, err := rueidis.NewClient(rueidis.ClientOption{
		InitAddress:  []string{cfg.Redis.Addr},
		SelectDB:     cfg.Redis.DB,
		DisableCache: true,
	})
	if err != nil {
		return err
	}
	defer client.Close()

	keys := []string{
		"captcha:rate:ip_127.0.0.1",
		"captcha:rate:ip___1",
		"captcha:rate:unknown",
	}
	for _, key := range keys {
		_ = client.Do(ctx, client.B().Del().Key(key).Build()).Error()
	}
	return nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
