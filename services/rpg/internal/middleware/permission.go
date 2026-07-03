// Package middleware RBAC 权限中间件，对齐 Nest PermissionGuard。
package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/pkg/ctxutil"
	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/pkg/redisutil"
	"github.com/Jiang-Xia/blog-server-go/pkg/response"
	"github.com/Jiang-Xia/blog-server-go/services/rpg/internal/auth"
	"github.com/cloudwego/hertz/pkg/app"
	"go.uber.org/zap"
)

const (
	redisKeyPublicPaths     = "public_api_paths"
	redisKeyAPIMappings     = "api_permission_mappings"
	redisKeyRolePermissions = "role_permissions:"
	cacheAPITTLSeconds      = 5 * 60
	cacheRolePermTTLSeconds = 2 * 60
	superAdminRoleID        = 1
)

// PermissionDeps Permission 中间件依赖，由 wire 注入（blog 无 user ent，角色来自 JWT + Redis 缓存）。
type PermissionDeps struct {
	Cfg   *config.Config
	Redis *redisutil.Store
	JWT   *auth.JWTService
	Log   *zap.Logger
}

// apiPermission 与 Nest ApiPermission 及 Redis 缓存结构一致。
type apiPermission struct {
	ID               int    `json:"id"`
	PathPattern      string `json:"pathPattern"`
	HTTPMethod       string `json:"httpMethod"`
	PrivilegeCode    string `json:"privilegeCode"`
	IsPublic         bool   `json:"isPublic"`
	RequireOwnership bool   `json:"requireOwnership"`
}

type publicPathEntry struct {
	Pattern string   `json:"pattern"`
	Methods []string `json:"methods"`
}

type pathMatcher struct {
	mu    sync.RWMutex
	cache map[string]*regexp.Regexp
}

func newPathMatcher() *pathMatcher {
	return &pathMatcher{cache: make(map[string]*regexp.Regexp)}
}

func (m *pathMatcher) match(actualPath, pattern string) bool {
	cleanPath := strings.SplitN(actualPath, "?", 2)[0]
	m.mu.RLock()
	re, ok := m.cache[pattern]
	m.mu.RUnlock()
	if !ok {
		re = compilePathPattern(pattern)
		m.mu.Lock()
		m.cache[pattern] = re
		m.mu.Unlock()
	}
	return re.MatchString(cleanPath)
}

// compilePathPattern 将 Nest pathPattern（:id、*、**）转为正则。
func compilePathPattern(pattern string) *regexp.Regexp {
	const (
		phGlobStarStar = "\x01"
		phGlobStar     = "\x02"
		phParam        = "\x03"
	)
	regexPattern := strings.ReplaceAll(pattern, "**", phGlobStarStar)
	regexPattern = strings.ReplaceAll(regexPattern, "*", phGlobStar)
	regexPattern = regexp.MustCompile(`:(\w+)`).ReplaceAllString(regexPattern, phParam)

	var sb strings.Builder
	for _, ch := range regexPattern {
		switch ch {
		case '\x01', '\x02', '\x03':
			sb.WriteRune(ch)
		default:
			if strings.ContainsRune(`.+?^${}|[]\\`, ch) {
				sb.WriteRune('\\')
			}
			sb.WriteRune(ch)
		}
	}
	out := sb.String()
	out = strings.ReplaceAll(out, phGlobStarStar, ".*")
	out = strings.ReplaceAll(out, phGlobStar, `[^/]*`)
	out = strings.ReplaceAll(out, phParam, `([^/]+)`)
	re, err := regexp.Compile("^" + out + "$")
	if err != nil {
		return regexp.MustCompile("^" + regexp.QuoteMeta(pattern) + "$")
	}
	return re
}

