// Package metrics 暴露 Prometheus 指标端点。
package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Handler 返回 /metrics 使用的 http.Handler。
func Handler() http.Handler {
	return promhttp.Handler()
}
