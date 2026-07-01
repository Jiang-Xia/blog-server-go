// sensitive_content 敏感词命中记录辅助，供评论/留言等 service 复用。
package service

import (
	"context"

	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/blog/util"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/sensitive"
)

// sensitiveApplyResult 敏感词过滤结果。
type sensitiveApplyResult struct {
	Content string
	Status  string // approved | pending
}

// applySensitiveFilter 分级检测内容；拒绝时返回 400，需审核时 status=pending。
func applySensitiveFilter(
	ctx context.Context,
	filter sensitive.FilterService,
	sourceType, sourceID, raw string,
	uid *int, ip *string,
) (*sensitiveApplyResult, error) {
	eval, err := filter.EvaluateContent(ctx, raw)
	if err != nil {
		return nil, err
	}
	if len(eval.HitWords) == 0 {
		return &sensitiveApplyResult{Content: util.EscapeHTML(raw), Status: "approved"}, nil
	}
	if eval.Rejected {
		_ = filter.CreateHitRecord(ctx, sensitive.CreateHitParams{
			SourceType: sourceType,
			SourceID:   "0",
			Content:    raw,
			HitWords:   eval.HitWords,
			UID:        uid,
			IP:         ip,
		})
		return nil, errcode.WithMessage(errcode.InvalidParam, "内容包含违规词汇，无法发布")
	}
	status := "approved"
	if eval.NeedReview {
		status = "pending"
	}
	content := util.EscapeHTML(eval.Content)
	return &sensitiveApplyResult{Content: content, Status: status}, nil
}

// recordSensitiveHit 内容已入库后写入命中记录。
func recordSensitiveHit(
	ctx context.Context,
	filter sensitive.FilterService,
	sourceType, sourceID, raw string,
	hitWords []string,
	uid *int, ip *string,
) {
	if len(hitWords) == 0 {
		return
	}
	_ = filter.CreateHitRecord(ctx, sensitive.CreateHitParams{
		SourceType: sourceType,
		SourceID:   sourceID,
		Content:    raw,
		HitWords:   hitWords,
		UID:        uid,
		IP:         ip,
	})
}