// Permission 全局 RBAC 校验：剥离 api_prefix、公共路径放行、超管绕过、权限码匹配。
func Permission(deps PermissionDeps) app.HandlerFunc {
	matcher := newPathMatcher()
	return func(ctx context.Context, c *app.RequestContext) {
		method := string(c.Method())
		path := string(c.Path())
		url := strings.TrimPrefix(path, deps.Cfg.App.APIPrefix)

		if deps.Log != nil {
			deps.Log.Debug("permission check", zap.String("method", method), zap.String("url", url))
		}

		if isPublic, err := checkPublicPath(ctx, deps, matcher, method, url); err != nil {
			if deps.Log != nil {
				deps.Log.Error("check public path failed", zap.Error(err))
			}
		} else if isPublic {
			c.Next(ctx)
			return
		}

		uid := ctxutil.UserID(ctx)
		if uid == 0 {
			uid = uidFromBearer(c, deps.JWT)
		}
		if uid == 0 {
			response.Error(ctx, c, errcode.WithMessage(errcode.Unauthorized, "用户未登录"))
			c.Abort()
			return
		}

		roleIDs, roleBriefs, isSuperAdmin := rolesFromJWT(c, deps.JWT)
		if len(roleIDs) == 0 {
			response.Error(ctx, c, errcode.WithMessage(errcode.Unauthorized, "用户未登录"))
			c.Abort()
			return
		}
		ctx = ctxutil.WithUserID(ctx, uid)
		ctx = ctxutil.WithRoles(ctx, roleBriefs)

		apiPerm, err := matchAPIPermission(ctx, deps, matcher, method, url)
		if err != nil {
			if deps.Log != nil {
				deps.Log.Error("match api permission failed", zap.Error(err))
			}
			response.Error(ctx, c, errcode.InternalError)
			c.Abort()
			return
		}
		if apiPerm == nil {
			if deps.Log != nil {
				deps.Log.Warn("api permission not configured", zap.String("method", method), zap.String("url", url))
			}
			// 开发环境：权限表/Redis 未同步时仍允许已登录用户访问（与 Plan 02 本地冒烟一致）。
			if deps.Cfg.IsDev() || isSuperAdmin {
				c.Next(ctx)
				return
			}
			response.Error(ctx, c, errcode.WithMessage(errcode.InternalError, "接口权限未配置，请联系管理员"))
			c.Abort()
			return
		}

		if isSuperAdmin {
			c.Next(ctx)
			return
		}

		ok, err := userHasPrivilegeCode(ctx, deps, roleIDs, apiPerm.PrivilegeCode)
		if err != nil {
			response.Error(ctx, c, errcode.InternalError)
			c.Abort()
			return
		}
		if !ok {
			response.Error(ctx, c, errcode.WithMessage(errcode.Forbidden, "权限不足，无法访问该资源"))
			c.Abort()
			return
		}
		c.Next(ctx)
	}
}

func uidFromBearer(c *app.RequestContext, jwtSvc *auth.JWTService) int {
	if jwtSvc == nil {
		return 0
	}
	token := bearerToken(c)
	if token == "" {
		return 0
	}
	claims, err := jwtSvc.Verify(token)
	if err != nil || claims == nil {
		return 0
	}
	return claims.ID
}

func checkPublicPath(ctx context.Context, deps PermissionDeps, matcher *pathMatcher, method, url string) (bool, error) {
	// 开发兜底：核心认证接口与 Nest isPublic=1 对齐，避免空权限库阻塞登录链路。
	for _, p := range defaultPublicAuthPaths() {
		if !matcher.match(url, p.Pattern) {
			continue
		}
		for _, m := range p.Methods {
			if m == "*" || strings.EqualFold(m, method) {
				return true, nil
			}
		}
	}
	paths, err := loadPublicPaths(ctx, deps)
	if err != nil {
		return false, err
	}
	for _, p := range paths {
		if !matcher.match(url, p.Pattern) {
			continue
		}
		for _, m := range p.Methods {
			if m == "*" || strings.EqualFold(m, method) {
				return true, nil
			}
		}
	}
	return false, nil
}

func loadPublicPaths(ctx context.Context, deps PermissionDeps) ([]publicPathEntry, error) {
	var paths []publicPathEntry
	raw, err := deps.Redis.Get(ctx, redisKeyPublicPaths)
	if err != nil {
		return nil, err
	}
	if raw != "" {
		if err := json.Unmarshal([]byte(raw), &paths); err != nil {
			paths = nil
		}
	}
	if deps.Cfg.IsDev() {
		paths = mergePublicPaths(paths, append(defaultBlogPublicPaths(), defaultRPGPublicPaths()...))
	}
	return paths, nil
}

