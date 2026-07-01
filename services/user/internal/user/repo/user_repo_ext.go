package repo

import (
	"context"
	"fmt"

	"github.com/Jiang-Xia/blog-server-go/pkg/crypto"
	"github.com/Jiang-Xia/blog-server-go/services/user/ent"
	"github.com/Jiang-Xia/blog-server-go/services/user/ent/dept"
	"github.com/Jiang-Xia/blog-server-go/services/user/ent/user"
)

const (
	defaultRegisterDeptID = 4
	defaultRegisterRoleID = 3
	userStatusActive      = "active"
	userStatusLocked      = "locked"
)

// UserListQuery 用户列表筛选。
type UserListQuery struct {
	Page     int
	PageSize int
	Username string
	Nickname string
}

// UserListItem 列表项，与 Nest user-list.vo 对齐。
type UserListItem struct {
	ID         int    `json:"id"`
	Username   string `json:"username"`
	Nickname   string `json:"nickname"`
	Avatar     string `json:"avatar"`
	Status     string `json:"status"`
	CreateTime string `json:"createTime"`
	UpdateTime string `json:"updateTime"`
	RoleNames  string `json:"roleNames"`
	DeptName   string `json:"deptName"`
	Role       string `json:"role,omitempty"`
}

// Pagination 分页元信息。
type Pagination struct {
	Total       int `json:"total"`
	PageSize    int `json:"pageSize"`
	CurrentPage int `json:"currentPage"`
	TotalPages  int `json:"totalPages"`
}

// FindByID 按 ID 查询用户。
func (r *UserRepo) FindByID(ctx context.Context, id int) (*ent.User, error) {
	return r.client.User.Query().
		Where(user.IDEQ(id), user.IsDeleteEQ(false)).
		Only(ctx)
}

// FindByIDs 批量按 ID 查询用户（未删除）。
func (r *UserRepo) FindByIDs(ctx context.Context, ids []int) ([]*ent.User, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	return r.client.User.Query().
		Where(user.IDIn(ids...), user.IsDeleteEQ(false)).
		All(ctx)
}

// FindByUsername 按用户名查询。
func (r *UserRepo) FindByUsername(ctx context.Context, username string) (*ent.User, error) {
	return r.client.User.Query().
		Where(user.UsernameEQ(username), user.IsDeleteEQ(false)).
		Only(ctx)
}

// FindByEmail 按邮箱查询。
func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*ent.User, error) {
	return r.client.User.Query().
		Where(user.EmailEQ(email), user.IsDeleteEQ(false)).
		Only(ctx)
}

// FindByGithubID 按 GitHub ID 查询。
func (r *UserRepo) FindByGithubID(ctx context.Context, githubID string) (*ent.User, error) {
	return r.client.User.Query().
		Where(user.GithubIdEQ(githubID), user.IsDeleteEQ(false)).
		Only(ctx)
}

// FindByWechatOpenID 按微信 openid 查询。
func (r *UserRepo) FindByWechatOpenID(ctx context.Context, openID string) (*ent.User, error) {
	return r.client.User.Query().
		Where(user.WechatOpenIdEQ(openID), user.IsDeleteEQ(false)).
		Only(ctx)
}

// FindDeptByID 按 ID 查询部门（仅选 Nest 表存在的列）。
func (r *UserRepo) FindDeptByID(ctx context.Context, id int) (*ent.Dept, error) {
	return r.client.Dept.Query().
		Where(dept.IDEQ(id)).
		Select(dept.FieldID, dept.FieldDeptName).
		Only(ctx)
}

// CreateUserInput 创建用户参数。
type CreateUserInput struct {
	Username string
	Nickname string
	Password string
	Email    string
	Avatar   string
	GithubID string
	WechatID string
	Homepage string
	DeptID   int
}

// Create 创建用户（密码 bcrypt）。
func (r *UserRepo) Create(ctx context.Context, in CreateUserInput) (*ent.User, error) {
	hash, err := crypto.Hash(in.Password)
	if err != nil {
		return nil, err
	}
	b := r.client.User.Create().
		SetNickname(in.Nickname).
		SetPassword(hash).
		SetSalt("").
		SetStatus(userStatusActive)
	if in.Username != "" {
		b.SetUsername(in.Username)
	}
	if in.Email != "" {
		b.SetEmail(in.Email)
	}
	if in.Avatar != "" {
		b.SetAvatar(in.Avatar)
	}
	if in.GithubID != "" {
		b.SetGithubId(in.GithubID)
	}
	if in.WechatID != "" {
		b.SetWechatOpenId(in.WechatID)
	}
	if in.Homepage != "" {
		b.SetHomepage(in.Homepage)
	}
	deptID := in.DeptID
	if deptID == 0 {
		deptID = defaultRegisterDeptID
	}
	b.SetDeptId(deptID)
	return b.Save(ctx)
}

