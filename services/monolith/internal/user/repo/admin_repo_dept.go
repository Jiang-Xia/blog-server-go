package repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
)

// --- Dept CRUD ---

// CreateDept 创建部门。
func (r *AdminRepo) CreateDept(ctx context.Context, d DeptEntity) (*DeptEntity, error) {
	q := fmt.Sprintf(`INSERT INTO %s (deptName, deptCode, parentId, leaderId, leaderName, orderNum, status, remark, createTime, updateTime)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())`, r.deptTable())
	res, err := r.db.ExecContext(ctx, q, d.DeptName, d.DeptCode, d.ParentID, d.LeaderID, d.LeaderName, d.OrderNum, d.Status, d.Remark)
	if err != nil {
		return nil, err
	}
	id64, _ := res.LastInsertId()
	return r.GetDeptByID(ctx, int(id64))
}

// ListDepts 分页查询部门，accessibleDeptIDs 为 nil 时不限制。
func (r *AdminRepo) ListDepts(ctx context.Context, page, pageSize int, deptName string, parentID *int, status *int, accessibleDeptIDs []int) ([]DeptEntity, NestPagination, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if accessibleDeptIDs != nil && len(accessibleDeptIDs) == 0 {
		return []DeptEntity{}, CalcNestPagination(0, pageSize, page), nil
	}
	where := []string{"1=1"}
	args := []any{}
	if accessibleDeptIDs != nil {
		ph, a := inPlaceholders(accessibleDeptIDs)
		where = append(where, fmt.Sprintf("id IN (%s)", ph))
		args = append(args, a...)
	}
	if deptName != "" {
		where = append(where, "deptName LIKE ?")
		args = append(args, "%"+deptName+"%")
	}
	if status != nil {
		where = append(where, "status = ?")
		args = append(args, *status)
	}
	if parentID != nil {
		where = append(where, "parentId = ?")
		args = append(args, *parentID)
	}
	whereSQL := "WHERE " + strings.Join(where, " AND ")
	var total int
	if err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s %s", r.deptTable(), whereSQL), args...).Scan(&total); err != nil {
		return nil, NestPagination{}, err
	}
	q := fmt.Sprintf(`SELECT id, deptName, deptCode, parentId, leaderId, leaderName, orderNum, status, remark, createTime, updateTime
		FROM %s %s ORDER BY orderNum ASC, createTime ASC LIMIT ? OFFSET ?`, r.deptTable(), whereSQL)
	args = append(args, pageSize, (page-1)*pageSize)
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, NestPagination{}, err
	}
	defer rows.Close()
	list, err := scanDeptRows(rows)
	if err != nil {
		return nil, NestPagination{}, err
	}
	return list, CalcNestPagination(total, pageSize, page), nil
}

