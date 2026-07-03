// grpc.go user-service gRPC 敏感词过滤实现。
package contentfilter

import (
	"context"

	"github.com/Jiang-Xia/blog-server-go/pkg/usersvc"
)

// GRPCFilter 经 user gRPC 做敏感词检测与命中记录写入。
type GRPCFilter struct {
	client usersvc.ContentFilter
}

// NewGRPCFilter 构造 FilterService。
func NewGRPCFilter(client usersvc.ContentFilter) *GRPCFilter {
	return &GRPCFilter{client: client}
}

// EvaluateContent 调用 user-service 敏感词检测。
func (f *GRPCFilter) EvaluateContent(ctx context.Context, content string) (*EvaluateResult, error) {
	result, err := f.client.EvaluateContent(ctx, content)
	if err != nil {
		return nil, err
	}
	return &EvaluateResult{
		Content:    result.Content,
		HitWords:   result.HitWords,
		HpPenalty:  result.HpPenalty,
		NeedReview: result.NeedReview,
		Rejected:   result.Rejected,
	}, nil
}

// CreateHitRecord 写入敏感词命中记录。
func (f *GRPCFilter) CreateHitRecord(ctx context.Context, params CreateHitParams) error {
	return f.client.CreateHitRecord(ctx, usersvc.FilterHitParams{
		SourceType: params.SourceType,
		SourceID:   params.SourceID,
		Content:    params.Content,
		HitWords:   params.HitWords,
		UID:        params.UID,
		IP:         params.IP,
	})
}
