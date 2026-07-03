package middleware

import "testing"

func TestCompilePublicProfileCollectsPattern(t *testing.T) {
	m := newPathMatcher()
	cases := map[string]string{
		"/user/public/:uid/collects": "/user/public/1/collects",
		"/user/public/:uid/likes":    "/user/public/1/likes",
	}
	for pattern, url := range cases {
		if !m.match(url, pattern) {
			t.Fatalf("pattern %q should match %q", pattern, url)
		}
	}
}
