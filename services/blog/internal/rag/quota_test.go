package rag_test

import (
	"testing"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/rag"
)

func TestRagDailyQuotaDefault(t *testing.T) {
	cfg := config.RagConfig{}
	if cfg.RagDailyQuotaOrDefault() != 20 {
		t.Fatalf("default daily quota want 20")
	}
	cfg.DailyQuota = 5
	if cfg.RagDailyQuotaOrDefault() != 5 {
		t.Fatalf("custom daily quota want 5")
	}
}

func TestNewQuotaServiceLimit(t *testing.T) {
	svc := rag.NewQuotaService(&config.Config{Rag: config.RagConfig{DailyQuota: 10}}, nil)
	// 通过 GetUsage 在无 redis 时会失败；此处仅验证构造不 panic
	if svc == nil {
		t.Fatal("nil quota service")
	}
}
