package rag_test

import (
	"strings"
	"testing"

	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rag"
)

func TestResolveQuestionFromMessages(t *testing.T) {
	body, err := rag.ParseQueryBody([]byte(`{"messages":[{"role":"user","content":"博客架构是什么？"}]}`))
	if err != nil {
		t.Fatal(err)
	}
	q, err := rag.ResolveQuestion(body)
	if err != nil {
		t.Fatal(err)
	}
	if q != "博客架构是什么？" {
		t.Fatalf("unexpected question: %q", q)
	}
}

func TestFormatUiMessageSSE(t *testing.T) {
	line := rag.FormatUiMessageSSE(map[string]string{"type": "start", "messageId": "m1"})
	if !strings.HasPrefix(line, "data: ") {
		t.Fatalf("bad prefix: %q", line)
	}
	if !strings.HasSuffix(line, "\n\n") {
		t.Fatalf("bad suffix: %q", line)
	}
	done := rag.FormatUiMessageSSEDone()
	if done != "data: [DONE]\n\n" {
		t.Fatalf("bad done frame: %q", done)
	}
}

func TestExtractChatHistory(t *testing.T) {
	body, _ := rag.ParseQueryBody([]byte(`{"messages":[
		{"role":"user","content":"第一篇"},
		{"role":"assistant","content":"好的"},
		{"role":"user","content":"再详细说说"}
	]}`))
	h := rag.ExtractChatHistory(body)
	if len(h) != 2 {
		t.Fatalf("want 2 history turns, got %d", len(h))
	}
}
