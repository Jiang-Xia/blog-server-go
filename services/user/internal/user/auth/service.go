// Package auth 实现登录、注册、刷新 token 与 OAuth 等认证用例。
package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/pkg/crypto"
	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/pkg/redisutil"
	"github.com/Jiang-Xia/blog-server-go/services/user/ent"
	"github.com/Jiang-Xia/blog-server-go/services/user/internal/user/email"
	"github.com/Jiang-Xia/blog-server-go/services/user/internal/user/repo"
	"github.com/Jiang-Xia/blog-server-go/services/user/internal/user/vo"
	"github.com/google/uuid"
)

// AuthService 认证服务，对齐 Nest UserService 中 auth 相关方法。
type AuthService struct {
	users     *repo.UserRepo
	roles     *repo.RoleRepo
	passwords *PasswordChecker
	jwt       *JWTService
	redis     *redisutil.Store
	cfg       *config.Config
	email     *email.Service
}

// NewAuthService 构造 AuthService。
func NewAuthService(
	users *repo.UserRepo,
	roles *repo.RoleRepo,
	passwords *PasswordChecker,
	jwt *JWTService,
	redis *redisutil.Store,
	cfg *config.Config,
	emailSvc *email.Service,
) *AuthService {
	return &AuthService{
		users:     users,
		roles:     roles,
		passwords: passwords,
		jwt:       jwt,
		redis:     redis,
		cfg:       cfg,
		email:     emailSvc,
	}
}

// LoginInput 账号密码登录参数（password 为 RSA 密文）。
type LoginInput struct {
	Username string
	Password string
}

// RegisterInput 自助注册参数。
type RegisterInput struct {
	Username       string
	Nickname       string
	Password       string
	PasswordRepeat string
	Avatar         string
}

// SendEmailCodeInput 发送邮箱验证码参数。
type SendEmailCodeInput struct {
	Email string
	Type  string // register | login | reset
}

// UpdatePasswordInput 自助修改密码参数。
type UpdatePasswordInput struct {
	ID             int
	PasswordOld    string
	Password       string
	PasswordRepeat string
}

// ResetPasswordInput 按用户名+昵称重置密码参数。
type ResetPasswordInput struct {
	Username string
	Nickname string
}

// LoginResult 登录响应 info 块。
type LoginResult struct {
	Token        string                 `json:"token"`
	AccessToken  string                 `json:"accessToken"`
	RefreshToken string                 `json:"refreshToken"`
	User         map[string]interface{} `json:"user"`
}

// Login 校验凭证、登录频控、签发 token。
func (s *AuthService) Login(ctx context.Context, in LoginInput, clientIP string) (*LoginResult, error) {
	if err := s.assertLoginAllowed(ctx, in.Username, clientIP); err != nil {
		return nil, err
	}
	cred, err := s.users.GetPasswordByUsername(ctx, in.Username)
	if err != nil {
		if ent.IsNotFound(err) {
			_ = s.recordLoginFailure(ctx, in.Username, clientIP)
			return nil, errcode.WithMessage(errcode.NotFound, "账号不存在")
		}
		return nil, err
	}
	if err := repo.AssertUserActive(cred.Status); err != nil {
		return nil, errcode.WithMessage(errcode.Unauthorized, "%s", err.Error())
	}
	verify, err := s.passwords.VerifyLoginPassword(ctx, in.Username, in.Password)
	if err != nil || verify == nil {
		_ = s.recordLoginFailure(ctx, in.Username, clientIP)
		return nil, errcode.WithMessage(errcode.NotFound, "密码错误")
	}
	if err := s.passwords.UpgradePasswordIfNeeded(ctx, verify); err != nil {
		return nil, err
	}
	_ = s.clearLoginFailure(ctx, in.Username, clientIP)
	u, err := s.users.FindByID(ctx, verify.UserID)
	if err != nil {
		return nil, err
	}
	return s.buildLoginResult(ctx, u)
}

