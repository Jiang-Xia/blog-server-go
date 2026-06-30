// Package handler RBAC 后台管理 HTTP 端点，路径对齐 Nest Role/Dept/Privilege/MenuController。
package handler

import (
	"context"
	"strconv"
	"strings"

	"github.com/Jiang-Xia/blog-server-go/pkg/ctxutil"
	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/pkg/response"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/admin"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/auth"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/repo"
	"github.com/cloudwego/hertz/pkg/app"
)

// AdminHandler RBAC 后台 handler。
type AdminHandler struct {
	svc *admin.Service
	jwt *auth.JWTService
}

// NewAdminHandler 构造 AdminHandler。
func NewAdminHandler(svc *admin.Service, jwt *auth.JWTService) *AdminHandler {
	return &AdminHandler{svc: svc, jwt: jwt}
}

// uid 读取当前用户 ID；对齐 Nest getUid(Authorization)，Permission 中间件 ctx 未透传时回退解析 Bearer。
func (h *AdminHandler) uid(ctx context.Context, c *app.RequestContext) int {
	if uid := ctxutil.UserID(ctx); uid != 0 {
		return uid
	}
	if h.jwt == nil {
		return 0
	}
	authz := strings.TrimSpace(string(c.GetHeader("Authorization")))
	if authz == "" {
		return 0
	}
	const prefix = "Bearer "
	token := authz
	if strings.HasPrefix(authz, prefix) {
		token = strings.TrimSpace(authz[len(prefix):])
	}
	claims, err := h.jwt.Verify(token)
	if err != nil || claims == nil {
		return 0
	}
	return claims.ID
}

func handleAdminResult(ctx context.Context, c *app.RequestContext, data interface{}, err error) {
	if err != nil {
		response.FromError(ctx, c, err)
		return
	}
	response.Success(ctx, c, data)
}

// --- Role ---

func (h *AdminHandler) RoleMenuPrivilegeTree(ctx context.Context, c *app.RequestContext) {
	data, err := h.svc.MenuPrivilegeTree(ctx)
	handleAdminResult(ctx, c, data, err)
}

func (h *AdminHandler) RoleCreate(ctx context.Context, c *app.RequestContext) {
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	data, err := h.svc.CreateRole(ctx, strField(body, "roleName"), strField(body, "roleDesc"), ifaceSlice(body["privileges"]), ifaceSlice(body["menus"]))
	handleAdminResult(ctx, c, data, err)
}

func (h *AdminHandler) RoleList(ctx context.Context, c *app.RequestContext) {
	page, _ := strconv.Atoi(string(c.Query("page")))
	pageSize, _ := strconv.Atoi(string(c.Query("pageSize")))
	data, err := h.svc.ListRoles(ctx, page, pageSize, string(c.Query("roleName")))
	handleAdminResult(ctx, c, data, err)
}

func (h *AdminHandler) RoleGetDataScope(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(c.Param("id"))
	data, err := h.svc.GetRoleDataScopes(ctx, id)
	handleAdminResult(ctx, c, data, err)
}

