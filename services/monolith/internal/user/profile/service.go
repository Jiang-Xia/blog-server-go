// Package profile 用户资料 CRUD 与权限断言，对齐 Nest UserService 资料部分。
package profile

import (
	"context"
	"fmt"

	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/repo"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/vo"
)

// Service 用户资料服务。
type Service struct {
	users *repo.UserRepo
	roles *repo.RoleRepo
}

// NewService 构造 ProfileService。
func NewService(users *repo.UserRepo, roles *repo.RoleRepo) *Service {
	return &Service{users: users, roles: roles}
}

// UserListResult 用户列表响应。
type UserListResult struct {
	List       []repo.UserListItem `json:"list"`
	Pagination repo.Pagination     `json:"pagination"`
}

// AdminCreateInput 管理员创建用户参数。
type AdminCreateInput struct {
	Nickname string
	Username string
	Password string
	RoleIDs  []int
	DeptID   int
	Intro    string
	Avatar   string
}

// AdminUpdateInput 管理员更新用户参数。
type AdminUpdateInput struct {
	Nickname string
	RoleIDs  []int
	DeptID   int
	Intro    string
	Avatar   string
}

// GetUserRolePrivilegeInfo 查询用户详情（含 roles.privileges、dept）。
func (s *Service) GetUserRolePrivilegeInfo(ctx context.Context, userID int) (map[string]interface{}, error) {
	u, err := s.users.FindByID(ctx, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errcode.WithMessage(errcode.NotFound, "用户不存在")
		}
		return nil, err
	}
	roles, err := s.roles.ListRolesWithPrivileges(ctx, userID)
	if err != nil {
		return nil, err
	}
	var dept *ent.Dept
	if u.DeptId != nil {
		dept, _ = s.users.FindDeptByID(ctx, *u.DeptId)
	}
	return vo.UserWithRolesAndPrivileges(u, roles, dept), nil
}

// GetUserInfo 查询用户详情（含 roles、dept），与 Nest findById 对齐。
func (s *Service) GetUserInfo(ctx context.Context, userID int) (map[string]interface{}, error) {
	u, err := s.users.FindByID(ctx, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errcode.WithMessage(errcode.NotFound, "用户不存在")
		}
		return nil, err
	}
	roles, err := s.roles.ListRolesByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	var dept *ent.Dept
	if u.DeptId != nil {
		dept, _ = s.users.FindDeptByID(ctx, *u.DeptId)
	}
	return vo.UserWithRoles(u, roles, dept), nil
}

// ListUsers 分页用户列表。
func (s *Service) ListUsers(ctx context.Context, q repo.UserListQuery) (*UserListResult, error) {
	list, page, err := s.users.List(ctx, q, s.roles)
	if err != nil {
		return nil, err
	}
	return &UserListResult{List: list, Pagination: page}, nil
}

// UpdateField 合并更新用户字段；status 变更时禁止操作超级管理员。
func (s *Service) UpdateField(ctx context.Context, fields map[string]interface{}) (map[string]interface{}, error) {
	idVal, ok := fields["id"]
	if !ok {
		return nil, errcode.InvalidParam
	}
	id, ok := toInt(idVal)
	if !ok || id <= 0 {
		return nil, errcode.InvalidParam
	}
	delete(fields, "id")

	if status, ok := fields["status"]; ok && status != nil {
		if err := s.assertNotSuperAdminTarget(ctx, id); err != nil {
			return nil, err
		}
	}

	u, err := s.users.UpdateFields(ctx, id, fields)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errcode.WithMessage(errcode.NotFound, "用户不存在")
		}
		return nil, err
	}
	return vo.SanitizeUser(u), nil
}

// DeleteUser 软删除用户，禁止删除超级管理员。
func (s *Service) DeleteUser(ctx context.Context, id int) error {
	if id <= 0 {
		return errcode.WithMessage(errcode.NotFound, "用户不存在")
	}
	if _, err := s.users.FindByID(ctx, id); err != nil {
		if ent.IsNotFound(err) {
			return errcode.WithMessage(errcode.NotFound, "用户不存在")
		}
		return err
	}
	if err := s.assertNotSuperAdminTarget(ctx, id); err != nil {
		return err
	}
	return s.users.SoftDelete(ctx, id)
}

// AdminCreate 管理员创建用户并绑定角色。
func (s *Service) AdminCreate(ctx context.Context, in AdminCreateInput) (map[string]interface{}, error) {
	if in.Username == "" {
		return nil, errcode.InvalidParam
	}
	if _, err := s.users.FindByUsername(ctx, in.Username); err == nil {
		return nil, errcode.WithMessage(errcode.InvalidParam, "该账号已被注册")
	} else if !ent.IsNotFound(err) {
		return nil, err
	}
	if len(in.RoleIDs) > 0 {
		if err := s.ensureRolesExist(ctx, in.RoleIDs); err != nil {
			return nil, err
		}
	}
	deptID := in.DeptID
	u, err := s.users.Create(ctx, repo.CreateUserInput{
		Username: in.Username,
		Nickname: in.Nickname,
		Password: in.Password,
		Avatar:   in.Avatar,
		DeptID:   deptID,
	})
	if err != nil {
		return nil, err
	}
	if in.Intro != "" {
		u, err = s.users.UpdateFields(ctx, u.ID, map[string]interface{}{"intro": in.Intro})
		if err != nil {
			return nil, err
		}
	}
	for _, rid := range in.RoleIDs {
		if err := s.roles.BindRole(ctx, u.ID, rid); err != nil {
			return nil, err
		}
	}
	return s.GetUserInfo(ctx, u.ID)
}