// Register 自助注册：校验表单、创建用户、绑定默认角色。
func (s *AuthService) Register(ctx context.Context, in RegisterInput, init bool) (map[string]interface{}, error) {
	if in.Password != in.PasswordRepeat {
		return nil, errcode.WithMessage(errcode.NotFound, "两次输入的密码不一致，请检查")
	}
	if _, err := s.users.FindByUsername(ctx, in.Username); err == nil {
		return nil, errcode.WithMessage(errcode.NotFound, "账号已存在")
	} else if !ent.IsNotFound(err) {
		return nil, err
	}
	u, err := s.users.Create(ctx, repo.CreateUserInput{
		Username: in.Username,
		Nickname: in.Nickname,
		Password: in.Password,
		Avatar:   in.Avatar,
	})
	if err != nil {
		return nil, err
	}
	if err := s.roles.BindRole(ctx, u.ID, repo.DefaultRegisterRoleID()); err != nil {
		return nil, err
	}
	_ = init // 初始化管理员账户时不发注册事件（Plan 05 事件总线接入后再补）
	return vo.SanitizeUser(u), nil
}

// Refresh 用 refresh token 轮换 access/refresh，旧 refresh 写入黑名单。
func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (map[string]interface{}, error) {
	hash := sha256.Sum256([]byte(refreshToken))
	key := refreshBlacklistKey(hex.EncodeToString(hash[:]))
	used, err := s.redis.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	if used != "" {
		return nil, errcode.WithMessage(errcode.Unauthorized, "refresh token 已失效，请重新登录")
	}
	claims, err := s.jwt.Verify(refreshToken)
	if err != nil {
		return nil, errcode.WithMessage(errcode.Unauthorized, "token 失效，请重新登录")
	}
	ttl := s.jwt.RemainingTTL(claims)
	if err := s.redis.Set(ctx, key, "1", ttl); err != nil {
		return nil, err
	}
	u, err := s.users.FindByID(ctx, claims.ID)
	if err != nil {
		return nil, errcode.WithMessage(errcode.Unauthorized, "token 失效，请重新登录")
	}
	tokens, err := s.Certificate(ctx, u)
	if err != nil {
		return nil, err
	}
	roles, _ := s.roles.ListRolesByUserID(ctx, u.ID)
	var dept *ent.Dept
	if u.DeptId != nil {
		dept, _ = s.users.FindDeptByID(ctx, *u.DeptId)
	}
	out := map[string]interface{}{
		"token":        tokens.Token,
		"accessToken":  tokens.AccessToken,
		"refreshToken": tokens.RefreshToken,
		"user":         vo.UserWithRoles(u, roles, dept),
		"message":      "刷新token成功",
	}
	return out, nil
}

// Certificate 签发 JWT 三 token，payload 与 Nest certificate() 一致。
func (s *AuthService) Certificate(ctx context.Context, u *ent.User) (*TokenTriple, error) {
	if u == nil {
		return nil, errcode.NotFound
	}
	roles, err := s.roles.ListRolesByUserID(ctx, u.ID)
	if err != nil {
		return nil, err
	}
	payloads := vo.RolePayloadsForJWT(roles)
	jwtRoles := make([]RolePayload, 0, len(payloads))
	for _, r := range payloads {
		jwtRoles = append(jwtRoles, RolePayload{ID: r.ID, RoleName: r.RoleName, RoleDesc: r.RoleDesc})
	}
	username := ""
	if u.Username != nil {
		username = *u.Username
	}
	return s.jwt.SignTriple(u.ID, u.Nickname, username, jwtRoles)
}

// CreateOAuthTicket 创建 OAuth 登录一次性 ticket，避免 token 出现在 URL。
func (s *AuthService) CreateOAuthTicket(ctx context.Context, tokens *TokenTriple) (string, error) {
	if tokens == nil {
		return "", errcode.InvalidParam
	}
	ticket := uuid.NewString()
	raw, err := json.Marshal(tokens)
	if err != nil {
		return "", err
	}
	if err := s.redis.Set(ctx, oauthTicketKey(ticket), string(raw), oauthTicketTTL); err != nil {
		return "", err
	}
	return ticket, nil
}

// ExchangeOAuthTicket 用一次性 ticket 换取 token（单次有效）。
func (s *AuthService) ExchangeOAuthTicket(ctx context.Context, ticket string) (*TokenTriple, error) {
	key := oauthTicketKey(ticket)
	stored, err := s.redis.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	if stored == "" {
		return nil, errcode.WithMessage(errcode.InvalidParam, "登录凭证无效或已过期")
	}
	_ = s.redis.Del(ctx, key)
	var tokens TokenTriple
	if err := json.Unmarshal([]byte(stored), &tokens); err != nil {
		return nil, err
	}
	return &tokens, nil
}

