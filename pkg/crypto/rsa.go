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

// RSAEncrypt 使用公钥加密明文，返回大写十六进制字符串（与 blog-admin rsaEncrypt 对齐）。
func RSAEncrypt(plain, publicKeyPEM string) (string, error) {
	if plain == "" || publicKeyPEM == "" {
		return "", errors.New("plain or public key empty")
	}
	pub, err := parseRSAPublicKey(publicKeyPEM)
	if err != nil {
		return "", err
	}
	cipher, err := rsa.EncryptPKCS1v15(rand.Reader, pub, []byte(plain))
	if err != nil {
		return "", err
	}
	return strings.ToUpper(hex.EncodeToString(cipher)), nil
}

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

func parseRSAPublicKey(pemText string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemText))
	if block == nil {
		return nil, errors.New("invalid PEM block")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not RSA public key")
	}
	return rsaPub, nil
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
