// Package apidoc 提供 Swagger UI 挂载与 swag 通用响应类型引用。
package apidoc

import (
	"strings"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/adaptor"
	httpSwagger "github.com/swaggo/http-swagger"
)

// Mount 在 Hertz 上注册 Swagger UI 与 doc.json；路径对齐 Nest `/api/v1/doc`。
func Mount(h *server.Hertz, cfg *config.Config) {
	if cfg == nil || !cfg.Swagger.Enabled {
		return
	}
	prefix := strings.TrimSuffix(cfg.Swagger.PathPrefix, "/")
	if prefix == "" {
		prefix = cfg.App.APIPrefix + "/doc"
	}
	docJSON := prefix + "/doc.json"
	h.GET(prefix+"/*any", adaptor.HertzHandler(httpSwagger.Handler(
		httpSwagger.URL(docJSON),
	)))
}
