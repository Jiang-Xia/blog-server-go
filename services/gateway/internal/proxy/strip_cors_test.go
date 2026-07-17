package proxy

import (
	"net/http"
	"testing"
)

func TestStripUpstreamCORS(t *testing.T) {
	resp := &http.Response{Header: make(http.Header)}
	resp.Header.Set("Access-Control-Allow-Origin", "http://evil.example")
	resp.Header.Set("Access-Control-Allow-Methods", "GET")
	resp.Header.Set("Access-Control-Allow-Headers", "X-Foo")
	resp.Header.Set("Access-Control-Expose-Headers", "X-Foo")
	resp.Header.Set("Access-Control-Max-Age", "1")
	resp.Header.Set("Access-Control-Allow-Credentials", "true")
	resp.Header.Set("Content-Type", "application/json")

	if err := stripUpstreamCORS(resp); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{
		"Access-Control-Allow-Origin",
		"Access-Control-Allow-Credentials",
		"Access-Control-Allow-Methods",
		"Access-Control-Allow-Headers",
		"Access-Control-Expose-Headers",
		"Access-Control-Max-Age",
	} {
		if got := resp.Header.Get(k); got != "" {
			t.Fatalf("%s still present: %q", k, got)
		}
	}
	if got := resp.Header.Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type stripped: %q", got)
	}
}
