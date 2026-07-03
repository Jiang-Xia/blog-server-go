package proxy

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"testing"

	"github.com/Jiang-Xia/blog-server-go/pkg/response"
)

func TestWriteUpstreamError(t *testing.T) {
	rec := httptest.NewRecorder()
	writeUpstreamError(rec, "blog")

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want 502", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Fatalf("content-type = %q", ct)
	}

	var body response.Body
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Code != upstreamUnavailableCode || body.BizCode != upstreamUnavailableCode {
		t.Fatalf("code/bizCode = %d/%d", body.Code, body.BizCode)
	}
	if body.Message == "" {
		t.Fatal("message should not be empty")
	}
}

func TestPickReturnsServiceName(t *testing.T) {
	r := &Router{user: &httputil.ReverseProxy{}, blog: &httputil.ReverseProxy{}, rpg: &httputil.ReverseProxy{}}
	svc, _ := r.pick("/api/v1/article/list", "/api/v1")
	if svc != "blog" {
		t.Fatalf("service = %q, want blog", svc)
	}
	svc, _ = r.pick("/api/v1/user/login", "/api/v1")
	if svc != "user" {
		t.Fatalf("service = %q, want user", svc)
	}
}
