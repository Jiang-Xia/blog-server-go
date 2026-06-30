package config

// DefaultRSAPrivateKey 与 Nest blog-server/src/config/ssh.ts 开发环境私钥一致。
const DefaultRSAPrivateKey = `-----BEGIN PRIVATE KEY-----
MIIBVgIBADANBgkqhkiG9w0BAQEFAASCAUAwggE8AgEAAkEAv2vyMqR85GmK6cXK
UXhfC82LTqxMPc3iFgsCYY2a+JnUiEKe7hVnSxKF2Psth+H9HDki6pjnldevrNUH
8vNDvQIDAQABAkEAi6NzSv4zHWzgqShgLo5gx3tp5DpMY8mM5Aej9QYXxsEtzq/+
oTPfooVF2rX4rE8NwTpNzwIfzOnrCw5vVCm1AQIhAOZalT7Rx1bqg6irko6MDkVk
9rKW7jebRZ7i3JbonM9DAiEA1Lu9aUWAb98pNRTBnszVzj9FGZKjlSrW/f/PWN2m
8P8CIQCXFFwESoQKDl9xda32jgciHljqwrDUiaL81V/GHiQSjwIgFApzt50ikmd1
nFiOPQWTBtETE2urGXxlsJwOzpJjDcUCIQCV4z96GjcuMYH92dVhmLKFC0ZRX30A
mO+bs1CWyhWG0g==
-----END PRIVATE KEY-----`

// CryptoConfig 传输层加解密配置（RSA 登录密码等）。
type CryptoConfig struct {
	RSAPrivateKey string `mapstructure:"rsa_private_key"`
}

// RSAPrivateKeyOrDefault 返回配置的 RSA 私钥，未配置时用 Nest 开发默认密钥。
func (c *CryptoConfig) RSAPrivateKeyOrDefault() string {
	if c != nil && c.RSAPrivateKey != "" {
		return c.RSAPrivateKey
	}
	return DefaultRSAPrivateKey
}
