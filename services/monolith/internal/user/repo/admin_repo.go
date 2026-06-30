// Package repo RBAC 后台管理表 CRUD（直查 MySQL，与 Nest TypeORM 表结构对齐）。
package repo

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	_ "github.com/go-sql-driver/mysql"
)

const dataScopeResourceArticle = "article"

// AdminRepo 角色/部门/菜单/权限/数据权限 CRUD。
type AdminRepo struct {
	db     *sql.DB
	prefix string
}

// NewAdminRepo 构造 AdminRepo。
func NewAdminRepo(cfg *config.Config) (*AdminRepo, error) {
	db, err := sql.Open("mysql", cfg.MySQL.FormatDSN())
	if err != nil {
		return nil, fmt.Errorf("open mysql for admin repo: %w", err)
	}
	return &AdminRepo{db: db, prefix: cfg.MySQL.TablePrefixOrDefault()}, nil
}

func (r *AdminRepo) deptTable() string            { return r.prefix + "dept" }
func (r *AdminRepo) menuTable() string            { return r.prefix + "menu" }
func (r *AdminRepo) roleMenusTable() string       { return r.prefix + "role_menus_menu" }
func (r *AdminRepo) roleDataScopeTable() string   { return r.prefix + "role_data_scope" }

func formatNestTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

func scanNullableTime(ns sql.NullTime) string {
	if ns.Valid {
		return formatNestTime(ns.Time)
	}
	return ""
}

// DeptEntity 部门记录。
type DeptEntity struct {
	ID         int     `json:"id"`
	DeptName   string  `json:"deptName"`
	DeptCode   string  `json:"deptCode"`
	ParentID   int     `json:"parentId"`
	LeaderID   *string `json:"leaderId,omitempty"`
	LeaderName *string `json:"leaderName,omitempty"`
	OrderNum   int     `json:"orderNum"`
	Status     int     `json:"status"`
	Remark     *string `json:"remark,omitempty"`
	CreateTime string  `json:"createTime"`
	UpdateTime string  `json:"updateTime"`
	Children   []DeptEntity `json:"children,omitempty"`
}

// MenuEntity 菜单记录（管理 CRUD 用，含顶层字段）。
type MenuEntity struct {
	ID           string  `json:"id"`
	PID          string  `json:"pid"`
	Path         string  `json:"path"`
	Name         string  `json:"name"`
	MenuCnName   *string `json:"menuCnName,omitempty"`
	Order        int     `json:"order"`
	Icon         string  `json:"icon"`
	Locale       string  `json:"locale"`
	RequiresAuth bool    `json:"requiresAuth"`
	FilePath     string  `json:"filePath"`
}

// MenuTreeNode 动态菜单树节点（order/icon 等收敛到 meta）。
type MenuTreeNode struct {
	ID         string         `json:"id"`
	PID        string         `json:"pid"`
	Path       string         `json:"path"`
	Name       string         `json:"name"`
	MenuCnName *string        `json:"menuCnName,omitempty"`
	FilePath   string         `json:"filePath,omitempty"`
	Meta       MenuMeta       `json:"meta"`
	Children   []MenuTreeNode `json:"children"`
}

// MenuMeta 菜单 meta 包装，对齐 Nest MenuService.findAll。
type MenuMeta struct {
	Order        int    `json:"order"`
	Icon         string `json:"icon"`
	Locale       string `json:"locale"`
	RequiresAuth bool   `json:"requiresAuth"`
}

// PrivilegeFullEntity 权限完整字段（含 isVisible 布尔化前的原始值）。
type PrivilegeFullEntity struct {
	ID               int     `json:"id"`
	PrivilegeName    string  `json:"privilegeName"`
	PrivilegeCode    string  `json:"privilegeCode"`
	PrivilegePage    string  `json:"privilegePage"`
	IsVisible        bool    `json:"isVisible"`
	PathPattern      string  `json:"pathPattern"`
	HTTPMethod       string  `json:"httpMethod"`
	IsPublic         bool    `json:"isPublic"`
	RequireOwnership bool    `json:"requireOwnership"`
	Description      *string `json:"description,omitempty"`
	CreateTime       string  `json:"createTime"`
	UpdateTime       string  `json:"updateTime"`
	PrivilegePageName string `json:"privilegePageName,omitempty"`
}

