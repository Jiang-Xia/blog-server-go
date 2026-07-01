// Package admin RBAC 后台管理业务逻辑，对齐 Nest admin/system 与 admin/menu。
package admin

import (
	"context"
	"fmt"

	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/services/user/internal/user/repo"
)

const dataScopeResourceArticle = "article"

// Service RBAC 后台管理服务。
type Service struct {
	admin *repo.AdminRepo
	roles *repo.RoleRepo
	users *repo.UserRepo
}

// NewService 构造 admin Service。
func NewService(admin *repo.AdminRepo, roles *repo.RoleRepo, users *repo.UserRepo) *Service {
	return &Service{admin: admin, roles: roles, users: users}
}

// RoleListResult 角色分页列表。
type RoleListResult struct {
	List       []repo.RoleListItemEntity `json:"list"`
	Pagination repo.NestPagination       `json:"pagination"`
}

// CreateRole 创建角色。
func (s *Service) CreateRole(ctx context.Context, roleName, roleDesc string, privilegeRaw, menuRaw []interface{}) (interface{}, error) {
	privIDs, err := repo.ParsePrivilegeIDs(privilegeRaw)
	if err != nil {
		return nil, errcode.WithMessage(errcode.InvalidParam, "权限 ID 格式错误")
	}
	menuIDs := repo.ParseMenuIDs(menuRaw)
	return s.admin.CreateRole(ctx, roleName, roleDesc, privIDs, menuIDs)
}

// ListRoles 分页查询角色。
func (s *Service) ListRoles(ctx context.Context, page, pageSize int, roleName string) (*RoleListResult, error) {
	list, pageInfo, err := s.admin.ListRoles(ctx, page, pageSize, roleName)
	if err != nil {
		return nil, err
	}
	return &RoleListResult{List: list, Pagination: pageInfo}, nil
}

// GetRole 查询角色详情。
func (s *Service) GetRole(ctx context.Context, id int) (interface{}, error) {
	role, err := s.admin.GetRoleDetail(ctx, id)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, errcode.WithMessage(errcode.NotFound, "角色不存在")
	}
	return role, nil
}

// UpdateRole 更新角色权限/菜单。
func (s *Service) UpdateRole(ctx context.Context, id int, privilegeRaw, menuRaw []interface{}) (interface{}, error) {
	privIDs, err := repo.ParsePrivilegeIDs(privilegeRaw)
	if err != nil {
		return nil, errcode.WithMessage(errcode.InvalidParam, "权限 ID 格式错误")
	}
	menuIDs := repo.ParseMenuIDs(menuRaw)
	role, err := s.admin.UpdateRole(ctx, id, privIDs, menuIDs)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, errcode.WithMessage(errcode.NotFound, "角色不存在")
	}
	return role, nil
}

// DeleteRole 删除角色。
func (s *Service) DeleteRole(ctx context.Context, id int) (bool, error) {
	ok, err := s.admin.DeleteRole(ctx, id)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, errcode.WithMessage(errcode.NotFound, "角色不存在")
	}
	return true, nil
}

// MenuPrivilegeTree 菜单+权限树。
func (s *Service) MenuPrivilegeTree(ctx context.Context) (interface{}, error) {
	return s.admin.GetMenuPrivilegeTree(ctx)
}

// GetRoleDataScopes 查询角色数据权限。
func (s *Service) GetRoleDataScopes(ctx context.Context, roleID int) (interface{}, error) {
	exists, err := s.roleExists(ctx, roleID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errcode.WithMessage(errcode.NotFound, "角色不存在")
	}
	return s.admin.ListDataScopesByRoleID(ctx, roleID)
}

// UpdateRoleDataScopes 更新角色数据权限。
func (s *Service) UpdateRoleDataScopes(ctx context.Context, roleID int, scopes []repo.RoleDataScopeEntity) (interface{}, error) {
	exists, err := s.roleExists(ctx, roleID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errcode.WithMessage(errcode.NotFound, "角色不存在")
	}
	return s.admin.UpsertRoleDataScopes(ctx, roleID, scopes)
}

func (s *Service) roleExists(ctx context.Context, roleID int) (bool, error) {
	role, err := s.admin.GetRoleDetail(ctx, roleID)
	if err != nil {
		return false, err
	}
	return role != nil, nil
}

// DeptListResult 部门分页列表。
type DeptListResult struct {
	List       []repo.DeptEntity     `json:"list"`
	Pagination repo.NestPagination   `json:"pagination"`
}

