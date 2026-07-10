package tools_test

import (
	"context"
	"testing"

	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/rag/tools"
)

func TestOrchestratorRankingRule(t *testing.T) {
	svc := tools.NewService(nil, nil, nil, nil)
	orch := tools.NewOrchestrator(svc, nil)
	// nil crossdb：Execute �?panic/fail �?只测不命中规则时返回�?	recs, err := orch.ResolveTools(context.Background(), "博客架构是什�?, tools.Context{})
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != 0 {
		t.Fatalf("tutorial question should not trigger tools, got %v", recs)
	}
}

func TestGetSiteNav(t *testing.T) {
	svc := tools.NewService(nil, nil, nil, nil)
	nav, err := svc.Execute(context.Background(), "get_site_nav", nil, tools.Context{})
	if err != nil {
		t.Fatal(err)
	}
	m, ok := nav.(map[string]interface{})
	if !ok {
		t.Fatal("expected map")
	}
	if m["navLinks"] == nil || m["featurePages"] == nil {
		t.Fatalf("missing nav keys: %+v", m)
	}
}