// RoleDetailEntity 角色详情（权限/菜单为 ID 数组）。
type RoleDetailEntity struct {
	ID          int                    `json:"id"`
	RoleName    string                 `json:"roleName"`
	RoleDesc    string                 `json:"roleDesc"`
	CreateTime  string                 `json:"createTime"`
	UpdateTime  string                 `json:"updateTime"`
	Privileges  []string               `json:"privileges"`
	Menus       []string               `json:"menus"`
	DataScopes  []RoleDataScopeEntity  `json:"dataScopes"`
}

// RoleListItemEntity 角色列表项（含数据权限摘要）。
type RoleListItemEntity struct {
	ID               int                   `json:"id"`
	RoleName         string                `json:"roleName"`
	RoleDesc         string                `json:"roleDesc"`
	CreateTime       string                `json:"createTime"`
	UpdateTime       string                `json:"updateTime"`
	DataScopes       []RoleDataScopeEntity `json:"dataScopes"`
	ArticleDataScope string                `json:"articleDataScope"`
}

// RoleDataScopeEntity 角色数据权限。
type RoleDataScopeEntity struct {
	ID           int    `json:"id"`
	RoleID       int    `json:"roleId"`
	ResourceType string `json:"resourceType"`
	ScopeType    string `json:"scopeType"`
	DeptIDs      []int  `json:"deptIds"`
	CreateTime   string `json:"createTime,omitempty"`
	UpdateTime   string `json:"updateTime,omitempty"`
}

// MenuPrivilegeTreeNode 角色配置页菜单+权限树。
type MenuPrivilegeTreeNode struct {
	ID                  interface{}             `json:"id"`
	PID                 string                  `json:"pid,omitempty"`
	Path                string                  `json:"path,omitempty"`
	Name                string                  `json:"name,omitempty"`
	MenuCnName          *string                 `json:"menuCnName,omitempty"`
	Order               int                     `json:"order,omitempty"`
	Icon                string                  `json:"icon,omitempty"`
	Locale              string                  `json:"locale,omitempty"`
	FilePath            string                  `json:"filePath,omitempty"`
	Children            []MenuPrivilegeTreeNode `json:"children"`
	Label               string                  `json:"label"`
	Value               interface{}             `json:"value"`
	Type                string                  `json:"type"`
	Privileges          []PrivilegeEntity       `json:"privileges,omitempty"`
	CheckedPrivilegeIDs []string                `json:"checkedPrivilegeIds,omitempty"`
	PrivilegeID         int                     `json:"privilegeId,omitempty"`
	PrivilegeName       string                  `json:"privilegeName,omitempty"`
	PrivilegeCode       string                  `json:"privilegeCode,omitempty"`
	PrivilegePage       string                  `json:"privilegePage,omitempty"`
	IsVisible           bool                    `json:"isVisible,omitempty"`
}

// --- Role CRUD ---

// CreateRole 创建角色并绑定权限/菜单。
func (r *AdminRepo) CreateRole(ctx context.Context, roleName, roleDesc string, privilegeIDs []int, menuIDs []string) (*RoleListItemEntity, error) {
	res, err := r.db.ExecContext(ctx,
		fmt.Sprintf(`INSERT INTO %s (roleName, roleDesc, createTime, updateTime, isDelete, version) VALUES (?, ?, NOW(), NOW(), 0, 1)`, r.roleTable()),
		roleName, roleDesc,
	)
	if err != nil {
		return nil, err
	}
	id64, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	roleID := int(id64)
	if err := r.replaceRolePrivileges(ctx, roleID, privilegeIDs); err != nil {
		return nil, err
	}
	if err := r.replaceRoleMenus(ctx, roleID, menuIDs); err != nil {
		return nil, err
	}
	return r.loadRoleListItem(ctx, roleID)
}

