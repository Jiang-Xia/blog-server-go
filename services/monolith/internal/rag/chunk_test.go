package rag_test

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rag"
)

func TestSplitMarkdownWithDescription(t *testing.T) {
	svc := rag.NewChunkService(&config.Config{})
	chunks := svc.SplitMarkdown("# Title\n\nHello world.", "Title", "摘要")
	if len(chunks) < 2 {
		t.Fatalf("expected at least 2 chunks, got %d", len(chunks))
	}
	if chunks[0].ChunkIndex != 0 {
		t.Fatalf("description chunk index want 0, got %d", chunks[0].ChunkIndex)
	}
}

func TestSplitMarkdownCodeBlock(t *testing.T) {
	svc := rag.NewChunkService(&config.Config{})
	md := "# Demo\n\n```go\nfmt.Println(\"hi\")\n```\n"
	chunks := svc.SplitMarkdown(md, "Demo", "")
	if len(chunks) == 0 {
		t.Fatal("expected chunks")
	}
	foundCode := false
	for _, c := range chunks {
		if c.ContentType == "code" {
			foundCode = true
		}
	}
	if !foundCode {
		t.Fatal("expected code chunk")
	}
}

func TestSplitMarkdownLongChineseValidUTF8(t *testing.T) {
	cfg := &config.Config{}
	cfg.Rag.Chunk.Size = 600
	cfg.Rag.Chunk.Overlap = 120
	svc := rag.NewChunkService(cfg)
	md := "# 标题\n\n" + strings.Repeat("中文段落内容用于测试分块边界。", 80)
	chunks := svc.SplitMarkdown(md, "标题", "摘要说明")
	for i, c := range chunks {
		if !utf8.ValidString(c.Content) {
			t.Fatalf("chunk %d has invalid UTF-8", i)
		}
	}
}
