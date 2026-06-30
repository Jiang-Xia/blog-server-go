//go:generate go run github.com/google/wire/cmd/wire
//go:build wireinject
// +build wireinject

// Package app wire 依赖注入声明（make wire 生成 wire_gen.go）。
package app

import (
	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/pkg/logger"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/data"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/handler"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/server"
	"github.com/google/wire"
)

// InitializeApp 装配单体应用全部依赖。
func InitializeApp(cfgPath string) (*App, error) {
	wire.Build(
		config.MustLoad,
		logger.New,
		data.NewEntClient,
		data.NewRedisClient,
		handler.NewHealthHandler,
		server.NewHTTPServer,
		NewApp,
	)
	return nil, nil
}