func (h *AdminHandler) RoleUpdateDataScope(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(c.Param("id"))
	var body struct {
		DataScopes []map[string]interface{} `json:"dataScopes"`
	}
	if err := c.Bind(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	scopes, err := admin.ParseDataScopesFromBody(body.DataScopes)
	if err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	data, err := h.svc.UpdateRoleDataScopes(ctx, id, scopes)
	handleAdminResult(ctx, c, data, err)
}

func (h *AdminHandler) RoleGet(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(c.Param("id"))
	data, err := h.svc.GetRole(ctx, id)
	handleAdminResult(ctx, c, data, err)
}

func (h *AdminHandler) RoleUpdate(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(c.Param("id"))
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	data, err := h.svc.UpdateRole(ctx, id, ifaceSlice(body["privileges"]), ifaceSlice(body["menus"]))
	handleAdminResult(ctx, c, data, err)
}

func (h *AdminHandler) RoleDelete(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(c.Param("id"))
	data, err := h.svc.DeleteRole(ctx, id)
	handleAdminResult(ctx, c, data, err)
}

// --- Dept ---

func (h *AdminHandler) DeptCreate(ctx context.Context, c *app.RequestContext) {
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	d := repo.DeptEntity{
		DeptName: strField(body, "deptName"),
		DeptCode: strField(body, "deptCode"),
		ParentID: intField(body, "parentId"),
		OrderNum: intField(body, "orderNum"),
		Status:   intFieldDefault(body, "status", 1),
	}
	if v, ok := body["leaderId"].(string); ok {
		d.LeaderID = &v
	}
	if v, ok := body["leaderName"].(string); ok {
		d.LeaderName = &v
	}
	if v, ok := body["remark"].(string); ok {
		d.Remark = &v
	}
	data, err := h.svc.CreateDept(ctx, d)
	handleAdminResult(ctx, c, data, err)
}

func (h *AdminHandler) DeptList(ctx context.Context, c *app.RequestContext) {
	page, _ := strconv.Atoi(string(c.Query("page")))
	pageSize, _ := strconv.Atoi(string(c.Query("pageSize")))
	var parentID *int
	if v := string(c.Query("parentId")); v != "" {
		n, _ := strconv.Atoi(v)
		parentID = &n
	}
	var status *int
	if v := string(c.Query("status")); v != "" {
		n, _ := strconv.Atoi(v)
		status = &n
	}
	data, err := h.svc.ListDepts(ctx, h.uid(ctx, c), page, pageSize, string(c.Query("deptName")), parentID, status)
	handleAdminResult(ctx, c, data, err)
}

func (h *AdminHandler) DeptTree(ctx context.Context, c *app.RequestContext) {
	var status *int
	if v := string(c.Query("status")); v != "" {
		n, _ := strconv.Atoi(v)
		status = &n
	}
	data, err := h.svc.DeptTree(ctx, h.uid(ctx, c), string(c.Query("id")), string(c.Query("deptName")), status)
	handleAdminResult(ctx, c, data, err)
}

func (h *AdminHandler) DeptGet(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(c.Param("id"))
	data, err := h.svc.GetDept(ctx, id)
	handleAdminResult(ctx, c, data, err)
}

func (h *AdminHandler) DeptUpdate(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(c.Param("id"))
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	data, err := h.svc.UpdateDept(ctx, id, body)
	handleAdminResult(ctx, c, data, err)
}

func (h *AdminHandler) DeptDelete(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(c.Param("id"))
	data, err := h.svc.DeleteDept(ctx, id)
	handleAdminResult(ctx, c, data, err)
}

// --- Privilege ---

func (h *AdminHandler) PrivilegeCreate(ctx context.Context, c *app.RequestContext) {
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	p := privilegeFromBody(body)
	data, err := h.svc.CreatePrivilege(ctx, p)
	handleAdminResult(ctx, c, data, err)
}

func (h *AdminHandler) PrivilegeList(ctx context.Context, c *app.RequestContext) {
	page, _ := strconv.Atoi(string(c.Query("page")))
	pageSize, _ := strconv.Atoi(string(c.Query("pageSize")))
	filters := repo.PrivilegeListFilters{
		PrivilegeName: string(c.Query("privilegeName")),
		PathPattern:   string(c.Query("pathPattern")),
		HTTPMethod:    string(c.Query("httpMethod")),
	}
	if v := string(c.Query("isPublic")); v != "" {
		b := v == "true" || v == "1"
		filters.IsPublic = &b
	}
	if v := string(c.Query("isVisible")); v != "" {
		b := v == "true" || v == "1"
		filters.IsVisible = &b
	}
	data, err := h.svc.ListPrivileges(ctx, page, pageSize, filters)
	handleAdminResult(ctx, c, data, err)
}

func (h *AdminHandler) PrivilegeGet(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(c.Param("id"))
	data, err := h.svc.GetPrivilege(ctx, id)
	handleAdminResult(ctx, c, data, err)
}

func (h *AdminHandler) PrivilegeUpdate(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(c.Param("id"))
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	data, err := h.svc.UpdatePrivilege(ctx, id, body)
	handleAdminResult(ctx, c, data, err)
}

func (h *AdminHandler) PrivilegeDelete(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(c.Param("id"))
	data, err := h.svc.DeletePrivilege(ctx, id)
	handleAdminResult(ctx, c, data, err)
}

// --- Menu ---

func (h *AdminHandler) MenuList(ctx context.Context, c *app.RequestContext) {
	data, err := h.svc.UserMenuTree(ctx, h.uid(ctx, c))
	handleAdminResult(ctx, c, data, err)
}

func (h *AdminHandler) MenuCreate(ctx context.Context, c *app.RequestContext) {
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	data, err := h.svc.CreateMenu(ctx, menuFromBody(body))
	handleAdminResult(ctx, c, data, err)
}

func (h *AdminHandler) MenuUpdate(ctx context.Context, c *app.RequestContext) {
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	id := strField(body, "id")
	data, err := h.svc.UpdateMenu(ctx, id, body)
	handleAdminResult(ctx, c, data, err)
}

func (h *AdminHandler) MenuDetail(ctx context.Context, c *app.RequestContext) {
	data, err := h.svc.GetMenu(ctx, string(c.Query("id")))
	handleAdminResult(ctx, c, data, err)
}

func (h *AdminHandler) MenuDelete(ctx context.Context, c *app.RequestContext) {
	data, err := h.svc.DeleteMenu(ctx, string(c.Query("id")))
	handleAdminResult(ctx, c, data, err)
}

func intFieldDefault(m map[string]interface{}, key string, def int) int {
	if m[key] == nil {
		return def
	}
	return intField(m, key)
}

func ifaceSlice(v interface{}) []interface{} {
	arr, ok := v.([]interface{})
	if !ok {
		return nil
	}
	return arr
}

func privilegeFromBody(body map[string]interface{}) repo.PrivilegeFullEntity {
	p := repo.PrivilegeFullEntity{
		PrivilegeName: strField(body, "privilegeName"),
		PrivilegeCode: strField(body, "privilegeCode"),
		PrivilegePage: strField(body, "privilegePage"),
		PathPattern:   strField(body, "pathPattern"),
		HTTPMethod:    strField(body, "httpMethod"),
	}
	if v, ok := body["isVisible"].(bool); ok {
		p.IsVisible = v
	} else {
		p.IsVisible = true
	}
	if v, ok := body["isPublic"].(bool); ok {
		p.IsPublic = v
	}
	if v, ok := body["requireOwnership"].(bool); ok {
		p.RequireOwnership = v
	}
	if v, ok := body["description"].(string); ok {
		p.Description = &v
	}
	return p
}

func menuFromBody(body map[string]interface{}) repo.MenuEntity {
	m := repo.MenuEntity{
		ID:       strField(body, "id"),
		PID:      strField(body, "pid"),
		Path:     strField(body, "path"),
		Name:     strField(body, "name"),
		Icon:     strField(body, "icon"),
		Locale:   strField(body, "locale"),
		FilePath: strField(body, "filePath"),
		Order:    intFieldDefault(body, "order", 1),
	}
	if v, ok := body["menuCnName"].(string); ok {
		m.MenuCnName = &v
	}
	if v, ok := body["requiresAuth"].(bool); ok {
		m.RequiresAuth = v
	} else {
		m.RequiresAuth = true
	}
	if m.PID == "" {
		m.PID = "0"
	}
	return m
}
