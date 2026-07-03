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