// UpdatePassword 用户自助修改密码：校验旧密码后重新哈希。
func (s *AuthService) UpdatePassword(ctx context.Context, in UpdatePasswordInput) error {
	if in.Password != in.PasswordRepeat {
		return errcode.WithMessage(errcode.NotFound, "两次输入的密码不一致，请检查")
	}
	cred, err := s.users.GetPasswordByUserID(ctx, in.ID)
	if err != nil {
		if ent.IsNotFound(err) {
			return errcode.WithMessage(errcode.NotFound, "用户不存在")
		}
		return err
	}
	if !crypto.Verify(cred.Password, in.PasswordOld, cred.Salt) {
		return errcode.WithMessage(errcode.NotFound, "旧密码不不正确，请检查！")
	}
	hash, err := crypto.Hash(in.Password)
	if err != nil {
		return err
	}
	return s.users.UpdatePasswordHash(ctx, in.ID, hash, "")
}

// ResetPassword 按用户名+昵称重置密码为默认 123456。
func (s *AuthService) ResetPassword(ctx context.Context, in ResetPasswordInput) (map[string]string, error) {
	u, err := s.users.FindByUsername(ctx, in.Username)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errcode.WithMessage(errcode.NotFound, "此用户不存在！")
		}
		return nil, err
	}
	if u.Nickname != in.Nickname {
		return nil, errcode.WithMessage(errcode.NotFound, "此用户不存在！")
	}
	ok, err := s.roles.IsSuperAdmin(ctx, u.ID)
	if err != nil {
		return nil, err
	}
	if ok {
		return nil, errcode.WithMessage(errcode.Forbidden, "不能对超级管理员执行此操作")
	}
	hash, err := crypto.Hash("123456")
	if err != nil {
		return nil, err
	}
	if err := s.users.UpdatePasswordHash(ctx, u.ID, hash, ""); err != nil {
		return nil, err
	}
	return map[string]string{"message": "重置密码成功，默认密码为：123456"}, nil
}

// SendEmailCode 发送邮箱验证码（含业务校验，委托 EmailService 发信）。
func (s *AuthService) SendEmailCode(ctx context.Context, in SendEmailCodeInput) (map[string]string, error) {
	if err := s.email.CheckSendFrequency(ctx, in.Email, in.Type); err != nil {
		return nil, err
	}
	switch in.Type {
	case "register":
		if u, err := s.users.FindByEmail(ctx, in.Email); err == nil && u != nil {
			return nil, errcode.WithMessage(errcode.InvalidParam, "该邮箱已被注册")
		} else if err != nil && !ent.IsNotFound(err) {
			return nil, err
		}
		if u, err := s.users.FindByUsername(ctx, in.Email); err == nil && u != nil {
			return nil, errcode.WithMessage(errcode.InvalidParam, "该邮箱已被注册")
		} else if err != nil && !ent.IsNotFound(err) {
			return nil, err
		}
	case "login", "reset":
		if _, err := s.users.FindByEmail(ctx, in.Email); err != nil {
			if ent.IsNotFound(err) {
				return nil, errcode.WithMessage(errcode.InvalidParam, "该邮箱尚未注册")
			}
			return nil, err
		}
	}
	if err := s.email.SendCode(ctx, in.Email, in.Type); err != nil {
		return nil, err
	}
	return map[string]string{"message": "验证码发送成功"}, nil
}

func (s *AuthService) buildLoginResult(ctx context.Context, u *ent.User) (*LoginResult, error) {
	tokens, err := s.Certificate(ctx, u)
	if err != nil {
		return nil, err
	}
	roles, err := s.roles.ListRolesByUserID(ctx, u.ID)
	if err != nil {
		return nil, err
	}
	var dept *ent.Dept
	if u.DeptId != nil {
		dept, _ = s.users.FindDeptByID(ctx, *u.DeptId)
	}
	return &LoginResult{
		Token:        tokens.Token,
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		User:         vo.UserWithRoles(u, roles, dept),
	}, nil
}

func (s *AuthService) assertLoginAllowed(ctx context.Context, username, clientIP string) error {
	locked, err := s.redis.Get(ctx, loginLockKey(username, clientIP))
	if err != nil {
		return err
	}
	if locked != "" {
		return errcode.WithMessage(errcode.Unauthorized, "登录失败次数过多，请稍后再试")
	}
	return nil
}