// CreateDept 创建部门。
func (s *Service) CreateDept(ctx context.Context, d repo.DeptEntity) (interface{}, error) {
	return s.admin.CreateDept(ctx, d)
}

// ListDepts 分页查询部门（含数据权限过滤）。
func (s *Service) ListDepts(ctx context.Context, uid int, page, pageSize int, deptName string, parentID *int, status *int) (*DeptListResult, error) {
	accessible, err := s.resolveAccessibleDeptIDs(ctx, uid)
	if err != nil {
		return nil, err
	}
	list, pageInfo, err := s.admin.ListDepts(ctx, page, pageSize, deptName, parentID, status, accessible)
	if err != nil {
		return nil, err
	}
	return &DeptListResult{List: list, Pagination: pageInfo}, nil
}

// DeptTree 部门树（含数据权限过滤）。
func (s *Service) DeptTree(ctx context.Context, uid int, rootID, deptName string, status *int) (interface{}, error) {
	accessible, err := s.resolveAccessibleDeptIDs(ctx, uid)
	if err != nil {
		return nil, err
	}
	all, err := s.admin.ListAllDepts(ctx)
	if err != nil {
		return nil, err
	}
	if accessible != nil {
		all = repo.FilterDeptsByScope(all, accessible)
	}
	filters := repo.NewDeptTreeFilters(deptName, status)
	var rootPtr *string
	if rootID != "" {
		rootPtr = &rootID
	}
	return repo.BuildDeptTree(all, rootPtr, filters), nil
}

// GetDept 查询部门详情。
func (s *Service) GetDept(ctx context.Context, id int) (interface{}, error) {
	d, err := s.admin.GetDeptByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if d == nil {
		return nil, errcode.WithMessage(errcode.NotFound, "部门不存在")
	}
	return d, nil
}

// UpdateDept 更新部门。
func (s *Service) UpdateDept(ctx context.Context, id int, fields map[string]interface{}) (interface{}, error) {
	d, err := s.admin.UpdateDept(ctx, id, fields)
	if err != nil {
		return nil, err
	}
	if d == nil {
		return nil, errcode.WithMessage(errcode.NotFound, "部门不存在")
	}
	return d, nil
}

// DeleteDept 删除部门。
func (s *Service) DeleteDept(ctx context.Context, id int) (bool, error) {
	n, err := s.admin.CountChildDepts(ctx, id)
	if err != nil {
		return false, err
	}
	if n > 0 {
		return false, errcode.WithMessage(errcode.InvalidParam, "该部门下存在子部门，无法删除")
	}
	ok, err := s.admin.DeleteDept(ctx, id)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, errcode.WithMessage(errcode.NotFound, "部门不存在")
	}
	return true, nil
}

func (s *Service) resolveAccessibleDeptIDs(ctx context.Context, uid int) ([]int, error) {
	if uid == 0 {
		return []int{}, nil
	}
	roles, err := s.roles.ListRolesByUserID(ctx, uid)
	if err != nil {
		return nil, err
	}
	roleIDs := make([]int, 0, len(roles))
	for _, r := range roles {
		roleIDs = append(roleIDs, r.ID)
	}
	for _, rid := range roleIDs {
		if rid == 1 {
			return nil, nil
		}
	}
	u, err := s.users.FindByID(ctx, uid)
	if err != nil {
		return nil, err
	}
	var deptID *int
	if u.DeptId != nil {
		deptID = u.DeptId
	}
	return s.admin.ResolveAccessibleDeptIDs(ctx, deptID, roleIDs, dataScopeResourceArticle)
}

// ResolveArticleAccessibleDeptIDs 解析用户对 article 资源的可访问机构；nil 表示全部（超管）。
func (s *Service) ResolveArticleAccessibleDeptIDs(ctx context.Context, uid int) ([]int, error) {
	return s.resolveAccessibleDeptIDs(ctx, uid)
}

// AssertArticleDeptAccess 校验用户是否有权访问指定机构下的文章。
func (s *Service) AssertArticleDeptAccess(ctx context.Context, uid int, articleDeptID *int) error {
	deptIDs, err := s.resolveAccessibleDeptIDs(ctx, uid)
	if err != nil {
		return err
	}
	if deptIDs == nil {
		return nil
	}
	if articleDeptID == nil {
		return errcode.WithMessage(errcode.Forbidden, "权限不足")
	}
	for _, id := range deptIDs {
		if id == *articleDeptID {
			return nil
		}
	}
	return errcode.WithMessage(errcode.Forbidden, "权限不足")
}

