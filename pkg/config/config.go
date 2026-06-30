// Package config 通过 Viper 加载 YAML 与环境变量，供 wire 注入全进程使用。
package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
)

// Config 单体应用配置，字段与 configs/monolith.yaml 及 Nest env 语义对齐。
type Config struct {
	App    AppConfig    `mapstructure:"app"`
	HTTP   HTTPConfig   `mapstructure:"http"`
	MySQL  MySQLConfig  `mapstructure:"mysql"`
	Redis  RedisConfig  `mapstructure:"redis"`
	JWT    JWTConfig    `mapstructure:"jwt"`
	Crypto CryptoConfig `mapstructure:"crypto"`
	OAuth  OAuthConfig  `mapstructure:"oauth"`
	Mail   MailConfig    `mapstructure:"mail"`
	Wechat  WechatConfig  `mapstructure:"wechat"`
	Storage StorageConfig `mapstructure:"storage"`
	Pay     PayConfig     `mapstructure:"pay"`
}

// AppConfig 应用级元信息。
type AppConfig struct {
	Name      string `mapstructure:"name"`
	Env       string `mapstructure:"env"`
	APIPrefix string `mapstructure:"api_prefix"`
	BlogHome  string `mapstructure:"blog_home"`
}

// HTTPConfig HTTP 服务监听与 CORS。
type HTTPConfig struct {
	Addr        string   `mapstructure:"addr"`
	CORSOrigins []string `mapstructure:"cors_origins"`
}

// MySQLConfig 数据库连接（password 含特殊字符时用结构化字段，避免 DSN 解析失败）。
type MySQLConfig struct {
	Host        string `mapstructure:"host"`
	Port        int    `mapstructure:"port"`
	User        string `mapstructure:"user"`
	Password    string `mapstructure:"password"`
	Database            string `mapstructure:"database"`
	TablePrefix         string `mapstructure:"table_prefix"`
	SchemaSourceDatabase string `mapstructure:"schema_source_database"`
	DSN                 string `mapstructure:"dsn"`
}

// TablePrefixOrDefault 返回表名前缀，默认 x_（本地库 x_my_blog 约定）。
func (m MySQLConfig) TablePrefixOrDefault() string {
	if p := strings.TrimSpace(m.TablePrefix); p != "" {
		return p
	}
	return "x_"
}

// FormatDSN 返回 go-sql-driver 可用的 DSN；优先结构化字段。
func (m MySQLConfig) FormatDSN() string {
	if m.Host != "" {
		port := m.Port
		if port == 0 {
			port = 3306
		}
		cfg := mysql.Config{
			User:                 m.User,
			Passwd:               m.Password,
			Net:                  "tcp",
			Addr:                 fmt.Sprintf("%s:%d", m.Host, port),
			DBName:               m.Database,
			ParseTime:            true,
			Loc:                  time.Local,
			AllowNativePasswords: true,
			Params: map[string]string{
				"charset": "utf8mb4",
			},
		}
		return cfg.FormatDSN()
	}
	return m.DSN
}

// RedisConfig Redis 连接。
type RedisConfig struct {
	Addr string `mapstructure:"addr"`
	DB   int    `mapstructure:"db"`
}

// JWTConfig 鉴权密钥与 TTL。
type JWTConfig struct {
	Secret     string        `mapstructure:"secret"`
	LegacyTTL  time.Duration `mapstructure:"legacy_ttl"`
	AccessTTL  time.Duration `mapstructure:"access_ttl"`
	RefreshTTL time.Duration `mapstructure:"refresh_ttl"`
}

// OAuthConfig GitHub OAuth 配置。
type OAuthConfig struct {
	GithubClientID     string `mapstructure:"github_client_id"`
	GithubClientSecret string `mapstructure:"github_client_secret"`
	GithubCallbackURL  string `mapstructure:"github_callback_url"`
}

// MailConfig SMTP 邮件发送。
type MailConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
	User string `mapstructure:"user"`
	Pass string `mapstructure:"pass"`
}

// WechatConfig 微信小程序登录。
type WechatConfig struct {
	AppID  string `mapstructure:"app_id"`
	Secret string `mapstructure:"secret"`
}

// PayConfig 支付宝充值配置（env：PAY_ALIPAY_APP_ID 等）。
type PayConfig struct {
	AlipayAppID              string `mapstructure:"alipay_app_id"`
	AlipayPrivateKey         string `mapstructure:"alipay_private_key"`
	AlipayPublicKey          string `mapstructure:"alipay_public_key"`
	AlipayGateway            string `mapstructure:"alipay_gateway"`
	AlipayNotifyURL          string `mapstructure:"alipay_notify_url"`
	AlipayReturnURL          string `mapstructure:"alipay_return_url"`
	AlipayMiniCashierPage    string `mapstructure:"alipay_mini_cashier_page"`
	Sandbox                  bool   `mapstructure:"sandbox"`
	UseLegacySandboxGateway  bool   `mapstructure:"use_legacy_sandbox_gateway"`
	WechatAppID              string `mapstructure:"wechat_app_id"`
	WechatSecret             string `mapstructure:"wechat_secret"`
}

// StorageConfig 文件上传与静态资源路径。
type StorageConfig struct {
	UploadPath   string `mapstructure:"upload_path"`
	PublicPrefix string `mapstructure:"public_prefix"`
}

// PublicPrefixOrDefault 静态 URL 前缀，默认 /static/。
func (s StorageConfig) PublicPrefixOrDefault() string {
	p := strings.TrimSpace(s.PublicPrefix)
	if p == "" {
		return "/static"
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return strings.TrimSuffix(p, "/")
}

// MailConfigured 是否已配置 SMTP。
func (m MailConfig) MailConfigured() bool {
	return m.Host != "" && m.User != "" && m.Pass != ""
}

// IsDev 是否为开发环境。
func (c *Config) IsDev() bool {
	return strings.EqualFold(c.App.Env, "development") || strings.EqualFold(c.App.Env, "dev")
}

// MustLoad 从 path 加载配置；path 为空时使用 CONFIG_PATH 或默认 configs/monolith.yaml。
func MustLoad(path string) (*Config, error) {
	if path == "" {
		path = "configs/monolith.yaml"
	}

	v := viper.New()
	v.SetConfigFile(path)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	return &cfg, nil
}
