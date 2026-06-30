package repo

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// --- Privilege CRUD ---

// CreatePrivilege 创建权限。
func (r *AdminRepo) CreatePrivilege(ctx context.Context, p PrivilegeFullEntity) (*PrivilegeFullEntity, error) {
	isVisible, isPublic, reqOwn := boolToTiny(p.IsVisible), boolToTiny(p.IsPublic), boolToTiny(p.RequireOwnership)
	q := fmt.Sprintf(`INSERT INTO %s (privilegeName, privilegeCode, privilegePage, isVisible, pathPattern, httpMethod, isPublic, requireOwnership, description, createTime, updateTime)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())`, r.privilegeTable())
	res, err := r.db.ExecContext(ctx, q, p.PrivilegeName, p.PrivilegeCode, p.PrivilegePage, isVisible, p.PathPattern, p.HTTPMethod, isPublic, reqOwn, p.Description)
	if err != nil {
		return nil, err
	}
	id64, _ := res.LastInsertId()
	return r.GetPrivilegeByID(ctx, int(id64))
}

// ListPrivileges 分页查询权限。
func (r *AdminRepo) ListPrivileges(ctx context.Context, page, pageSize int, filters privilegeFilters) ([]PrivilegeFullEntity, NestPagination, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	where := []string{"1=1"}
	args := []any{}
	if filters.PrivilegeName != "" {
		where = append(where, "privilegeName LIKE ?")
		args = append(args, "%"+filters.PrivilegeName+"%")
	}
	if filters.PathPattern != "" {
		where = append(where, "pathPattern LIKE ?")
		args = append(args, "%"+filters.PathPattern+"%")
	}
	if filters.HTTPMethod != "" {
		where = append(where, "httpMethod = ?")
		args = append(args, filters.HTTPMethod)
	}
	if filters.IsPublic != nil {
		where = append(where, "isPublic = ?")
		args = append(args, boolToTiny(*filters.IsPublic))
	}
	if filters.IsVisible != nil {
		where = append(where, "isVisible = ?")
		args = append(args, boolToTiny(*filters.IsVisible))
	}
	whereSQL := "WHERE " + strings.Join(where, " AND ")
	var total int
	if err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s %s", r.privilegeTable(), whereSQL), args...).Scan(&total); err != nil {
		return nil, NestPagination{}, err
	}
	q := fmt.Sprintf(`SELECT id, privilegeName, privilegeCode, privilegePage, isVisible, pathPattern, httpMethod, isPublic, requireOwnership, description, createTime, updateTime
		FROM %s %s ORDER BY createTime ASC LIMIT ? OFFSET ?`, r.privilegeTable(), whereSQL)
	args = append(args, pageSize, (page-1)*pageSize)
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, NestPagination{}, err
	}
	defer rows.Close()
	list, err := scanPrivilegeFullRows(rows)
	if err != nil {
		return nil, NestPagination{}, err
	}
	menuNames, _ := r.loadMenuNameMap(ctx, collectPrivilegePages(list))
	for i := range list {
		list[i].PrivilegePageName = menuNames[list[i].PrivilegePage]
	}
	return list, CalcNestPagination(total, pageSize, page), nil
}

type privilegeFilters struct {
	PrivilegeName string
	PathPattern   string
	HTTPMethod    string
	IsPublic      *bool
	IsVisible     *bool
}

// PrivilegeListFilters 权限列表筛选（导出别名供 service 使用）。
type PrivilegeListFilters = privilegeFilters

