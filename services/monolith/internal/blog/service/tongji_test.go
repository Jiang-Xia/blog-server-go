package service

import (
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

func TestIsBaiduTokenExpired(t *testing.T) {
	if !isBaiduTokenExpired(map[string]interface{}{"error_code": float64(110)}) {
		t.Fatal("110 should be expired")
	}
	if isBaiduTokenExpired(map[string]interface{}{"error_code": float64(0)}) {
		t.Fatal("0 should not be expired")
	}
}

func TestTongjiTokenRedisKeyConstant(t *testing.T) {
	if tongjiTokenRedisKey != "baidu:tongji:access_token" {
		t.Fatalf("redis key mismatch: %s", tongjiTokenRedisKey)
	}
}
