// Package repo 角色与权限关联查询。
package repo

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	_ "github.com/go-sql-driver/mysql"
)

// RoleRepo 角色与用户-角色关联查询（直查 MySQL，规避 Ent 与 Nest 表结构差异）。
type RoleRepo struct {
	db     *sql.DB
	prefix string
}

// NewRoleRepo 构造 RoleRepo。
func NewRoleRepo(cfg *config.Config) (*RoleRepo, error) {
	db, err := sql.Open("mysql", cfg.MySQL.FormatDSN())
	if err != nil {
		return nil, fmt.Errorf("open mysql for role repo: %w", err)
	}
	return &RoleRepo{
		db:     db,
		prefix: cfg.MySQL.TablePrefixOrDefault(),
	}, nil
}

func (r *RoleRepo) roleTable() string       { return r.prefix + "role" }
func (r *RoleRepo) privilegeTable() string  { return r.prefix + "privilege" }
func (r *RoleRepo) roleUsersTable() string  { return r.prefix + "role_users_user" }
func (r *RoleRepo) rolePrivsTable() string  { return r.prefix + "role_privileges_privilege" }

// RoleEntity 角色及关联权限（info 接口用）。
type RoleEntity struct {
	ID         int               `json:"id"`
	RoleName   string            `json:"roleName"`
	RoleDesc   string            `json:"roleDesc"`
	Privileges []PrivilegeEntity `json:"privileges,omitempty"`
}

// PrivilegeEntity 权限项。
type PrivilegeEntity struct {
	ID               int     `json:"id"`
	PrivilegeName    string  `json:"privilegeName"`
	PrivilegeCode    string  `json:"privilegeCode"`
	PrivilegePage    string  `json:"privilegePage"`
	PathPattern      string  `json:"pathPattern"`
	HTTPMethod       string  `json:"httpMethod"`
	IsPublic         int     `json:"isPublic"`
	RequireOwnership int     `json:"requireOwnership"`
	Description      *string `json:"description,omitempty"`
}

// ListRolesByUserID 查询用户绑定的角色。
func (r *RoleRepo) ListRolesByUserID(ctx context.Context, userID int) ([]RoleEntity, error) {
	roleIDs, err := r.listRoleIDsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return r.loadRolesByIDs(ctx, roleIDs)
}

// ListRolesWithPrivileges 查询用户角色及每个角色的权限列表。
func (r *RoleRepo) ListRolesWithPrivileges(ctx context.Context, userID int) ([]RoleEntity, error) {
	roles, err := r.ListRolesByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	for i := range roles {
		privs, err := r.listPrivilegesByRoleID(ctx, roles[i].ID)
		if err != nil {
			return nil, err
		}
		roles[i].Privileges = privs
	}
	return roles, nil
}

func (r *RoleRepo) listPrivilegesByRoleID(ctx context.Context, roleID int) ([]PrivilegeEntity, error) {
	privIDs, err := r.listPrivilegeIDsByRoleID(ctx, roleID)
	if err != nil {
		return nil, err
	}
	return r.loadPrivilegesByIDs(ctx, privIDs)
}

// UserHasPrivilegeCode 判断用户是否拥有指定权限码。
func (r *RoleRepo) UserHasPrivilegeCode(ctx context.Context, userID int, code string) (bool, error) {
	roles, err := r.ListRolesWithPrivileges(ctx, userID)
	if err != nil {
		return false, err
	}
	for _, role := range roles {
		for _, p := range role.Privileges {
			if p.PrivilegeCode == code {
				return true, nil
			}
		}
	}
	return false, nil
}

// IsSuperAdmin 是否超级管理员。
func (r *RoleRepo) IsSuperAdmin(ctx context.Context, userID int) (bool, error) {
	roles, err := r.ListRolesByUserID(ctx, userID)
	if err != nil {
		return false, err
	}
	for _, role := range roles {
		if role.ID == 1 {
			return true, nil
		}
	}
	return false, nil
}

// BindRole 绑定用户角色（注册默认作者）。
func (r *RoleRepo) BindRole(ctx context.Context, userID, roleID int) error {
	q := fmt.Sprintf("INSERT INTO %s (userId, roleId) VALUES (?, ?)", r.roleUsersTable())
	_, err := r.db.ExecContext(ctx, q, userID, roleID)
	return err
}