func mergePublicPaths(base, extra []publicPathEntry) []publicPathEntry {
	seen := make(map[string]struct{}, len(base)+len(extra))
	out := make([]publicPathEntry, 0, len(base)+len(extra))
	for _, p := range append(base, extra...) {
		key := p.Pattern + "|" + strings.Join(p.Methods, ",")
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, p)
	}
	return out
}

// defaultRPGPublicPaths 开发环境 RPG 公开路由兜底（与 register.go 无 jwt 路由对齐子集）。
func defaultRPGPublicPaths() []publicPathEntry {
	return []publicPathEntry{
		{Pattern: "/user/public/:uid", Methods: []string{"GET"}},
		{Pattern: "/user/public/:uid/articles", Methods: []string{"GET"}},
		{Pattern: "/user/public/:uid/collects", Methods: []string{"GET"}},
		{Pattern: "/user/public/:uid/likes", Methods: []string{"GET"}},
		{Pattern: "/rpg/public/:uid/status", Methods: []string{"GET"}},
		{Pattern: "/rpg/public/status/batch", Methods: []string{"GET"}},
		{Pattern: "/rpg/leaderboard", Methods: []string{"GET"}},
		{Pattern: "/rpg/level-rewards", Methods: []string{"GET"}},
		{Pattern: "/rpg/quests", Methods: []string{"GET"}},
		{Pattern: "/rpg/lottery/pool", Methods: []string{"GET"}},
		{Pattern: "/rpg/pets/catalog", Methods: []string{"GET"}},
		{Pattern: "/rpg/activities/current", Methods: []string{"GET"}},
		{Pattern: "/rpg/weather-buff", Methods: []string{"GET"}},
		{Pattern: "/rpg/guilds", Methods: []string{"GET"}},
		{Pattern: "/rpg/guild/:id", Methods: []string{"GET"}},
		{Pattern: "/pay/notice", Methods: []string{"POST"}},
	}
}

func matchAPIPermission(ctx context.Context, deps PermissionDeps, matcher *pathMatcher, method, url string) (*apiPermission, error) {
	mappings, err := loadAPIMappings(ctx, deps)
	if err != nil {
		return nil, err
	}
	for _, perm := range mappings {
		if perm.HTTPMethod != "*" && !strings.EqualFold(perm.HTTPMethod, method) {
			continue
		}
		if matcher.match(url, perm.PathPattern) {
			p := perm
			return &p, nil
		}
	}
	return nil, nil
}

func loadAPIMappings(ctx context.Context, deps PermissionDeps) ([]apiPermission, error) {
	raw, err := deps.Redis.Get(ctx, redisKeyAPIMappings)
	if err != nil {
		return nil, err
	}
	if raw != "" {
		var cache map[string]apiPermission
		if err := json.Unmarshal([]byte(raw), &cache); err == nil {
			out := make([]apiPermission, 0, len(cache))
			for _, v := range cache {
				out = append(out, v)
			}
			return out, nil
		}
	}
	return nil, nil
}

func userHasPrivilegeCode(ctx context.Context, deps PermissionDeps, roleIDs []int, code string) (bool, error) {
	for _, roleID := range roleIDs {
		codes, err := rolePrivilegeCodes(ctx, deps, roleID)
		if err != nil {
			if deps.Log != nil {
				deps.Log.Error("load role permissions failed", zap.Int("roleId", roleID), zap.Error(err))
			}
			continue
		}
		for _, c := range codes {
			if c == code {
				return true, nil
			}
		}
	}
	return false, nil
}

func rolePrivilegeCodes(ctx context.Context, deps PermissionDeps, roleID int) ([]string, error) {
	key := redisKeyRolePermissions + fmt.Sprintf("%d", roleID)
	raw, err := deps.Redis.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	if raw != "" {
		var codes []string
		if err := json.Unmarshal([]byte(raw), &codes); err == nil {
			return codes, nil
		}
	}
	return nil, nil
}

