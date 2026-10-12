package service

import (
	"context"
	"testing"
)

func TestNormalizeTongjiQueryRequiresURL(t *testing.T) {
	_, err := normalizeTongjiQuery(map[string]string{})
	if err == nil {
		t.Fatal("expected error for missing url")
	}
}

func TestNormalizeTongjiQueryRejectsInvalidPrefix(t *testing.T) {
	_, err := normalizeTongjiQuery(map[string]string{"url": "http://evil.com/rest"})
	if err == nil {
		t.Fatal("expected error for invalid url prefix")
	}
}

func TestNormalizeTongjiQueryAcceptsValidPath(t *testing.T) {
	out, err := normalizeTongjiQuery(map[string]string{
		"url":    "/rest/2.0/tongji/report/getData",
		"siteId": "1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if out["url"] != "/rest/2.0/tongji/report/getData" || out["siteId"] != "1" {
		t.Fatalf("unexpected out: %+v", out)
	}
}

// 代理模式下 ForceRefresh 不做实际刷新，返回 managed_by 提示。
func TestForceRefreshReturnsNestManagedInfo(t *testing.T) {
	svc := &ResourcesService{}
	out, err := svc.ForceRefreshTongjiAccessToken(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if v, _ := out["managed_by"].(string); v != "blog-server" {
		t.Fatalf("unexpected managed_by: %v", out["managed_by"])
	}
	if refreshed, _ := out["refreshed"].(bool); refreshed {
		t.Fatal("proxy mode should not report refreshed=true")
	}
}
