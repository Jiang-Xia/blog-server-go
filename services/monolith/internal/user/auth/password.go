// Package auth 实现用户认证相关用例（登录密码校验、静默升级等）。
package auth

import (
	"context"
	"fmt"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/pkg/crypto"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/repo"
)

// PasswordChecker 校验登录密码并在必要时静默升级为 bcrypt。
type PasswordChecker struct {
	repo           *repo.UserRepo
	rsaPrivateKey  string
}

// NewPasswordChecker 构造 PasswordChecker；rsaPrivateKey 与 Nest ssh.privateKey 对齐。
func NewPasswordChecker(userRepo *repo.UserRepo, rsaPrivateKey string) *PasswordChecker {
	if rsaPrivateKey == "" {
		rsaPrivateKey = config.DefaultRSAPrivateKey
	}
	return &PasswordChecker{repo: userRepo, rsaPrivateKey: rsaPrivateKey}
}

// VerifyResult 密码校验结果。
type VerifyResult struct {
	UserID      int
	NeedsUpgrade bool
	PlainPassword string
}

// VerifyLoginPassword 解密传输层密码并校验数据库哈希。
func (p *PasswordChecker) VerifyLoginPassword(ctx context.Context, username, encryptedPassword string) (*VerifyResult, error) {
	cred, err := p.repo.GetPasswordByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	plain := crypto.RSADecrypt(encryptedPassword, p.rsaPrivateKey)
	if !crypto.Verify(cred.Password, plain, cred.Salt) {
		return nil, fmt.Errorf("password mismatch")
	}
	return &VerifyResult{
		UserID:        cred.UserID,
		NeedsUpgrade:  crypto.NeedsUpgrade(cred.Password),
		PlainPassword: plain,
	}, nil
}

// UpgradePasswordIfNeeded 登录成功后静默将 PBKDF2 密码升级为 bcrypt。
func (p *PasswordChecker) UpgradePasswordIfNeeded(ctx context.Context, result *VerifyResult) error {
	if result == nil || !result.NeedsUpgrade {
		return nil
	}
	hash, err := crypto.UpgradeHash(result.PlainPassword)
	if err != nil {
		return err
	}
	return p.repo.UpdatePasswordHash(ctx, result.UserID, hash, "")
}