// AdminUpdate 管理员更新用户资料与角色。
func (s *Service) AdminUpdate(ctx context.Context, userID int, in AdminUpdateInput) (map[string]interface{}, error) {
	u, err := s.users.FindByID(ctx, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errcode.WithMessage(errcode.NotFound, "用户不存在")
		}
		return nil, err
	}
	fields := map[string]interface{}{}
	if in.Nickname != "" {
		fields["nickname"] = in.Nickname
	}
	if in.Intro != "" {
		fields["intro"] = in.Intro
	}
	if in.Avatar != "" {
		fields["avatar"] = in.Avatar
	}
	if in.DeptID > 0 {
		fields["deptId"] = in.DeptID
	}
	if len(fields) > 0 {
		if _, err := s.users.UpdateFields(ctx, u.ID, fields); err != nil {
			return nil, err
		}
	}
	if in.RoleIDs != nil {
		if len(in.RoleIDs) > 0 {
			if err := s.ensureRolesExist(ctx, in.RoleIDs); err != nil {
				return nil, err
			}
		}
		if err := s.roles.ReplaceUserRoles(ctx, userID, in.RoleIDs); err != nil {
			return nil, err
		}
	}
	return s.GetUserInfo(ctx, userID)
}

// ResolveProfileUID 解析 /user/info 查询目标：默认本人；?id= 仅超管可查他人。
func (s *Service) ResolveProfileUID(ctx context.Context, operatorUID int, requestedUserID *int) (int, error) {
	if requestedUserID == nil || *requestedUserID <= 0 || *requestedUserID == operatorUID {
		return operatorUID, nil
	}
	ok, err := s.roles.IsSuperAdmin(ctx, operatorUID)
	if err != nil {
		return 0, err
	}
	if !ok {
		return 0, errcode.WithMessage(errcode.Forbidden, "无权查看其他用户信息")
	}
	return *requestedUserID, nil
}

// AssertSuperAdmin 仅超级管理员可操作。
func (s *Service) AssertSuperAdmin(ctx context.Context, operatorUID int) error {
	ok, err := s.roles.IsSuperAdmin(ctx, operatorUID)
	if err != nil {
		return err
	}
	if !ok {
		return errcode.WithMessage(errcode.Forbidden, "仅超级管理员可操作")
	}
	return nil
}

// AssertSelfOrSuperAdmin 仅允许操作本人资料，或超级管理员操作任意用户。
func (s *Service) AssertSelfOrSuperAdmin(ctx context.Context, operatorUID, targetUserID int) error {
	if operatorUID == targetUserID {
		return nil
	}
	ok, err := s.roles.IsSuperAdmin(ctx, operatorUID)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	return errcode.WithMessage(errcode.Forbidden, "无权修改其他用户资料")
}

// GetUserDTO 供 UserService 端口使用的用户摘要。
func (s *Service) GetUserDTO(ctx context.Context, id uint64) (*UserDTO, error) {
	u, err := s.users.FindByID(ctx, int(id))
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errcode.WithMessage(errcode.NotFound, "用户不存在")
		}
		return nil, err
	}
	return entUserToDTO(u), nil
}

// GetUserDTOBatch 批量查询用户摘要。
func (s *Service) GetUserDTOBatch(ctx context.Context, ids []uint64) ([]*UserDTO, error) {
	intIDs := make([]int, 0, len(ids))
	for _, id := range ids {
		intIDs = append(intIDs, int(id))
	}
	rows, err := s.users.FindByIDs(ctx, intIDs)
	if err != nil {
		return nil, err
	}
	byID := make(map[int]*ent.User, len(rows))
	for _, u := range rows {
		byID[u.ID] = u
	}
	out := make([]*UserDTO, 0, len(ids))
	for _, id := range ids {
		if u, ok := byID[int(id)]; ok {
			out = append(out, entUserToDTO(u))
		}
	}
	return out, nil
}

// UserDTO 跨模块用户摘要（与 user 包 port 结构一致，避免循环依赖时由 service 层转换）。
type UserDTO struct {
	ID       uint64 `json:"id"`
	Nickname string `json:"nickname"`
	Username string `json:"username,omitempty"`
	Avatar   string `json:"avatar"`
	Email    string `json:"email,omitempty"`
	Status   string `json:"status,omitempty"`
}

func entUserToDTO(u *ent.User) *UserDTO {
	if u == nil {
		return nil
	}
	dto := &UserDTO{
		ID:       uint64(u.ID),
		Nickname: u.Nickname,
		Avatar:   u.Avatar,
		Status:   u.Status,
	}
	if u.Username != nil {
		dto.Username = *u.Username
	}
	if u.Email != nil {
		dto.Email = *u.Email
	}
	return dto
}

func (s *Service) assertNotSuperAdminTarget(ctx context.Context, userID int) error {
	ok, err := s.roles.IsSuperAdmin(ctx, userID)
	if err != nil {
		return err
	}
	if ok {
		return errcode.WithMessage(errcode.Forbidden, "不能对超级管理员执行此操作")
	}
	return nil
}

func (s *Service) ensureRolesExist(ctx context.Context, roleIDs []int) error {
	ok, err := s.roles.RolesExist(ctx, roleIDs)
	if err != nil {
		return err
	}
	if !ok {
		return errcode.WithMessage(errcode.InvalidParam, "部分角色不存在")
	}
	return nil
}

func toInt(v interface{}) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case int64:
		return int(n), true
	case float64:
		return int(n), true
	case fmt.Stringer:
		var i int
		_, err := fmt.Sscan(n.String(), &i)
		return i, err == nil
	default:
		return 0, false
	}
}