func (s *AuthService) recordLoginFailure(ctx context.Context, username, clientIP string) error {
	failKey := loginFailKey(username, clientIP)
	n, err := s.redis.Incr(ctx, failKey)
	if err != nil {
		return err
	}
	if n == 1 {
		_ = s.redis.Expire(ctx, failKey, loginFailWindowSec)
	}
	if n >= loginFailMaxCount {
		return s.redis.Set(ctx, loginLockKey(username, clientIP), "1", loginLockSec)
	}
	return nil
}

func (s *AuthService) clearLoginFailure(ctx context.Context, username, clientIP string) error {
	return s.redis.Del(ctx, loginFailKey(username, clientIP), loginLockKey(username, clientIP))
}

// LoginAfterOAuth OAuth/GitHub/微信登录成功后签发 token 的便捷方法。
func (s *AuthService) LoginAfterOAuth(ctx context.Context, u *ent.User) (*LoginResult, error) {
	if err := repo.AssertUserActive(u.Status); err != nil {
		return nil, errcode.WithMessage(errcode.Unauthorized, "%s", err.Error())
	}
	return s.buildLoginResult(ctx, u)
}

// CreateUserWithDefaultRole 创建用户并绑定默认作者角色（OAuth 新用户）。
func (s *AuthService) CreateUserWithDefaultRole(ctx context.Context, in repo.CreateUserInput) (*ent.User, error) {
	u, err := s.users.Create(ctx, in)
	if err != nil {
		return nil, err
	}
	if err := s.roles.BindRole(ctx, u.ID, repo.DefaultRegisterRoleID()); err != nil {
		return nil, err
	}
	return u, nil
}

// RandomPassword 生成随机明文密码（微信/GitHub 新用户占位）。
func RandomPassword() string {
	return uuid.NewString()[:12]
}

// EmailRegisterInput 邮箱注册参数。
type EmailRegisterInput struct {
	Email            string
	Nickname         string
	Password         string
	PasswordRepeat   string
	VerificationCode string
	Avatar           string
}

// EmailLoginInput 邮箱验证码登录参数。
type EmailLoginInput struct {
	Email            string
	VerificationCode string
}

// EmailRegister 邮箱验证码注册。
func (s *AuthService) EmailRegister(ctx context.Context, in EmailRegisterInput) (map[string]interface{}, error) {
	if in.Password != in.PasswordRepeat {
		return nil, errcode.WithMessage(errcode.NotFound, "两次输入的密码不一致，请检查")
	}
	if err := s.email.VerifyCode(ctx, in.Email, in.VerificationCode, "register"); err != nil {
		return nil, err
	}
	if _, err := s.users.FindByEmail(ctx, in.Email); err == nil {
		return nil, errcode.WithMessage(errcode.InvalidParam, "该邮箱已被注册")
	} else if !ent.IsNotFound(err) {
		return nil, err
	}
	u, err := s.users.Create(ctx, repo.CreateUserInput{
		Email: in.Email, Nickname: in.Nickname, Password: in.Password, Avatar: in.Avatar,
		Username: in.Email,
	})
	if err != nil {
		return nil, err
	}
	if err := s.roles.BindRole(ctx, u.ID, repo.DefaultRegisterRoleID()); err != nil {
		return nil, err
	}
	return vo.SanitizeUser(u), nil
}

// EmailLogin 邮箱验证码登录。
func (s *AuthService) EmailLogin(ctx context.Context, in EmailLoginInput) (*LoginResult, error) {
	if err := s.email.VerifyCode(ctx, in.Email, in.VerificationCode, "login"); err != nil {
		return nil, err
	}
	u, err := s.users.FindByEmail(ctx, in.Email)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errcode.WithMessage(errcode.InvalidParam, "该邮箱尚未注册")
		}
		return nil, err
	}
	return s.LoginAfterOAuth(ctx, u)
}

// MustHash 哈希密码，失败 panic（仅 OAuth 快捷路径）。
func MustHash(plain string) string {
	h, err := crypto.Hash(plain)
	if err != nil {
		panic(fmt.Sprintf("hash password: %v", err))
	}
	return h
}
