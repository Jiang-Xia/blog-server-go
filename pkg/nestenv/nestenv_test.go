package nestenv

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParseFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "env.production")
	content := `# comment
db_host = 127.0.0.1
db_password = 'jxblog2048!@#'
serve_corsOrigins = https://a.example,https://b.example

auth_jwtSecret = secret-value
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := ParseFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"db_host":           "127.0.0.1",
		"db_password":       "jxblog2048!@#",
		"serve_corsOrigins": "https://a.example,https://b.example",
		"auth_jwtSecret":    "secret-value",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parse mismatch:\nwant %#v\ngot  %#v", want, got)
	}
}

func TestSplitCSV(t *testing.T) {
	got := SplitCSV(" https://a.example , https://b.example ")
	if len(got) != 2 || got[0] != "https://a.example" || got[1] != "https://b.example" {
		t.Fatalf("unexpected: %#v", got)
	}
}

func TestBlogYAMLMapsTongjiFromNestEnv(t *testing.T) {
	m := map[string]string{
		"app_tongjiRefreshToken": "refresh-token",
		"app_tongjiClientId":     "client-id",
		"app_tongjiClientSecret": "client-secret",
		"app_notifyEmail":        "admin@example.com",
	}
	yaml := BlogYAML(m)
	app, ok := yaml["app"].(map[string]any)
	if !ok {
		t.Fatal("missing app block")
	}
	if app["tongji_refresh_token"] != "refresh-token" {
		t.Fatalf("tongji_refresh_token: %#v", app["tongji_refresh_token"])
	}
	if app["tongji_client_id"] != "client-id" {
		t.Fatalf("tongji_client_id: %#v", app["tongji_client_id"])
	}
	if app["tongji_client_secret"] != "client-secret" {
		t.Fatalf("tongji_client_secret: %#v", app["tongji_client_secret"])
	}
	if app["notify_email"] != "admin@example.com" {
		t.Fatalf("notify_email: %#v", app["notify_email"])
	}
}

func TestRagBlockMapsNestEnv(t *testing.T) {
	m := map[string]string{
		"rag_enabled":                  "true",
		"rag_api_key":                  "llm-key",
		"rag_api_base_url":             "https://api.deepseek.com/v1",
		"rag_embedding_api_key":        "embed-key",
		"rag_embedding_api_base_url":   "https://api.siliconflow.cn/v1",
		"rag_embedding_model":          "BAAI/bge-large-zh-v1.5",
		"rag_chat_model":               "deepseek-chat",
		"rag_daily_query_limit":        "30",
		"rag_top_k":                    "8",
		"rag_allow_local_fallback":     "false",
	}
	block := RagBlock(m)
	if block["enabled"] != true {
		t.Fatalf("enabled: %#v", block["enabled"])
	}
	if block["daily_quota"] != 30 {
		t.Fatalf("daily_quota: %#v", block["daily_quota"])
	}
	if block["top_k"] != 8 {
		t.Fatalf("top_k: %#v", block["top_k"])
	}
	if block["allow_local_fallback"] != false {
		t.Fatalf("allow_local_fallback: %#v", block["allow_local_fallback"])
	}
	embedding, ok := block["embedding"].(map[string]any)
	if !ok {
		t.Fatal("missing embedding block")
	}
	if embedding["mode"] != "remote" {
		t.Fatalf("embedding.mode: %#v", embedding["mode"])
	}
	if embedding["api_key"] != "embed-key" {
		t.Fatalf("embedding.api_key: %#v", embedding["api_key"])
	}
	llm, ok := block["llm"].(map[string]any)
	if !ok {
		t.Fatal("missing llm block")
	}
	if llm["api_key"] != "llm-key" {
		t.Fatalf("llm.api_key: %#v", llm["api_key"])
	}
	if llm["model"] != "deepseek-chat" {
		t.Fatalf("llm.model: %#v", llm["model"])
	}
}

func TestBlogYAMLIncludesRagBlock(t *testing.T) {
	m := map[string]string{
		"rag_enabled":           "true",
		"rag_api_key":           "llm-key",
		"rag_embedding_api_key": "embed-key",
	}
	yaml := BlogYAML(m)
	if _, ok := yaml["rag"]; !ok {
		t.Fatal("blog yaml missing rag block")
	}
}