// GetPrivilegeByID 查询权限详情。
func (r *AdminRepo) GetPrivilegeByID(ctx context.Context, id int) (*PrivilegeFullEntity, error) {
	q := fmt.Sprintf(`SELECT id, privilegeName, privilegeCode, privilegePage, isVisible, pathPattern, httpMethod, isPublic, requireOwnership, description, createTime, updateTime FROM %s WHERE id = ?`, r.privilegeTable())
	row := r.db.QueryRowContext(ctx, q, id)
	p, err := scanPrivilegeFullFromScanner(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return p, err
}

// UpdatePrivilege 更新权限。
func (r *AdminRepo) UpdatePrivilege(ctx context.Context, id int, fields map[string]interface{}) (*PrivilegeFullEntity, error) {
	existing, err := r.GetPrivilegeByID(ctx, id)
	if err != nil || existing == nil {
		return nil, err
	}
	sets := []string{"updateTime = NOW()"}
	args := []any{}
	for k, v := range fields {
		if k == "isVisible" || k == "isPublic" || k == "requireOwnership" {
			if b, ok := v.(bool); ok {
				v = boolToTiny(b)
			}
		}
		sets = append(sets, k+" = ?")
		args = append(args, v)
	}
	args = append(args, id)
	q := fmt.Sprintf("UPDATE %s SET %s WHERE id = ?", r.privilegeTable(), strings.Join(sets, ", "))
	if _, err := r.db.ExecContext(ctx, q, args...); err != nil {
		return nil, err
	}
	return r.GetPrivilegeByID(ctx, id)
}

// DeletePrivilege 删除权限。
func (r *AdminRepo) DeletePrivilege(ctx context.Context, id int) (bool, error) {
	existing, err := r.GetPrivilegeByID(ctx, id)
	if err != nil || existing == nil {
		return false, err
	}
	_, err = r.db.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s WHERE id = ?", r.privilegeTable()), id)
	return err == nil, err
}

func boolToTiny(v bool) int {
	if v {
		return 1
	}
	return 0
}

func tinyToBool(v int) bool { return v == 1 }

func scanPrivilegeFullRows(rows *sql.Rows) ([]PrivilegeFullEntity, error) {
	out := make([]PrivilegeFullEntity, 0)
	for rows.Next() {
		p, err := scanPrivilegeFullFromScanner(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *p)
	}
	return out, rows.Err()
}

func scanPrivilegeFullFromScanner(scanner interface{ Scan(dest ...any) error }) (*PrivilegeFullEntity, error) {
	var p PrivilegeFullEntity
	var desc sql.NullString
	var isVisible, isPublic, reqOwn int
	var ct, ut sql.NullTime
	if err := scanner.Scan(&p.ID, &p.PrivilegeName, &p.PrivilegeCode, &p.PrivilegePage, &isVisible, &p.PathPattern, &p.HTTPMethod, &isPublic, &reqOwn, &desc, &ct, &ut); err != nil {
		return nil, err
	}
	p.IsVisible = tinyToBool(isVisible)
	p.IsPublic = tinyToBool(isPublic)
	p.RequireOwnership = tinyToBool(reqOwn)
	if desc.Valid {
		p.Description = &desc.String
	}
	p.CreateTime = scanNullableTime(ct)
	p.UpdateTime = scanNullableTime(ut)
	return &p, nil
}

func collectPrivilegePages(list []PrivilegeFullEntity) []string {
	seen := make(map[string]struct{})
	out := make([]string, 0)
	for _, p := range list {
		if p.PrivilegePage == "" {
			continue
		}
		if _, ok := seen[p.PrivilegePage]; ok {
			continue
		}
		seen[p.PrivilegePage] = struct{}{}
		out = append(out, p.PrivilegePage)
	}
	return out
}

func (r *AdminRepo) loadMenuNameMap(ctx context.Context, menuIDs []string) (map[string]string, error) {
	out := make(map[string]string)
	if len(menuIDs) == 0 {
		return out, nil
	}
	ph := make([]string, len(menuIDs))
	args := make([]any, len(menuIDs))
	for i, id := range menuIDs {
		ph[i] = "?"
		args[i] = id
	}
	q := fmt.Sprintf(`SELECT id, menuCnName, name FROM %s WHERE id IN (%s)`, r.menuTable(), strings.Join(ph, ","))
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		var id, name string
		var cnNull sql.NullString
		if err := rows.Scan(&id, &cnNull, &name); err != nil {
			return out, err
		}
		label := name
		if cnNull.Valid && cnNull.String != "" {
			label = cnNull.String
		}
		out[id] = label
	}
	return out, rows.Err()
}
