package rag_test

import (
	"testing"

	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rag"
)

func TestLoadStaticPageMarkdown(t *testing.T) {
	if len(rag.RAGStaticPages) == 0 {
		t.Fatal("no static pages")
	}
	md, err := rag.LoadStaticPageMarkdown(rag.RAGStaticPages[0])
	if err != nil {
		t.Fatal(err)
	}
	if len(md) < 50 {
		t.Fatalf("markdown too short: %d", len(md))
	}
}

func TestRagPageSourceKey(t *testing.T) {
	if rag.RagPageSourceKey("features/rpg-guide") != "page:features/rpg-guide" {
		t.Fatal("unexpected source key")
	}
}
