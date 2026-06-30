// Package logger 提供 zap 结构化日志，开发/生产模式由配置 env 决定。
package logger

import (
	"strings"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New 根据应用环境创建 zap.Logger；开发环境输出 console，生产环境 JSON。
func New(cfg *config.Config) (*zap.Logger, error) {
	var zapCfg zap.Config
	if cfg.IsDev() {
		zapCfg = zap.NewDevelopmentConfig()
		zapCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		zapCfg = zap.NewProductionConfig()
	}
	zapCfg.EncoderConfig.TimeKey = "time"
	zapCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	zapCfg.OutputPaths = []string{"stdout"}
	zapCfg.ErrorOutputPaths = []string{"stderr"}
	zapCfg.InitialFields = map[string]interface{}{
		"app": strings.TrimSpace(cfg.App.Name),
		"env": strings.TrimSpace(cfg.App.Env),
	}
	return zapCfg.Build()
}