// ListRoles 分页查询角色。
func (r *AdminRepo) ListRoles(ctx context.Context, page, pageSize int, roleName string) ([]RoleListItemEntity, NestPagination, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	where := "WHERE isDelete = 0"
	args := []any{}
	if roleName != "" {
		where += " AND roleName LIKE ?"
		args = append(args, "%"+roleName+"%")
	}
	var total int
	if err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s %s", r.roleTable(), where), args...).Scan(&total); err != nil {
		return nil, NestPagination{}, err
	}
	q := fmt.Sprintf(`SELECT id, roleName, roleDesc, createTime, updateTime FROM %s %s ORDER BY createTime ASC LIMIT ? OFFSET ?`, r.roleTable(), where)
	args = append(args, pageSize, (page-1)*pageSize)
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, NestPagination{}, err
	}
	defer rows.Close()
	list := make([]RoleListItemEntity, 0)
	for rows.Next() {
		item, err := scanRoleListRow(rows)
		if err != nil {
			return nil, NestPagination{}, err
		}
		scopes, _ := r.ListDataScopesByRoleID(ctx, item.ID)
		item.DataScopes = scopes
		for _, s := range scopes {
			if s.ResourceType == dataScopeResourceArticle {
				item.ArticleDataScope = s.ScopeType
				break
			}
		}
		list = append(list, item)
	}
	return list, CalcNestPagination(total, pageSize, page), rows.Err()
}

// GetRoleDetail 查询角色详情（权限/菜单 ID 数组）。
func (r *AdminRepo) GetRoleDetail(ctx context.Context, id int) (*RoleDetailEntity, error) {
	q := fmt.Sprintf(`SELECT id, roleName, roleDesc, createTime, updateTime FROM %s WHERE id = ? AND isDelete = 0`, r.roleTable())
	row := r.db.QueryRowContext(ctx, q, id)
	detail, err := scanRoleDetailRow(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	privIDs, err := r.listPrivilegeIDsByRoleID(ctx, id)
	if err != nil {
		return nil, err
	}
	detail.Privileges = intSliceToStringSlice(privIDs)
	menuIDs, err := r.listMenuIDsByRoleID(ctx, id)
	if err != nil {
		return nil, err
	}
	detail.Menus = menuIDs
	scopes, err := r.ListDataScopesByRoleID(ctx, id)
	if err != nil {
		return nil, err
	}
	detail.DataScopes = scopes
	return detail, nil
}

// UpdateRole 更新角色权限/菜单绑定。
func (r *AdminRepo) UpdateRole(ctx context.Context, id int, privilegeIDs []int, menuIDs []string) (*RoleListItemEntity, error) {
	exists, err := r.roleExists(ctx, id)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}
	if err := r.replaceRolePrivileges(ctx, id, privilegeIDs); err != nil {
		return nil, err
	}
	if err := r.replaceRoleMenus(ctx, id, menuIDs); err != nil {
		return nil, err
	}
	_, err = r.db.ExecContext(ctx, fmt.Sprintf(`UPDATE %s SET updateTime = NOW() WHERE id = ?`, r.roleTable()), id)
	if err != nil {
		return nil, err
	}
	return r.loadRoleListItem(ctx, id)
}

// DeleteRole 删除角色。
func (r *AdminRepo) DeleteRole(ctx context.Context, id int) (bool, error) {
	exists, err := r.roleExists(ctx, id)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, nil
	}
	_, err = r.db.ExecContext(ctx, fmt.Sprintf(`DELETE FROM %s WHERE id = ?`, r.roleTable()), id)
	return err == nil, err
}

