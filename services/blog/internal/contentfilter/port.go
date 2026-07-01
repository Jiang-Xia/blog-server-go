// Package contentfilter 敏感词过滤端口（实现由 user-service 提供；blog 侧注入）。
package contentfilter

import "context"

// HitDetail 单条敏感词命中详情。
type HitDetail struct {
	Word       string `json:"word"`
	Level      int    `json:"level"`
	HpPenalty  int    `json:"hpPenalty"`
	NeedReview int    `json:"needReview"`
	Action     int    `json:"action"`
}

// EvaluateResult 分级检测结果。
type EvaluateResult struct {
	Content    string      `json:"content"`
	Hits       []HitDetail `json:"hits"`
	HitWords   []string    `json:"hitWords"`
	HpPenalty  int         `json:"hpPenalty"`
	NeedReview bool        `json:"needReview"`
	Rejected   bool        `json:"rejected"`
}

// CreateHitParams 写入敏感词命中记录参数。
type CreateHitParams struct {
	SourceType string
	SourceID   string
	Content    string
	HitWords   []string
	UID        *int
	IP         *string
}

// FilterService 供 comment/msgboard 等模块注入的过滤接口。
type FilterService interface {
	EvaluateContent(ctx context.Context, content string) (*EvaluateResult, error)
	CreateHitRecord(ctx context.Context, params CreateHitParams) error
}