// GetDeptByID 查询部门详情。
func (r *AdminRepo) GetDeptByID(ctx context.Context, id int) (*DeptEntity, error) {
	q := fmt.Sprintf(`SELECT id, deptName, deptCode, parentId, leaderId, leaderName, orderNum, status, remark, createTime, updateTime FROM %s WHERE id = ?`, r.deptTable())
	row := r.db.QueryRowContext(ctx, q, id)
	d, err := scanDeptRow(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return d, err
}

// UpdateDept 更新部门。
func (r *AdminRepo) UpdateDept(ctx context.Context, id int, fields map[string]interface{}) (*DeptEntity, error) {
	existing, err := r.GetDeptByID(ctx, id)
	if err != nil || existing == nil {
		return nil, err
	}
	sets := []string{"updateTime = NOW()"}
	args := []any{}
	for k, v := range fields {
		sets = append(sets, k+" = ?")
		args = append(args, v)
	}
	args = append(args, id)
	q := fmt.Sprintf("UPDATE %s SET %s WHERE id = ?", r.deptTable(), strings.Join(sets, ", "))
	if _, err := r.db.ExecContext(ctx, q, args...); err != nil {
		return nil, err
	}
	return r.GetDeptByID(ctx, id)
}

// DeleteDept 删除部门（存在子部门时由 service 层校验）。
func (r *AdminRepo) DeleteDept(ctx context.Context, id int) (bool, error) {
	existing, err := r.GetDeptByID(ctx, id)
	if err != nil || existing == nil {
		return false, err
	}
	_, err = r.db.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s WHERE id = ?", r.deptTable()), id)
	return err == nil, err
}

// CountChildDepts 统计子部门数量。
func (r *AdminRepo) CountChildDepts(ctx context.Context, parentID int) (int, error) {
	var n int
	err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE parentId = ?", r.deptTable()), parentID).Scan(&n)
	return n, err
}

// ListAllDepts 查询全部部门（树构建用）。
func (r *AdminRepo) ListAllDepts(ctx context.Context) ([]DeptEntity, error) {
	q := fmt.Sprintf(`SELECT id, deptName, deptCode, parentId, leaderId, leaderName, orderNum, status, remark, createTime, updateTime FROM %s ORDER BY orderNum ASC, createTime ASC`, r.deptTable())
	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanDeptRows(rows)
}

// GetDescendantDeptIDs 获取部门及全部子孙 ID。
func (r *AdminRepo) GetDescendantDeptIDs(ctx context.Context, deptID int) ([]int, error) {
	all, err := r.ListAllDepts(ctx)
	if err != nil {
		return nil, err
	}
	childrenMap := make(map[int][]int)
	for _, d := range all {
		if d.ParentID > 0 {
			childrenMap[d.ParentID] = append(childrenMap[d.ParentID], d.ID)
		}
	}
	result := []int{deptID}
	queue := []int{deptID}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, child := range childrenMap[cur] {
			result = append(result, child)
			queue = append(queue, child)
		}
	}
	return result, nil
}

func scanDeptRows(rows *sql.Rows) ([]DeptEntity, error) {
	out := make([]DeptEntity, 0)
	for rows.Next() {
		d, err := scanDeptFromScanner(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *d)
	}
	return out, rows.Err()
}

func scanDeptRow(row *sql.Row) (*DeptEntity, error) {
	return scanDeptFromScanner(row)
}

func scanDeptFromScanner(scanner interface{ Scan(dest ...any) error }) (*DeptEntity, error) {
	var d DeptEntity
	var leaderID, leaderName, remark sql.NullString
	var ct, ut sql.NullTime
	if err := scanner.Scan(&d.ID, &d.DeptName, &d.DeptCode, &d.ParentID, &leaderID, &leaderName, &d.OrderNum, &d.Status, &remark, &ct, &ut); err != nil {
		return nil, err
	}
	if leaderID.Valid {
		d.LeaderID = &leaderID.String
	}
	if leaderName.Valid {
		d.LeaderName = &leaderName.String
	}
	if remark.Valid {
		d.Remark = &remark.String
	}
	d.CreateTime = scanNullableTime(ct)
	d.UpdateTime = scanNullableTime(ut)
	return &d, nil
}

// FilterDeptsByScope 按数据权限过滤并保留祖先路径。
func FilterDeptsByScope(all []DeptEntity, accessibleDeptIDs []int) []DeptEntity {
	if accessibleDeptIDs == nil {
		return all
	}
	if len(accessibleDeptIDs) == 0 {
		return []DeptEntity{}
	}
	allowed := make(map[int]struct{}, len(accessibleDeptIDs))
	for _, id := range accessibleDeptIDs {
		allowed[id] = struct{}{}
	}
	byID := make(map[int]DeptEntity, len(all))
	for _, d := range all {
		byID[d.ID] = d
	}
	keep := make(map[int]struct{})
	for id := range allowed {
		cur, ok := byID[id]
		for ok {
			keep[cur.ID] = struct{}{}
			if cur.ParentID <= 0 {
				break
			}
			cur, ok = byID[cur.ParentID]
		}
	}
	out := make([]DeptEntity, 0)
	for _, d := range all {
		if _, ok := keep[d.ID]; ok {
			out = append(out, d)
		}
	}
	return out
}

