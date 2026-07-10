package middleware

import "testing"

func TestCompilePublicProfilePattern(t *testing.T) {
	m := newPathMatcher()
	cases := map[string]string{
		"/user/public/:uid":          "/user/public/5",
		"/user/public/:uid/articles": "/user/public/5/articles",
		"/user/public/:uid/collects": "/user/public/5/collects",
		"/user/public/:uid/likes":    "/user/public/5/likes",
	}
	for pattern, url := range cases {
		if !m.match(url, pattern) {
			t.Fatalf("pattern %q should match %q", pattern, url)
		}
	}
}
