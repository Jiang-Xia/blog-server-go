// Package config 通过 Viper 加载 YAML 与环境变量，供 wire 注入全进程使用。
package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
)

// ServiceMode 运行形态：monolith 单体；user/blog/rpg 微服务；gateway 仅 BFF/代理。
type ServiceMode string

const (
	ModeMonolith ServiceMode = "monolith"
	ModeGateway  ServiceMode = "gateway"
	ModeUser     ServiceMode = "user"
	ModeBlog     ServiceMode = "blog"
	ModeRPG      ServiceMode = "rpg"
)

// Config 应用配置，字段与 configs/*.yaml 及 Nest env 语义对齐。
type Config struct {
	App      AppConfig      `mapstructure:"app"`
	HTTP     HTTPConfig     `mapstructure:"http"`
	Kitex    KitexConfig    `mapstructure:"kitex"`
	Registry RegistryConfig `mapstructure:"registry"`
	Proxy    ProxyConfig    `mapstructure:"proxy"`
	MySQL  MySQLConfig  `mapstructure:"mysql"`
	Redis  RedisConfig  `mapstructure:"redis"`
	JWT    JWTConfig    `mapstructure:"jwt"`
	Crypto CryptoConfig `mapstructure:"crypto"`
	OAuth  OAuthConfig  `mapstructure:"oauth"`
	Mail   MailConfig    `mapstructure:"mail"`
	Wechat  WechatConfig  `mapstructure:"wechat"`
	Storage StorageConfig `mapstructure:"storage"`
	Pay     PayConfig     `mapstructure:"pay"`
	Observability ObservabilityConfig `mapstructure:"observability"`
	Swagger       SwaggerConfig       `mapstructure:"swagger"`
	Backup        BackupConfig        `mapstructure:"backup"`
	Rag           RagConfig           `mapstructure:"rag"`
}

// RagConfig RAG 知识库开关、配额、Embedding 与 LLM 配置（对齐 Nest ragConfig）。
type RagConfig struct {
	Enabled            bool              `mapstructure:"enabled"`
	DailyQuota         int               `mapstructure:"daily_quota"`
	TopK               int               `mapstructure:"top_k"`
	AllowLocalFallback bool              `mapstructure:"allow_local_fallback"`
	Embedding          RagEmbeddingConfig `mapstructure:"embedding"`
	LLM                RagLLMConfig       `mapstructure:"llm"`
	Chunk              RagChunkConfig     `mapstructure:"chunk"`
}

// RagEmbeddingConfig OpenAI 兼容 Embedding API。
type RagEmbeddingConfig struct {
	Mode      string `mapstructure:"mode"`
	RemoteURL string `mapstructure:"remote_url"`
	APIKey    string `mapstructure:"api_key"`
	Model     string `mapstructure:"model"`
}

// RagLLMConfig 对话模型（OpenAI 兼容 chat/completions）。
type RagLLMConfig struct {
	BaseURL string `mapstructure:"base_url"`
	APIKey  string `mapstructure:"api_key"`
	Model   string `mapstructure:"model"`
}

// RagChunkConfig Markdown 分块参数。
type RagChunkConfig struct {
	Size    int `mapstructure:"size"`
	Overlap int `mapstructure:"overlap"`
}

// RagDailyQuotaOrDefault 每日问答配额，默认 20。
func (r RagConfig) RagDailyQuotaOrDefault() int {
	if r.DailyQuota > 0 {
		return r.DailyQuota
	}
	return 20
}

// RagTopKOrDefault 检索 Top-K，默认 6。
func (r RagConfig) RagTopKOrDefault() int {
	if r.TopK > 0 {
		return r.TopK
	}
	return 6
}

// SwaggerConfig OpenAPI/Swagger UI 开关与路径。
type SwaggerConfig struct {
	Enabled     bool   `mapstructure:"enabled"`
	PathPrefix  string `mapstructure:"path_prefix"`
}

// AppConfig 应用级元信息。
type AppConfig struct {
	Name               string      `mapstructure:"name"`
	Env                string      `mapstructure:"env"`
	ServiceMode        ServiceMode `mapstructure:"service_mode"`
	APIPrefix          string      `mapstructure:"api_prefix"`
	BlogHome           string      `mapstructure:"blog_home"`
	NotifyEmail        string      `mapstructure:"notify_email"`
	TongjiRefreshToken string      `mapstructure:"tongji_refresh_token"`
	TongjiClientID     string      `mapstructure:"tongji_client_id"`
	TongjiClientSecret string      `mapstructure:"tongji_client_secret"`
}

// BackupConfig 数据库定时备份（mysqldump）配置。
type BackupConfig struct {
	Dir          string `mapstructure:"dir"`
	MysqldumpPath string `mapstructure:"mysqldump_path"`
}

// ServiceModeOrDefault 未配置时视为 monolith。
func (a AppConfig) ServiceModeOrDefault() ServiceMode {
	if a.ServiceMode == "" {
		return ModeMonolith
	}
	return a.ServiceMode
}

// HTTPConfig HTTP 服务监听与 CORS。
type HTTPConfig struct {
	Addr        string   `mapstructure:"addr"`
	CORSOrigins []string `mapstructure:"cors_origins"`
}

// KitexConfig 本服务 Kitex RPC 监听（学习路径微服务；monolith 不启）。
type KitexConfig struct {
	// Addr 本进程 Kitex 监听地址，如 ":50052"。
	Addr string `mapstructure:"addr"`
}

// RegistryConfig Nacos 服务注册/发现（学习路径；monolith 不需要）。
type RegistryConfig struct {
	// NacosAddr Nacos 地址，如 "127.0.0.1:8848"（Docker 内为 "nacos:8848"）。
	NacosAddr string `mapstructure:"nacos_addr"`
	// NamespaceID Nacos 命名空间；空或 public 表示默认 public。
	NamespaceID string `mapstructure:"namespace_id"`
	// Group 服务分组，默认 DEFAULT_GROUP。
	Group string `mapstructure:"group"`
	// Username / Password 可选；学习环境通常关闭鉴权。
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

// Enabled 是否配置了 Nacos 地址（学习路径微服务/gateway 需要）。
func (r RegistryConfig) Enabled() bool {
	return strings.TrimSpace(r.NacosAddr) != ""
}

// Kitex 服务注册名（与 Nacos / 客户端发现一致）。
const (
	KitexServiceUser = "blog.user"
	KitexServiceBlog = "blog.blog"
	KitexServiceRPG  = "blog.rpg"
)

// ProxyConfig gateway 反向代理上游 HTTP 地址。
type ProxyConfig struct {
	UserURL string `mapstructure:"user_url"`
	BlogURL string `mapstructure:"blog_url"`
	RPGURL  string `mapstructure:"rpg_url"`
}

// ObservabilityConfig 可观测性开关。
type ObservabilityConfig struct {
	EnableMetrics bool   `mapstructure:"enable_metrics"`
	EnablePprof   bool   `mapstructure:"enable_pprof"`
	PprofAddr     string `mapstructure:"pprof_addr"`
	OTLPEndpoint  string `mapstructure:"otlp_endpoint"`
	ServiceName   string `mapstructure:"service_name"`
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

// IsMicroservice 是否为拆分后的内部服务或 gateway。
func (c *Config) IsMicroservice() bool {
	switch c.App.ServiceModeOrDefault() {
	case ModeUser, ModeBlog, ModeRPG, ModeGateway:
		return true
	default:
		return false
	}
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
