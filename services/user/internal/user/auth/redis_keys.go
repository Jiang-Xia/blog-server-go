package auth

import "strings"

const (
	loginFailWindowSec = 600
	loginFailMaxCount  = 5
	loginLockSec       = 15 * 60
	oauthTicketTTL     = 60
)

func loginFailKey(username, ip string) string {
	return "auth:login:fail:" + redisSafe(username) + ":" + redisSafe(ip)
}

func loginLockKey(username, ip string) string {
	return "auth:login:lock:" + redisSafe(username) + ":" + redisSafe(ip)
}

func refreshBlacklistKey(tokenHash string) string {
	return "auth:refresh:blacklist:" + tokenHash
}

func oauthTicketKey(ticket string) string {
	return "auth:oauth:ticket:" + ticket
}

func redisSafe(v string) string {
	if v == "" {
		v = "unknown"
	}
	var b strings.Builder
	for _, r := range v {
		if (r >= '0' && r <= '9') || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '.' || r == '_' || r == '-' {
			b.WriteRune(r)
		} else {
			b.WriteRune('_')
		}
	}
	return b.String()
}
