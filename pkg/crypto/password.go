// Package crypto 提供密码哈希与旧 PBKDF2 格式兼容验证，登录成功后可静默升级为 bcrypt。
package crypto

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/pbkdf2"
)

const bcryptPrefix = "$2"

// MakeSalt 生成 3 字节随机盐（base64），与 Nest makeSalt 一致。
func MakeSalt() (string, error) {
	buf := make([]byte, 3)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buf), nil
}

// Hash 使用 bcrypt 哈希明文密码（新注册用户）。
func Hash(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Verify 校验明文密码；bcrypt 或 legacy PBKDF2（需 salt）。
func Verify(hashed, plain, salt string) bool {
	if hashed == "" || plain == "" {
		return false
	}
	if strings.HasPrefix(hashed, bcryptPrefix) {
		return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain)) == nil
	}
	return legacyPBKDF2Verify(plain, hashed, salt)
}

// NeedsUpgrade 判断存量密码是否需升级为 bcrypt。
func NeedsUpgrade(hashed string) bool {
	return hashed != "" && !strings.HasPrefix(hashed, bcryptPrefix)
}

// UpgradeHash 将明文密码哈希为 bcrypt；升级后 salt 字段应清空（仅存 bcrypt 串）。
func UpgradeHash(plain string) (string, error) {
	return Hash(plain)
}

// legacyPBKDF2Verify 校验 Nest crypto-js / Node PBKDF2 哈希（10000 次 SHA1，16 字节）。
func legacyPBKDF2Verify(password, dbPassword, salt string) bool {
	if salt == "" {
		return false
	}
	saltBytes, err := base64.StdEncoding.DecodeString(salt)
	if err != nil {
		return false
	}
	hash := pbkdf2.Key([]byte(password), saltBytes, 10000, 16, sha1.New)
	return base64.StdEncoding.EncodeToString(hash) == dbPassword
}