// UpdateFields 按 ID 合并更新字段。
func (r *UserRepo) UpdateFields(ctx context.Context, id int, fields map[string]interface{}) (*ent.User, error) {
	u, err := r.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	up := r.client.User.UpdateOneID(u.ID)
	for k, v := range fields {
		switch k {
		case "nickname":
			up.SetNickname(v.(string))
		case "avatar":
			up.SetAvatar(v.(string))
		case "intro":
			up.SetIntro(v.(string))
		case "homepage":
			up.SetHomepage(v.(string))
		case "email":
			up.SetEmail(v.(string))
		case "status":
			up.SetStatus(v.(string))
		case "username":
			up.SetUsername(v.(string))
		case "deptId":
			switch d := v.(type) {
			case int:
				up.SetDeptId(d)
			case float64:
				up.SetDeptId(int(d))
			}
		}
	}
	return up.Save(ctx)
}

// SoftDelete 软删除用户。
func (r *UserRepo) SoftDelete(ctx context.Context, id int) error {
	return r.client.User.UpdateOneID(id).SetIsDelete(true).Exec(ctx)
}

// CountAll 统计未删除用户数。
func (r *UserRepo) CountAll(ctx context.Context) (int, error) {
	return r.client.User.Query().Where(user.IsDeleteEQ(false)).Count(ctx)
}

// List 分页查询用户列表。
func (r *UserRepo) List(ctx context.Context, q UserListQuery, roleRepo *RoleRepo) ([]UserListItem, Pagination, error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 {
		q.PageSize = 20
	}
	query := r.client.User.Query().Where(user.IsDeleteEQ(false))
	if q.Username != "" {
		query = query.Where(user.UsernameContains(q.Username))
	}
	if q.Nickname != "" {
		query = query.Where(user.NicknameContains(q.Nickname))
	}
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, Pagination{}, err
	}
	rows, err := query.
		Order(ent.Asc(user.FieldCreateTime)).
		Offset((q.Page - 1) * q.PageSize).
		Limit(q.PageSize).
		All(ctx)
	if err != nil {
		return nil, Pagination{}, err
	}
	list := make([]UserListItem, 0, len(rows))
	for _, u := range rows {
		item := UserListItem{
			ID:         u.ID,
			Nickname:   u.Nickname,
			Avatar:     u.Avatar,
			Status:     u.Status,
			CreateTime: u.CreateTime.Format("2006-01-02T15:04:05.000Z"),
			UpdateTime: u.UpdateTime.Format("2006-01-02T15:04:05.000Z"),
		}
		if u.Username != nil {
			item.Username = *u.Username
		}
		roles, _ := roleRepo.ListRolesByUserID(ctx, u.ID)
		names := make([]string, 0, len(roles))
		for _, role := range roles {
			names = append(names, role.RoleName)
			if role.ID == 1 {
				item.Role = "super"
			}
		}
		if len(names) > 0 {
			item.RoleNames = joinChinese(names)
		}
		if u.DeptId != nil {
			d, err := r.client.Dept.Query().Where(dept.IDEQ(*u.DeptId)).Only(ctx)
			if err == nil {
				item.DeptName = d.DeptName
			}
		}
		list = append(list, item)
	}
	totalPages := total / q.PageSize
	if total%q.PageSize != 0 {
		totalPages++
	}
	return list, Pagination{
		Total: total, PageSize: q.PageSize, CurrentPage: q.Page, TotalPages: totalPages,
	}, nil
}

func joinChinese(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	out := parts[0]
	for i := 1; i < len(parts); i++ {
		out += "、" + parts[i]
	}
	return out
}

// DefaultRegisterRoleID 自助注册默认角色。
func DefaultRegisterRoleID() int { return defaultRegisterRoleID }

// AssertUserActive 用户未锁定。
func AssertUserActive(status string) error {
	if status == userStatusLocked {
		return fmt.Errorf("账号已被锁定")
	}
	return nil
}

// ListActiveUserIDs 返回未删除且 active 的用户 ID 列表，供 C 端文章列表过滤（避免跨表 JOIN）。
func (r *UserRepo) ListActiveUserIDs(ctx context.Context) ([]int, error) {
	rows, err := r.client.User.Query().
		Where(user.IsDeleteEQ(false), user.StatusEQ(userStatusActive)).
		Select(user.FieldID).
		All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]int, 0, len(rows))
	for _, u := range rows {
		out = append(out, u.ID)
	}
	return out, nil
}