// UserMenuTree 当前用户动态菜单树。
func (s *Service) UserMenuTree(ctx context.Context, uid int) (interface{}, error) {
	if uid == 0 {
		return []repo.MenuTreeNode{}, nil
	}
	roles, err := s.roles.ListRolesByUserID(ctx, uid)
	if err != nil {
		return nil, err
	}
	roleIDs := make([]int, 0, len(roles))
	for _, r := range roles {
		roleIDs = append(roleIDs, r.ID)
	}
	menus, err := s.admin.ListMenusByUserID(ctx, roleIDs)
	if err != nil {
		return nil, err
	}
	return repo.BuildMenuTree(menus), nil
}

// CreateMenu 创建菜单。
func (s *Service) CreateMenu(ctx context.Context, m repo.MenuEntity) (interface{}, error) {
	item, err := s.admin.CreateMenu(ctx, m)
	if err != nil {
		if err.Error() == "menu exists" {
			return nil, errcode.WithMessage(errcode.InternalError, "菜单已存在")
		}
		return nil, err
	}
	return item, nil
}

// UpdateMenu 更新菜单。
func (s *Service) UpdateMenu(ctx context.Context, id string, fields map[string]interface{}) (interface{}, error) {
	delete(fields, "id")
	item, err := s.admin.UpdateMenu(ctx, id, fields)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, errcode.WithMessage(errcode.NotFound, "菜单不存在")
	}
	return item, nil
}

// GetMenu 查询菜单详情。
func (s *Service) GetMenu(ctx context.Context, id string) (interface{}, error) {
	item, err := s.admin.GetMenuByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, errcode.WithMessage(errcode.NotFound, "菜单不存在")
	}
	return item, nil
}

// DeleteMenu 删除菜单。
func (s *Service) DeleteMenu(ctx context.Context, id string) (bool, error) {
	ok, err := s.admin.DeleteMenu(ctx, id)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, errcode.WithMessage(errcode.NotFound, "菜单不存在")
	}
	return true, nil
}

// PrivilegeListResult 权限分页列表。
type PrivilegeListResult struct {
	List       []repo.PrivilegeFullEntity `json:"list"`
	Pagination repo.NestPagination        `json:"pagination"`
}

// CreatePrivilege 创建权限。
func (s *Service) CreatePrivilege(ctx context.Context, p repo.PrivilegeFullEntity) (interface{}, error) {
	return s.admin.CreatePrivilege(ctx, p)
}

// ListPrivileges 分页查询权限。
func (s *Service) ListPrivileges(ctx context.Context, page, pageSize int, filters repo.PrivilegeListFilters) (*PrivilegeListResult, error) {
	list, pageInfo, err := s.admin.ListPrivileges(ctx, page, pageSize, filters)
	if err != nil {
		return nil, err
	}
	return &PrivilegeListResult{List: list, Pagination: pageInfo}, nil
}

// GetPrivilege 查询权限详情。
func (s *Service) GetPrivilege(ctx context.Context, id int) (interface{}, error) {
	p, err := s.admin.GetPrivilegeByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, errcode.WithMessage(errcode.NotFound, "权限不存在")
	}
	return p, nil
}

// UpdatePrivilege 更新权限。
func (s *Service) UpdatePrivilege(ctx context.Context, id int, fields map[string]interface{}) (interface{}, error) {
	p, err := s.admin.UpdatePrivilege(ctx, id, fields)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, errcode.WithMessage(errcode.NotFound, "权限不存在")
	}
	return p, nil
}

// DeletePrivilege 删除权限。
func (s *Service) DeletePrivilege(ctx context.Context, id int) (bool, error) {
	ok, err := s.admin.DeletePrivilege(ctx, id)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, errcode.WithMessage(errcode.NotFound, "权限不存在")
	}
	return true, nil
}

// ParseDataScopesFromBody 解析更新数据权限请求体。
func ParseDataScopesFromBody(raw []map[string]interface{}) ([]repo.RoleDataScopeEntity, error) {
	out := make([]repo.RoleDataScopeEntity, 0, len(raw))
	for _, item := range raw {
		s := repo.RoleDataScopeEntity{
			ResourceType: fmt.Sprint(item["resourceType"]),
			ScopeType:    fmt.Sprint(item["scopeType"]),
		}
		if ids, ok := item["deptIds"].([]interface{}); ok {
			for _, v := range ids {
				var id int
				switch n := v.(type) {
				case float64:
					id = int(n)
				case int:
					id = n
				default:
					fmt.Sscan(fmt.Sprint(v), &id)
				}
				s.DeptIDs = append(s.DeptIDs, id)
			}
		}
		out = append(out, s)
	}
	return out, nil
}