// GetMenuPrivilegeTree 构建菜单+权限树（对齐 Nest RoleService.getMenuPrivilegeTree）。
func (r *AdminRepo) GetMenuPrivilegeTree(ctx context.Context) ([]MenuPrivilegeTreeNode, error) {
	menus, err := r.listAllMenus(ctx)
	if err != nil {
		return nil, err
	}
	privs, err := r.loadAllPrivilegesWithVisible(ctx)
	if err != nil {
		return nil, err
	}
	menuMap := make(map[string]*MenuPrivilegeTreeNode, len(menus))
	for _, m := range menus {
		label := m.Name
		if m.MenuCnName != nil && *m.MenuCnName != "" {
			label = *m.MenuCnName
		}
		node := &MenuPrivilegeTreeNode{
			ID: m.ID, PID: m.PID, Path: m.Path, Name: m.Name, MenuCnName: m.MenuCnName,
			Order: m.Order, Icon: m.Icon, Locale: m.Locale, FilePath: m.FilePath,
			Children: []MenuPrivilegeTreeNode{}, Label: label, Value: m.ID, Type: "menu",
			Privileges: []PrivilegeEntity{}, CheckedPrivilegeIDs: []string{},
		}
		menuMap[m.ID] = node
	}
	for _, m := range menus {
		current := menuMap[m.ID]
		for _, p := range privs {
			if p.PrivilegePage != m.ID {
				continue
			}
			current.Privileges = append(current.Privileges, p)
			current.Children = append(current.Children, MenuPrivilegeTreeNode{
				ID: p.ID, Label: p.PrivilegeName, Value: fmt.Sprintf("%d", p.ID), Type: "privilege",
				PrivilegeID: p.ID, PrivilegeName: p.PrivilegeName, PrivilegeCode: p.PrivilegeCode,
				PrivilegePage: p.PrivilegePage, IsVisible: p.IsVisible == 1,
				Children: []MenuPrivilegeTreeNode{},
			})
		}
	}
	roots := make([]MenuPrivilegeTreeNode, 0)
	for _, m := range menus {
		current := menuMap[m.ID]
		if m.PID == "0" || m.PID == "" {
			roots = append(roots, *current)
		} else if parent := menuMap[m.PID]; parent != nil {
			parent.Children = append(parent.Children, *current)
		}
	}
	return roots, nil
}

func (r *AdminRepo) roleTable() string { return r.prefix + "role" }

