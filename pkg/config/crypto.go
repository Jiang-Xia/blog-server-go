package config

// DefaultRSAPublicKey 与 Nest blog-server/src/config/ssh.ts 开发环境公钥一致（前端 serverPublicKey）。
const DefaultRSAPublicKey = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAqoyuytldk/BDUvuGUDwW
RplrndYgD0rS+5iP3OOMLH6peR/mjCNkeBQWHmXMfl6KXfwNncYkqtUzANqE7fOv
AoYtiDsnoCAhifhCqIp8+6a0k4WfPztOMSfXyWsTKG85kyMseXE1EWQ+Dy5C0WA8
7adrEr2ceR+jPbsWZA1byQfaL/sHTm9DrcP3QNM9RSVG5tiRQlG8cGSuM3WaKKsA
Fp6wIfkDfKkTXLxQq8C+foFciklA9bCy7gy8k9myT7n+6XIjAIOxeGhcGavB6fAW
wKSr0Q6iopAGJBye+/4pdeuJjdr2MkYR6442lSuBTmNdYKJv/4r3TSGG5u9Inl57
owIDAQAB
-----END PUBLIC KEY-----`

// DefaultRSAPrivateKey 与 Nest blog-server/src/config/ssh.ts 开发环境私钥一致。
const DefaultRSAPrivateKey = `-----BEGIN PRIVATE KEY-----
MIIEvwIBADANBgkqhkiG9w0BAQEFAASCBKkwggSlAgEAAoIBAQCqjK7K2V2T8ENS
+4ZQPBZGmWud1iAPStL7mI/c44wsfql5H+aMI2R4FBYeZcx+Xopd/A2dxiSq1TMA
2oTt868Chi2IOyegICGJ+EKoinz7prSThZ8/O04xJ9fJaxMobzmTIyx5cTURZD4P
LkLRYDztp2sSvZx5H6M9uxZkDVvJB9ov+wdOb0Otw/dA0z1FJUbm2JFCUbxwZK4z
dZooqwAWnrAh+QN8qRNcvFCrwL5+gVyKSUD1sLLuDLyT2bJPuf7pciMAg7F4aFwZ
q8Hp8BbApKvRDqKikAYkHJ77/il164mN2vYyRhHrjjaVK4FOY11gom//ivdNIYbm
70ieXnujAgMBAAECggEAAUKVQ76PQF42JtgFVYckH9mN6RHYcKbFO5TLV1Eay4dd
eeaypt+zNmegKMhWMc/4lCx5pZC0koGxwGtc5shRfjTIA8UEsNxAfFQfLNEvae4A
4x8AjGv692q5TTbjPf3hh2BrOX0M2mWva/u4cjwCS85cN2EEQ3JECbaCXsVNo6mE
4qw6IWMubt1DEG2NRO5rOifDZO0h+M8/jWaw5D1uQd5NoRDCAhf+KH3gbpS2M64b
kHUT0We4qgqtcb0V8fmU5yoqhPY7/9O/K1i+1Tf/3TQMYBLMQTH54TYRK1IKwswy
4RRKgUieGDHGnfR6kKT8kD30wv0MsAVmIiLH8Ec7KQKBgQDZ6ZUB26WRtYiLuGYg
whfmwn6KA1OPlW3Q9mpuZCdHeMEBlb9djI1hTPa5sul2c6yspYAjuuNVQzxUZjOs
Pi74a5uz0KtwcYhcKEyzOYZ8v8q35QB3REo81CA5t7bxTSAjXJyuAnmoGLtkSvEi
WO2mci27uqLDnMWlN+qetK8BawKBgQDIW91FDrNIbwv9KZ2lkLLisIkX6bVOrzFu
47eYo5phfSFuYZL+8yqhTX89FEHGy/ERx9JL1tvUmNjCl1VNEegQaAyaeNIx0DbS
ArZPUls1JIJDAqjnDrIg5pPtzdpXZi6Mc3dftjXM+iDnM8iUEJ9VYVxE5UahnpBx
R9FNs0qkqQKBgQDCf2EXMnQ6dLw2cVT+gh8RQaHuD4AC1U4aSAbBFzWT7mzevEj8
95mtMiO5pb4Jkr+ZrXKM8Z89F/D+J+aRI0OLSYjE20hZ6YvbLcEf6wVlzU3kC0r6
KQLV/kPxIE/yWnOH/H6ANgdeCQ0hPF6z2IntmVXGjJQAYqpqiFkqZH0blwKBgQCF
aNXfA7jRBFAhVDKKad0TIKVIbO+y4uixg6NeWaGURHOcQhBU/qlmGmocy3441tdA
FaOrp/dciiHbAFVmKgQx2HZ77PNZA+Fc2T13d8kPi63I3N6Ga7ykN6RMdgNqqrlY
ABJ+gp80S2DZUhu+q4rMmTWNEnWDbEeVhlKRwE51cQKBgQC+8v0weQzgGUwAVGgD
q7QYsFWM5QYt3DvYw0/F3zx8F9G89wVpSqUrGdpVT56KFJ2JMp+wkZDw06zk8l61
fv0c3XDU6EM8KYAQEPEV5r4FEPhOg3tt/Nf/gs2zCeWZhs+unfrv7JZ7pi7AFbkx
Pply3Dh75GACRALumwbBE13Ptw==
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

// RSAPublicKeyOrDefault 返回 RSA 公钥；当前仅支持 Nest 开发默认密钥对。
func RSAPublicKeyOrDefault() string {
	return DefaultRSAPublicKey
}