// BuildDeptTree 构建部门树。
func BuildDeptTree(all []DeptEntity, rootParentID *string, filters deptTreeFilters) []DeptEntity {
	filtered := all
	if filters.hasFilter() {
		byID := make(map[int]DeptEntity, len(all))
		for _, d := range all {
			byID[d.ID] = d
		}
		keep := make(map[int]struct{})
		for _, d := range all {
			if !filters.match(d) {
				continue
			}
			cur, ok := byID[d.ID]
			for ok {
				keep[cur.ID] = struct{}{}
				if cur.ParentID <= 0 {
					break
				}
				cur, ok = byID[cur.ParentID]
			}
		}
		tmp := make([]DeptEntity, 0)
		for _, d := range all {
			if _, ok := keep[d.ID]; ok {
				tmp = append(tmp, d)
			}
		}
		filtered = tmp
	}
	nodeMap := make(map[int]*DeptEntity, len(filtered))
	for i := range filtered {
		filtered[i].Children = []DeptEntity{}
		nodeMap[filtered[i].ID] = &filtered[i]
	}
	roots := make([]*DeptEntity, 0)
	for i := range filtered {
		d := &filtered[i]
		if d.ParentID == 0 {
			roots = append(roots, d)
			continue
		}
		if parent, ok := nodeMap[d.ParentID]; ok {
			parent.Children = append(parent.Children, *d)
		}
	}
	if rootParentID != nil && *rootParentID != "" {
		var rootID int
		fmt.Sscan(*rootParentID, &rootID)
		if node, ok := nodeMap[rootID]; ok {
			return []DeptEntity{materializeDeptNode(nodeMap, node.ID)}
		}
		return []DeptEntity{}
	}
	out := make([]DeptEntity, 0, len(roots))
	for _, root := range roots {
		out = append(out, materializeDeptNode(nodeMap, root.ID))
	}
	return out
}

func materializeDeptNode(nodeMap map[int]*DeptEntity, id int) DeptEntity {
	base := *nodeMap[id]
	if len(base.Children) == 0 {
		return base
	}
	children := make([]DeptEntity, 0, len(base.Children))
	for _, ch := range base.Children {
		children = append(children, materializeDeptNode(nodeMap, ch.ID))
	}
	base.Children = children
	return base
}

type deptTreeFilters struct {
	DeptName string
	Status   *int
}

// NewDeptTreeFilters 构造部门树筛选条件。
func NewDeptTreeFilters(deptName string, status *int) deptTreeFilters {
	return deptTreeFilters{DeptName: deptName, Status: status}
}

func (f deptTreeFilters) hasFilter() bool {
	return f.DeptName != "" || f.Status != nil
}

func (f deptTreeFilters) match(d DeptEntity) bool {
	if f.DeptName != "" && !strings.Contains(d.DeptName, f.DeptName) {
		return false
	}
	if f.Status != nil && d.Status != *f.Status {
		return false
	}
	return true
}

// --- Data Scope ---

// ListDataScopesByRoleID 查询角色数据权限。
func (r *AdminRepo) ListDataScopesByRoleID(ctx context.Context, roleID int) ([]RoleDataScopeEntity, error) {
	q := fmt.Sprintf(`SELECT id, roleId, resourceType, scopeType, deptIds, createTime, updateTime FROM %s WHERE roleId = ? ORDER BY resourceType ASC`, r.roleDataScopeTable())
	rows, err := r.db.QueryContext(ctx, q, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]RoleDataScopeEntity, 0)
	for rows.Next() {
		var item RoleDataScopeEntity
		var deptRaw sql.NullString
		var ct, ut sql.NullTime
		if err := rows.Scan(&item.ID, &item.RoleID, &item.ResourceType, &item.ScopeType, &deptRaw, &ct, &ut); err != nil {
			return nil, err
		}
		item.DeptIDs = parseDeptIDsJSON(deptRaw)
		item.CreateTime = scanNullableTime(ct)
		item.UpdateTime = scanNullableTime(ut)
		out = append(out, item)
	}
	return out, rows.Err()
}