func rolesFromJWT(c *app.RequestContext, jwtSvc *auth.JWTService) ([]int, []ctxutil.RoleBrief, bool) {
	if jwtSvc == nil {
		return nil, nil, false
	}
	token := bearerToken(c)
	if token == "" {
		return nil, nil, false
	}
	claims, err := jwtSvc.Verify(token)
	if err != nil || claims == nil {
		return nil, nil, false
	}
	roleIDs := make([]int, 0, len(claims.Role))
	briefs := make([]ctxutil.RoleBrief, 0, len(claims.Role))
	isSuperAdmin := false
	for _, r := range claims.Role {
		roleIDs = append(roleIDs, r.ID)
		briefs = append(briefs, ctxutil.RoleBrief{ID: r.ID, RoleName: r.RoleName})
		if r.ID == superAdminRoleID {
			isSuperAdmin = true
		}
	}
	return roleIDs, briefs, isSuperAdmin
}

// defaultBlogPublicPaths 开发环境 C 端公开路由兜底（与 Nest isPublic=1 对齐子集）。
func defaultBlogPublicPaths() []publicPathEntry {
	base := defaultPublicAuthPaths()
	blog := []publicPathEntry{
		{Pattern: "/article/list", Methods: []string{"POST"}},
		{Pattern: "/article/info", Methods: []string{"GET"}},
		{Pattern: "/article/views", Methods: []string{"POST"}},
		{Pattern: "/article/likes", Methods: []string{"POST"}},
		{Pattern: "/article/archives", Methods: []string{"GET"}},
		{Pattern: "/article/related", Methods: []string{"GET"}},
		{Pattern: "/article/statistics", Methods: []string{"GET"}},
		{Pattern: "/category", Methods: []string{"GET"}},
		{Pattern: "/category/:id", Methods: []string{"GET"}},
		{Pattern: "/tag", Methods: []string{"GET"}},
		{Pattern: "/tag/:id", Methods: []string{"GET"}},
		{Pattern: "/tag/:id/article", Methods: []string{"GET"}},
		{Pattern: "/comment/findAll", Methods: []string{"GET"}},
		{Pattern: "/comment/admin", Methods: []string{"GET"}},
		{Pattern: "/reply/findAll", Methods: []string{"GET"}},
		{Pattern: "/like", Methods: []string{"POST"}},
		{Pattern: "/collect/count", Methods: []string{"GET"}},
		{Pattern: "/msgboard", Methods: []string{"GET", "POST"}},
		{Pattern: "/link", Methods: []string{"GET", "POST"}},
		{Pattern: "/link/:id", Methods: []string{"GET", "PATCH"}},
		{Pattern: "/resources/daily-img", Methods: []string{"GET"}},
		{Pattern: "/resources/weather", Methods: []string{"GET"}},
		{Pattern: "/resources/files", Methods: []string{"GET"}},
		{Pattern: "/resources/register-avatars", Methods: []string{"GET"}},
		{Pattern: "/resources/file/:id", Methods: []string{"GET"}},
		{Pattern: "/resources/upload-media/register-avatar", Methods: []string{"POST"}},
		{Pattern: "/resources/folder", Methods: []string{"POST"}},
	}
	return append(base, blog...)
}

// defaultPublicAuthPaths Plan 02 认证公开路由（与 Nest privilege isPublic 一致）。
func defaultPublicAuthPaths() []publicPathEntry {
	return []publicPathEntry{
		{Pattern: "/user/authCode", Methods: []string{"GET"}},
		{Pattern: "/user/register", Methods: []string{"POST"}},
		{Pattern: "/user/login", Methods: []string{"POST"}},
		{Pattern: "/user/refresh", Methods: []string{"GET"}},
		{Pattern: "/user/email/sendCode", Methods: []string{"POST"}},
		{Pattern: "/user/email/register", Methods: []string{"POST"}},
		{Pattern: "/user/email/login", Methods: []string{"POST"}},
		{Pattern: "/user/auth/github", Methods: []string{"GET"}},
		{Pattern: "/user/auth/github/callback", Methods: []string{"GET"}},
		{Pattern: "/user/auth/ticket/exchange", Methods: []string{"POST"}},
		{Pattern: "/user/auth/wechat/miniprogram", Methods: []string{"POST"}},
		{Pattern: "/captcha", Methods: []string{"GET"}},
		{Pattern: "/captcha/verify", Methods: []string{"POST"}},
		{Pattern: "/pub/stats", Methods: []string{"GET"}},
	}
}
