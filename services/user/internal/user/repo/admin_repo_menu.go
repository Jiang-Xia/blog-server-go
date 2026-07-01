package repo

import (
	"context"
	"database/sql"
	"fmt"
)

// --- Menu CRUD ---

func (r *AdminRepo) listAllMenus(ctx context.Context) ([]MenuEntity, error) {
	q := fmt.Sprintf(`SELECT id, pid, path, name, menuCnName, `+"`order`"+`, icon, locale, requiresAuth, filePath FROM %s ORDER BY `+"`order`"+` ASC`, r.menuTable())
	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMenuRows(rows)
}

// ListMenusByRoleID 查询角色绑定的菜单。
func (r *AdminRepo) ListMenusByRoleID(ctx context.Context, roleID int) ([]MenuEntity, error) {
	q := fmt.Sprintf(`SELECT m.id, m.pid, m.path, m.name, m.menuCnName, m.`+"`order`"+`, m.icon, m.locale, m.requiresAuth, m.filePath
		FROM %s m INNER JOIN %s rm ON m.id = rm.menuId WHERE rm.roleId = ? ORDER BY m.`+"`order`"+` ASC`, r.menuTable(), r.roleMenusTable())
	rows, err := r.db.QueryContext(ctx, q, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMenuRows(rows)
}

// ListMenusByUserID 合并用户全部角色的菜单并去重。
func (r *AdminRepo) ListMenusByUserID(ctx context.Context, roleIDs []int) ([]MenuEntity, error) {
	if len(roleIDs) == 0 {
		return []MenuEntity{}, nil
	}
	seen := make(map[string]MenuEntity)
	for _, rid := range roleIDs {
		menus, err := r.ListMenusByRoleID(ctx, rid)
		if err != nil {
			return nil, err
		}
		for _, m := range menus {
			seen[m.ID] = m
		}
	}
	out := make([]MenuEntity, 0, len(seen))
	for _, m := range seen {
		out = append(out, m)
	}
	return out, nil
}

// BuildMenuTree 构建带 meta 的动态菜单树。
func BuildMenuTree(menus []MenuEntity) []MenuTreeNode {
	nodeMap := make(map[string]*MenuTreeNode, len(menus))
	for _, m := range menus {
		nodeMap[m.ID] = &MenuTreeNode{
			ID: m.ID, PID: m.PID, Path: m.Path, Name: m.Name, MenuCnName: m.MenuCnName, FilePath: m.FilePath,
			Meta: MenuMeta{Order: m.Order, Icon: m.Icon, Locale: m.Locale, RequiresAuth: m.RequiresAuth},
			Children: []MenuTreeNode{},
		}
	}
	roots := make([]*MenuTreeNode, 0)
	for _, m := range menus {
		node := nodeMap[m.ID]
		if m.PID == "0" || m.PID == "" {
			roots = append(roots, node)
			continue
		}
		if parent, ok := nodeMap[m.PID]; ok {
			parent.Children = append(parent.Children, *node)
		}
	}
	out := make([]MenuTreeNode, 0, len(roots))
	for _, root := range roots {
		out = append(out, materializeMenuNode(nodeMap, root.ID))
	}
	return out
}

func materializeMenuNode(nodeMap map[string]*MenuTreeNode, id string) MenuTreeNode {
	base := *nodeMap[id]
	if len(base.Children) == 0 {
		return base
	}
	children := make([]MenuTreeNode, 0, len(base.Children))
	for _, ch := range base.Children {
		children = append(children, materializeMenuNode(nodeMap, ch.ID))
	}
	base.Children = children
	return base
}

// CreateMenu 创建菜单。
func (r *AdminRepo) CreateMenu(ctx context.Context, m MenuEntity) (*MenuEntity, error) {
	var exists int
	if err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE id = ?", r.menuTable()), m.ID).Scan(&exists); err != nil {
		return nil, err
	}
	if exists > 0 {
		return nil, fmt.Errorf("menu exists")
	}
	if m.PID == "" {
		m.PID = "0"
	}
	reqAuth := 0
	if m.RequiresAuth {
		reqAuth = 1
	}
	q := fmt.Sprintf(`INSERT INTO %s (id, pid, path, name, menuCnName, `+"`order`"+`, icon, locale, requiresAuth, filePath, isDelete) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0)`, r.menuTable())
	_, err := r.db.ExecContext(ctx, q, m.ID, m.PID, m.Path, m.Name, m.MenuCnName, m.Order, m.Icon, m.Locale, reqAuth, m.FilePath)
	if err != nil {
		return nil, err
	}
	return r.GetMenuByID(ctx, m.ID)
}

// GetMenuByID 查询菜单。
func (r *AdminRepo) GetMenuByID(ctx context.Context, id string) (*MenuEntity, error) {
	q := fmt.Sprintf(`SELECT id, pid, path, name, menuCnName, `+"`order`"+`, icon, locale, requiresAuth, filePath FROM %s WHERE id = ?`, r.menuTable())
	row := r.db.QueryRowContext(ctx, q, id)
	m, err := scanMenuFromScanner(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return m, err
}

// UpdateMenu 局部更新菜单。
func (r *AdminRepo) UpdateMenu(ctx context.Context, id string, fields map[string]interface{}) (*MenuEntity, error) {
	existing, err := r.GetMenuByID(ctx, id)
	if err != nil || existing == nil {
		return nil, err
	}
	sets := make([]string, 0, len(fields))
	args := make([]any, 0, len(fields)+1)
	for k, v := range fields {
		if k == "requiresAuth" {
			if b, ok := v.(bool); ok {
				if b {
					v = 1
				} else {
					v = 0
				}
			}
			sets = append(sets, "`requiresAuth` = ?")
		} else if k == "order" {
			sets = append(sets, "`order` = ?")
		} else {
			sets = append(sets, k+" = ?")
		}
		args = append(args, v)
	}
	args = append(args, id)
	q := fmt.Sprintf("UPDATE %s SET %s WHERE id = ?", r.menuTable(), joinSets(sets))
	if _, err := r.db.ExecContext(ctx, q, args...); err != nil {
		return nil, err
	}
	return r.GetMenuByID(ctx, id)
}

// DeleteMenu 删除菜单。
func (r *AdminRepo) DeleteMenu(ctx context.Context, id string) (bool, error) {
	existing, err := r.GetMenuByID(ctx, id)
	if err != nil || existing == nil {
		return false, err
	}
	_, err = r.db.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s WHERE id = ?", r.menuTable()), id)
	return err == nil, err
}

func scanMenuRows(rows *sql.Rows) ([]MenuEntity, error) {
	out := make([]MenuEntity, 0)
	for rows.Next() {
		m, err := scanMenuFromScanner(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *m)
	}
	return out, rows.Err()
}

func scanMenuFromScanner(scanner interface{ Scan(dest ...any) error }) (*MenuEntity, error) {
	var m MenuEntity
	var cn sql.NullString
	var reqAuth int
	if err := scanner.Scan(&m.ID, &m.PID, &m.Path, &m.Name, &cn, &m.Order, &m.Icon, &m.Locale, &reqAuth, &m.FilePath); err != nil {
		return nil, err
	}
	if cn.Valid {
		m.MenuCnName = &cn.String
	}
	m.RequiresAuth = reqAuth == 1
	if m.PID == "" {
		m.PID = "0"
	}
	return &m, nil
}

func joinSets(sets []string) string {
	if len(sets) == 0 {
		return "id = id"
	}
	out := sets[0]
	for i := 1; i < len(sets); i++ {
		out += ", " + sets[i]
	}
	return out
}
