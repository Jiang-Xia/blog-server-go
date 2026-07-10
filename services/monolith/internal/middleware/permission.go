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
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/auth"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/repo"
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

// PermissionDeps Permission 中间件依赖，由 wire 注入。
type PermissionDeps struct {
	Cfg      *config.Config
	Redis    *redisutil.Store
	RoleRepo *repo.RoleRepo
	JWT      *auth.JWTService
	Log      *zap.Logger
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
	regexPattern := pattern
	regexPattern = regexp.MustCompile(`:(\w+)`).ReplaceAllString(regexPattern, `([^/]+)`)
	regexPattern = strings.ReplaceAll(regexPattern, "**", ".*")
	regexPattern = strings.ReplaceAll(regexPattern, "*", `[^/]*`)
	var sb strings.Builder
	for _, ch := range regexPattern {
		if strings.ContainsRune(`.+?^${}|[]\\`, ch) {
			sb.WriteRune('\\')
		}
		sb.WriteRune(ch)
	}
	re, err := regexp.Compile("^" + sb.String() + "$")
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

		roles, err := deps.RoleRepo.ListRolesByUserID(ctx, uid)
		if err != nil {
			response.Error(ctx, c, errcode.InternalError)
			c.Abort()
			return
		}
		roleIDs := make([]int, 0, len(roles))
		isSuperAdmin := false
		roleBriefs := make([]ctxutil.RoleBrief, 0, len(roles))
		for _, r := range roles {
			roleIDs = append(roleIDs, r.ID)
			roleBriefs = append(roleBriefs, ctxutil.RoleBrief{ID: r.ID, RoleName: r.RoleName})
			if r.ID == superAdminRoleID {
				isSuperAdmin = true
			}
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
	raw, err := deps.Redis.Get(ctx, redisKeyPublicPaths)
	if err != nil {
		return nil, err
	}
	if raw != "" {
		var paths []publicPathEntry
		if err := json.Unmarshal([]byte(raw), &paths); err == nil {
			return paths, nil
		}
	}
	privs, err := deps.RoleRepo.LoadAllPrivileges(ctx)
	if err != nil {
		return nil, err
	}
	paths := make([]publicPathEntry, 0)
	for _, p := range privs {
		if p.IsPublic != 1 {
			continue
		}
		paths = append(paths, publicPathEntry{
			Pattern: p.PathPattern,
			Methods: []string{p.HTTPMethod},
		})
	}
	if b, err := json.Marshal(paths); err == nil {
		_ = deps.Redis.Set(ctx, redisKeyPublicPaths, string(b), cacheAPITTLSeconds)
	}
	return paths, nil
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
	privs, err := deps.RoleRepo.LoadAllPrivileges(ctx)
	if err != nil {
		return nil, err
	}
	cache := make(map[string]apiPermission, len(privs))
	out := make([]apiPermission, 0, len(privs))
	for _, p := range privs {
		perm := apiPermission{
			ID:               p.ID,
			PathPattern:      p.PathPattern,
			HTTPMethod:       p.HTTPMethod,
			PrivilegeCode:    p.PrivilegeCode,
			IsPublic:         p.IsPublic == 1,
			RequireOwnership: p.RequireOwnership == 1,
		}
		key := fmt.Sprintf("%s:%s:%s", perm.HTTPMethod, perm.PathPattern, perm.PrivilegeCode)
		cache[key] = perm
		out = append(out, perm)
	}
	if b, err := json.Marshal(cache); err == nil {
		_ = deps.Redis.Set(ctx, redisKeyAPIMappings, string(b), cacheAPITTLSeconds)
	}
	return out, nil
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
	codes, err := deps.RoleRepo.PrivilegeCodesByRoleID(ctx, roleID)
	if err != nil {
		return nil, err
	}
	if b, err := json.Marshal(codes); err == nil {
		_ = deps.Redis.Set(ctx, key, string(b), cacheRolePermTTLSeconds)
	}
	return codes, nil
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
		{Pattern: "/rag/status", Methods: []string{"GET"}},
	}
}