// RolesExist 判断 roleIDs 是否全部存在。
func (r *RoleRepo) RolesExist(ctx context.Context, roleIDs []int) (bool, error) {
	if len(roleIDs) == 0 {
		return true, nil
	}
	ph, args := inPlaceholders(roleIDs)
	q := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE id IN (%s)", r.roleTable(), ph)
	var n int
	if err := r.db.QueryRowContext(ctx, q, args...).Scan(&n); err != nil {
		return false, err
	}
	return n == len(roleIDs), nil
}

// LoadAllPrivileges 加载全部权限配置（permission 中间件 DB 回退）。
func (r *RoleRepo) LoadAllPrivileges(ctx context.Context) ([]PrivilegeEntity, error) {
	q := fmt.Sprintf(`SELECT id, privilegeName, privilegeCode, privilegePage, pathPattern, httpMethod, isPublic, requireOwnership, description FROM %s`, r.privilegeTable())
	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]PrivilegeEntity, 0)
	for rows.Next() {
		p, err := scanPrivilege(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// PrivilegeCodesByRoleID 查询角色绑定的权限码列表（permission 中间件 Redis 回退用）。
func (r *RoleRepo) PrivilegeCodesByRoleID(ctx context.Context, roleID int) ([]string, error) {
	privs, err := r.listPrivilegesByRoleID(ctx, roleID)
	if err != nil {
		return nil, err
	}
	codes := make([]string, 0, len(privs))
	for _, p := range privs {
		codes = append(codes, p.PrivilegeCode)
	}
	return codes, nil
}

// ReplaceUserRoles 替换用户全部角色绑定。
func (r *RoleRepo) ReplaceUserRoles(ctx context.Context, userID int, roleIDs []int) error {
	del := fmt.Sprintf("DELETE FROM %s WHERE userId = ?", r.roleUsersTable())
	if _, err := r.db.ExecContext(ctx, del, userID); err != nil {
		return err
	}
	for _, rid := range roleIDs {
		if err := r.BindRole(ctx, userID, rid); err != nil {
			return err
		}
	}
	return nil
}

func (r *RoleRepo) listRoleIDsByUserID(ctx context.Context, userID int) ([]int, error) {
	q := fmt.Sprintf("SELECT roleId FROM %s WHERE userId = ?", r.roleUsersTable())
	rows, err := r.db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanIntColumn(rows)
}

func (r *RoleRepo) listPrivilegeIDsByRoleID(ctx context.Context, roleID int) ([]int, error) {
	q := fmt.Sprintf("SELECT privilegeId FROM %s WHERE roleId = ?", r.rolePrivsTable())
	rows, err := r.db.QueryContext(ctx, q, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanIntColumn(rows)
}

func (r *RoleRepo) loadRolesByIDs(ctx context.Context, ids []int) ([]RoleEntity, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	ph, args := inPlaceholders(ids)
	q := fmt.Sprintf("SELECT id, roleName, roleDesc FROM %s WHERE id IN (%s)", r.roleTable(), ph)
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]RoleEntity, 0, len(ids))
	for rows.Next() {
		var item RoleEntity
		if err := rows.Scan(&item.ID, &item.RoleName, &item.RoleDesc); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *RoleRepo) loadPrivilegesByIDs(ctx context.Context, ids []int) ([]PrivilegeEntity, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	ph, args := inPlaceholders(ids)
	q := fmt.Sprintf(`SELECT id, privilegeName, privilegeCode, privilegePage, pathPattern, httpMethod, isPublic, requireOwnership, description FROM %s WHERE id IN (%s)`, r.privilegeTable(), ph)
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]PrivilegeEntity, 0, len(ids))
	for rows.Next() {
		p, err := scanPrivilege(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func scanPrivilege(rows *sql.Rows) (PrivilegeEntity, error) {
	var p PrivilegeEntity
	var desc sql.NullString
	if err := rows.Scan(
		&p.ID, &p.PrivilegeName, &p.PrivilegeCode, &p.PrivilegePage,
		&p.PathPattern, &p.HTTPMethod, &p.IsPublic, &p.RequireOwnership, &desc,
	); err != nil {
		return PrivilegeEntity{}, err
	}
	if desc.Valid {
		p.Description = &desc.String
	}
	return p, nil
}

func scanIntColumn(rows *sql.Rows) ([]int, error) {
	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func inPlaceholders(ids []int) (string, []any) {
	parts := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		parts[i] = "?"
		args[i] = id
	}
	return strings.Join(parts, ","), args
}
