//go:debug rsa1024min=0

package crypto_test

import (
	"testing"

	"github.com/Jiang-Xia/blog-server-go/pkg/crypto"
)

// 与 blog-server/src/config/ssh.ts 开发环境密钥一致，仅用于契约测试。
const testPrivateKey = `-----BEGIN PRIVATE KEY-----
MIIBVgIBADANBgkqhkiG9w0BAQEFAASCAUAwggE8AgEAAkEAv2vyMqR85GmK6cXK
UXhfC82LTqxMPc3iFgsCYY2a+JnUiEKe7hVnSxKF2Psth+H9HDki6pjnldevrNUH
8vNDvQIDAQABAkEAi6NzSv4zHWzgqShgLo5gx3tp5DpMY8mM5Aej9QYXxsEtzq/+
oTPfooVF2rX4rE8NwTpNzwIfzOnrCw5vVCm1AQIhAOZalT7Rx1bqg6irko6MDkVk
9rKW7jebRZ7i3JbonM9DAiEA1Lu9aUWAb98pNRTBnszVzj9FGZKjlSrW/f/PWN2m
8P8CIQCXFFwESoQKDl9xda32jgciHljqwrDUiaL81V/GHiQSjwIgFApzt50ikmd1
nFiOPQWTBtETE2urGXxlsJwOzpJjDcUCIQCV4z96GjcuMYH92dVhmLKFC0ZRX30A
mO+bs1CWyhWG0g==
-----END PRIVATE KEY-----`

const (
	testRSAHelloHex = "A470074926D6007E6046A0F092DEDAD5E1EBB3DDB37354B2E0A65BEA51C960D17718918D26B0504290A56D4E6D3D71C89D3044F224CBEFE1A50CA1D0E8AA3424"
	testRSASampleHex = "B293FD85FE71EC8006DBC9E0EB1D76E1216AA6959257F96903F1FA737EF99F18C787101D62C1FB19CB9B7B2BD206BEB116E1C33E28D71B5FA7B9D47F60BB5838"
)

func TestRSADecryptHello(t *testing.T) {
	got := crypto.RSADecrypt(testRSAHelloHex, testPrivateKey)
	if got != "hello" {
		t.Fatalf("want hello, got %q", got)
	}
}

func TestRSADecryptSample(t *testing.T) {
	got := crypto.RSADecrypt(testRSASampleHex, testPrivateKey)
	if got != "彩票中奖号码:666" {
		t.Fatalf("unexpected plain: %q", got)
	}
}

func TestRSADecryptFallbackOnInvalid(t *testing.T) {
	raw := "not-hex-cipher"
	if crypto.RSADecrypt(raw, testPrivateKey) != raw {
		t.Fatal("invalid cipher should return original text")
	}
}

func TestRSAEncryptDecryptRoundTrip(t *testing.T) {
	const pubKey = `-----BEGIN PUBLIC KEY-----
MFwwDQYJKoZIhvcNAQEBBQADSwAwSAJBAL9r8jKkfORpiunFylF4XwvNi06sTD3N
4hYLAmGNmviZ1IhCnu4VZ0sShdj7LYfh/Rw5IuqY55XXr6zVB/LzQ70CAwEAAQ==
-----END PUBLIC KEY-----`
	cipher, err := crypto.RSAEncrypt("super", pubKey)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if got := crypto.RSADecrypt(cipher, testPrivateKey); got != "super" {
		t.Fatalf("want super, got %q", got)
	}
}
