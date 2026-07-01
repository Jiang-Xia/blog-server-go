package contentfilter

import (
	"context"

	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/blog/util"
)

// NoopFilter 无敏感词库时的透传实现（Plan 11 待 user gRPC 扩展后替换）。
type NoopFilter struct{}

// NewNoopFilter 构造透传 FilterService。
func NewNoopFilter() *NoopFilter {
	return &NoopFilter{}
}

// EvaluateContent 直接 HTML 转义并放行。
func (n *NoopFilter) EvaluateContent(_ context.Context, content string) (*EvaluateResult, error) {
	return &EvaluateResult{Content: util.EscapeHTML(content)}, nil
}

// CreateHitRecord 空实现。
func (n *NoopFilter) CreateHitRecord(_ context.Context, _ CreateHitParams) error {
	return nil
}
