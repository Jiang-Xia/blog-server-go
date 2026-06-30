// Package crypto 提供与 Nest cryptogram.util 对齐的 RSA 解密（登录密码传输层）。
// 512 位 RSA 密钥需在 main 或测试中设置 //go:debug rsa1024min=0（与 Nest 开发密钥兼容）。
package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"strings"
)

// RSADecrypt 解密前端 RSA 密文（大写十六进制），与 Nest rsaDecrypt 行为一致。
// 解密失败时返回原文，便于兼容明文调试场景。
func RSADecrypt(encryptedHex, privateKeyPEM string) string {
	if encryptedHex == "" || privateKeyPEM == "" {
		return encryptedHex
	}
	plain, err := rsaDecrypt(encryptedHex, privateKeyPEM)
	if err != nil {
		return encryptedHex
	}
	return plain
}

func rsaDecrypt(encryptedHex, privateKeyPEM string) (string, error) {
	cipherBytes, err := hex.DecodeString(strings.TrimSpace(encryptedHex))
	if err != nil {
		return "", err
	}
	priv, err := parseRSAPrivateKey(privateKeyPEM)
	if err != nil {
		return "", err
	}
	plain, err := rsa.DecryptPKCS1v15(rand.Reader, priv, cipherBytes)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

func parseRSAPrivateKey(pemText string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemText))
	if block == nil {
		return nil, errors.New("invalid PEM block")
	}
	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		if rsaKey, ok := key.(*rsa.PrivateKey); ok {
			return rsaKey, nil
		}
		return nil, errors.New("not RSA private key")
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}
