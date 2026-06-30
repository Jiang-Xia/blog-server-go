// Package handler Cookie 与客户端身份辅助，对齐 Nest user/captcha controller。
package handler

import (
	"strings"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/google/uuid"
)

const (
	cookieCaptchaID = "captcha_id"
	cookieBrowserID = "browser_id"
	captchaCookieSec = 120
	browserCookieSec = 365 * 24 * 60 * 60
)

func setCaptchaIDCookie(c *app.RequestContext, cfg *config.Config, id string) {
	c.SetCookie(cookieCaptchaID, id, captchaCookieSec, "/", "", protocol.CookieSameSiteLaxMode, !cfg.IsDev(), true)
}

func ensureBrowserIDCookie(c *app.RequestContext, cfg *config.Config) string {
	if v := cookieValue(c, cookieBrowserID); v != "" {
		return v
	}
	id := strings.ReplaceAll(uuid.NewString(), "-", "")
	c.SetCookie(cookieBrowserID, id, browserCookieSec, "/", "", protocol.CookieSameSiteLaxMode, !cfg.IsDev(), true)
	return id
}

func cookieValue(c *app.RequestContext, name string) string {
	return string(c.Cookie(name))
}

func resolveCaptchaID(c *app.RequestContext, bodyID string) string {
	if bodyID != "" {
		return bodyID
	}
	return cookieValue(c, cookieCaptchaID)
}

func captchaIdentity(ip, browserID string) string {
	if ip != "" && ip != "unknown" {
		return "ip:" + ip
	}
	return "bid:" + browserID
}

func clientIP(c *app.RequestContext) string {
	if xff := string(c.GetHeader("X-Forwarded-For")); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	if xri := string(c.GetHeader("X-Real-IP")); xri != "" {
		return strings.TrimSpace(xri)
	}
	ip := c.ClientIP()
	if ip == "" {
		return "unknown"
	}
	return ip
}
