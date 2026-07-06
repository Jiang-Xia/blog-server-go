package jsonkey_test

import (
	"testing"

	"github.com/Jiang-Xia/blog-server-go/pkg/jsonkey"
)

func TestSnakeToCamel(t *testing.T) {
	cases := map[string]string{
		"article_id":     "articleId",
		"chunk_count":    "chunkCount",
		"create_at":      "createAt",
		"citations_json": "citationsJson",
		"id":             "id",
		"isFolder":       "isFolder",
	}
	for in, want := range cases {
		if got := jsonkey.SnakeToCamel(in); got != want {
			t.Fatalf("SnakeToCamel(%q) = %q, want %q", in, got, want)
		}
	}
}