// UpsertRoleDataScopes 按 resourceType upsert 数据权限。
func (r *AdminRepo) UpsertRoleDataScopes(ctx context.Context, roleID int, scopes []RoleDataScopeEntity) ([]RoleDataScopeEntity, error) {
	for _, s := range scopes {
		var deptJSON interface{}
		if s.ScopeType == "CUSTOM" {
			b, _ := json.Marshal(s.DeptIDs)
			deptJSON = string(b)
		}
		existingID := 0
		err := r.db.QueryRowContext(ctx,
			fmt.Sprintf("SELECT id FROM %s WHERE roleId = ? AND resourceType = ?", r.roleDataScopeTable()),
			roleID, s.ResourceType,
		).Scan(&existingID)
		if err == sql.ErrNoRows {
			_, err = r.db.ExecContext(ctx,
				fmt.Sprintf(`INSERT INTO %s (roleId, resourceType, scopeType, deptIds, createTime, updateTime) VALUES (?, ?, ?, ?, NOW(), NOW())`, r.roleDataScopeTable()),
				roleID, s.ResourceType, s.ScopeType, deptJSON,
			)
		} else if err == nil {
			_, err = r.db.ExecContext(ctx,
				fmt.Sprintf(`UPDATE %s SET scopeType = ?, deptIds = ?, updateTime = NOW() WHERE id = ?`, r.roleDataScopeTable()),
				s.ScopeType, deptJSON, existingID,
			)
		}
		if err != nil {
			return nil, err
		}
	}
	return r.ListDataScopesByRoleID(ctx, roleID)
}

func parseDeptIDsJSON(raw sql.NullString) []int {
	if !raw.Valid || raw.String == "" || raw.String == "null" {
		return nil
	}
	var ids []int
	if err := json.Unmarshal([]byte(raw.String), &ids); err == nil {
		return ids
	}
	return nil
}

// ResolveAccessibleDeptIDs 解析用户可访问部门；null=全部，[] = 无。
func (r *AdminRepo) ResolveAccessibleDeptIDs(ctx context.Context, userDeptID *int, roleIDs []int, resourceType string) ([]int, error) {
	for _, rid := range roleIDs {
		if rid == 1 {
			return nil, nil
		}
	}
	if len(roleIDs) == 0 {
		return []int{}, nil
	}
	ph, args := inPlaceholders(roleIDs)
	args = append(args, resourceType)
	q := fmt.Sprintf(`SELECT scopeType, deptIds FROM %s WHERE roleId IN (%s) AND resourceType = ?`, r.roleDataScopeTable(), ph)
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	type scopeRow struct {
		scopeType string
		deptIDs   []int
	}
	scopes := make([]scopeRow, 0)
	for rows.Next() {
		var st string
		var deptRaw sql.NullString
		if err := rows.Scan(&st, &deptRaw); err != nil {
			return nil, err
		}
		scopes = append(scopes, scopeRow{scopeType: st, deptIDs: parseDeptIDsJSON(deptRaw)})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(scopes) == 0 {
		return []int{}, nil
	}
	for _, s := range scopes {
		if s.scopeType == "ALL" {
			return nil, nil
		}
	}
	accessible := make(map[int]struct{})
	for _, s := range scopes {
		switch s.scopeType {
		case "DEPT":
			if userDeptID != nil {
				accessible[*userDeptID] = struct{}{}
			}
		case "DEPT_AND_CHILDREN":
			if userDeptID != nil {
				desc, _ := r.GetDescendantDeptIDs(ctx, *userDeptID)
				for _, id := range desc {
					accessible[id] = struct{}{}
				}
			}
		case "CUSTOM":
			for _, id := range s.deptIDs {
				accessible[id] = struct{}{}
			}
		}
	}
	out := make([]int, 0, len(accessible))
	for id := range accessible {
		out = append(out, id)
	}
	return out, nil
}
