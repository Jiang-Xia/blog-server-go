// Package repo 封装 user 域 Ent 数据访问。
package repo

import (
	"context"

	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent/user"
)

// UserRepo 用户表读写，供 auth/profile 模块使用。
type UserRepo struct {
	client *ent.Client
}

// NewUserRepo 构造 UserRepo。
func NewUserRepo(client *ent.Client) *UserRepo {
	return &UserRepo{client: client}
}

// PasswordCredential 登录校验所需的密码字段。
type PasswordCredential struct {
	UserID   int
	Password string
	Salt     string
	Status   string
}

// GetPasswordByUsername 按用户名查询密码与盐（含软删除过滤）。
func (r *UserRepo) GetPasswordByUsername(ctx context.Context, username string) (*PasswordCredential, error) {
	u, err := r.client.User.Query().
		Where(
			user.UsernameEQ(username),
			user.IsDeleteEQ(false),
		).
		Select(user.FieldID, user.FieldPassword, user.FieldSalt, user.FieldStatus).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return &PasswordCredential{
		UserID:   u.ID,
		Password: u.Password,
		Salt:     u.Salt,
		Status:   u.Status,
	}, nil
}

// UpdatePasswordHash 更新密码哈希；bcrypt 升级后 salt 置空。
func (r *UserRepo) UpdatePasswordHash(ctx context.Context, userID int, passwordHash, salt string) error {
	return r.client.User.UpdateOneID(userID).
		SetPassword(passwordHash).
		SetSalt(salt).
		Exec(ctx)
}

// GetPasswordByUserID 按用户 ID 查询密码与盐（修改密码用）。
func (r *UserRepo) GetPasswordByUserID(ctx context.Context, userID int) (*PasswordCredential, error) {
	u, err := r.client.User.Query().
		Where(
			user.IDEQ(userID),
			user.IsDeleteEQ(false),
		).
		Select(user.FieldID, user.FieldPassword, user.FieldSalt, user.FieldStatus).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return &PasswordCredential{
		UserID:   u.ID,
		Password: u.Password,
		Salt:     u.Salt,
		Status:   u.Status,
	}, nil
}