func (r *AdminRepo) loadAllPrivilegesWithVisible(ctx context.Context) ([]PrivilegeEntity, error) {
	q := fmt.Sprintf(`SELECT id, privilegeName, privilegeCode, privilegePage, pathPattern, httpMethod, isPublic, requireOwnership, description, isVisible FROM %s`, r.privilegeTable())
	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]PrivilegeEntity, 0)
	for rows.Next() {
		var p PrivilegeEntity
		var desc sql.NullString
		var isVisible int
		if err := rows.Scan(&p.ID, &p.PrivilegeName, &p.PrivilegeCode, &p.PrivilegePage, &p.PathPattern, &p.HTTPMethod, &p.IsPublic, &p.RequireOwnership, &desc, &isVisible); err != nil {
			return nil, err
		}
		if desc.Valid {
			p.Description = &desc.String
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *AdminRepo) privilegeTable() string { return r.prefix + "privilege" }

func (r *AdminRepo) roleExists(ctx context.Context, id int) (bool, error) {
	var n int
	err := r.db.QueryRowContext(ctx, fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE id = ? AND isDelete = 0`, r.roleTable()), id).Scan(&n)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *AdminRepo) loadRoleListItem(ctx context.Context, id int) (*RoleListItemEntity, error) {
	q := fmt.Sprintf(`SELECT id, roleName, roleDesc, createTime, updateTime FROM %s WHERE id = ?`, r.roleTable())
	row := r.db.QueryRowContext(ctx, q, id)
	item, err := scanRoleListRow(row)
	if err != nil {
		return nil, err
	}
	scopes, _ := r.ListDataScopesByRoleID(ctx, id)
	item.DataScopes = scopes
	for _, s := range scopes {
		if s.ResourceType == dataScopeResourceArticle {
			item.ArticleDataScope = s.ScopeType
		}
	}
	return &item, nil
}

func scanRoleListRow(scanner interface{ Scan(dest ...any) error }) (RoleListItemEntity, error) {
	var item RoleListItemEntity
	var ct, ut sql.NullTime
	if err := scanner.Scan(&item.ID, &item.RoleName, &item.RoleDesc, &ct, &ut); err != nil {
		return RoleListItemEntity{}, err
	}
	item.CreateTime = scanNullableTime(ct)
	item.UpdateTime = scanNullableTime(ut)
	item.DataScopes = []RoleDataScopeEntity{}
	return item, nil
}

func scanRoleDetailRow(row *sql.Row) (*RoleDetailEntity, error) {
	var d RoleDetailEntity
	var ct, ut sql.NullTime
	if err := row.Scan(&d.ID, &d.RoleName, &d.RoleDesc, &ct, &ut); err != nil {
		return nil, err
	}
	d.CreateTime = scanNullableTime(ct)
	d.UpdateTime = scanNullableTime(ut)
	return &d, nil
}

func (r *AdminRepo) listPrivilegeIDsByRoleID(ctx context.Context, roleID int) ([]int, error) {
	q := fmt.Sprintf("SELECT privilegeId FROM %s WHERE roleId = ?", r.prefix+"role_privileges_privilege")
	rows, err := r.db.QueryContext(ctx, q, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanIntColumn(rows)
}

func (r *AdminRepo) listMenuIDsByRoleID(ctx context.Context, roleID int) ([]string, error) {
	q := fmt.Sprintf("SELECT menuId FROM %s WHERE roleId = ?", r.roleMenusTable())
	rows, err := r.db.QueryContext(ctx, q, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *AdminRepo) replaceRolePrivileges(ctx context.Context, roleID int, privilegeIDs []int) error {
	del := fmt.Sprintf("DELETE FROM %s WHERE roleId = ?", r.prefix+"role_privileges_privilege")
	if _, err := r.db.ExecContext(ctx, del, roleID); err != nil {
		return err
	}
	for _, pid := range privilegeIDs {
		ins := fmt.Sprintf("INSERT INTO %s (roleId, privilegeId) VALUES (?, ?)", r.prefix+"role_privileges_privilege")
		if _, err := r.db.ExecContext(ctx, ins, roleID, pid); err != nil {
			return err
		}
	}
	return nil
}

func (r *AdminRepo) replaceRoleMenus(ctx context.Context, roleID int, menuIDs []string) error {
	del := fmt.Sprintf("DELETE FROM %s WHERE roleId = ?", r.roleMenusTable())
	if _, err := r.db.ExecContext(ctx, del, roleID); err != nil {
		return err
	}
	for _, mid := range menuIDs {
		ins := fmt.Sprintf("INSERT INTO %s (roleId, menuId) VALUES (?, ?)", r.roleMenusTable())
		if _, err := r.db.ExecContext(ctx, ins, roleID, mid); err != nil {
			return err
		}
	}
	return nil
}

func intSliceToStringSlice(ids []int) []string {
	out := make([]string, len(ids))
	for i, id := range ids {
		out[i] = fmt.Sprintf("%d", id)
	}
	return out
}

// ParsePrivilegeIDs 将前端 string/int 混合 ID 数组转为 int 切片。
func ParsePrivilegeIDs(raw []interface{}) ([]int, error) {
	out := make([]int, 0, len(raw))
	for _, v := range raw {
		switch n := v.(type) {
		case float64:
			out = append(out, int(n))
		case string:
			var id int
			if _, err := fmt.Sscan(n, &id); err != nil {
				return nil, err
			}
			out = append(out, id)
		case int:
			out = append(out, n)
		default:
			return nil, fmt.Errorf("invalid privilege id: %v", v)
		}
	}
	return out, nil
}

// ParseMenuIDs 将前端菜单 ID 数组转为 string 切片。
func ParseMenuIDs(raw []interface{}) []string {
	out := make([]string, 0, len(raw))
	for _, v := range raw {
		out = append(out, fmt.Sprint(v))
	}
	return out
}
