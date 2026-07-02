package timeutil_test

import (
	"testing"
	"time"

	"github.com/Jiang-Xia/blog-server-go/pkg/timeutil"
)

func TestFormat(t *testing.T) {
	ts := time.Date(2026, 7, 2, 15, 4, 5, 0, time.Local)
	got := timeutil.Format(ts)
	if got != "2026-07-02 15:04:05" {
		t.Fatalf("unexpected format: %s", got)
	}
}

func TestNow(t *testing.T) {
	before := time.Now()
	got := timeutil.Now()
	after := time.Now()
	if got.Before(before) || got.After(after) {
		t.Fatalf("Now() out of range: %v", got)
	}
}
