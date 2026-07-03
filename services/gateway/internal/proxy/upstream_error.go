package proxy

import (
	"fmt"
	"net/http"

	"github.com/Jiang-Xia/blog-server-go/pkg/response"
)

const upstreamUnavailableCode = 502

func upstreamMessage(service string) string {
	if service == "" {
		return "上游服务暂不可用，请确认相关微服务已启动后重试"
	}
	return fmt.Sprintf("%s 服务暂不可用，请确认服务已启动后重试", service)
}

func writeUpstreamError(w http.ResponseWriter, service string) {
	response.WriteHTTPError(w, http.StatusBadGateway, upstreamUnavailableCode, upstreamMessage(service))
}
