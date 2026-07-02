//go:build smoke

package smoke_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Jiang-Xia/blog-server-go/test/testutil"
	"github.com/gorilla/websocket"
)

// 冒烟：health、公开 BFF、登录、WS ping/pong、dev 推送端点。
func TestSmokeHealthAndPublic(t *testing.T) {
	c, err := testutil.NewClient("")
	if err != nil {
		t.Fatal(err)
	}
	testutil.RequireServer(t, c)
	ctx := context.Background()

	resp, status, err := c.GET(ctx, "/health", "")
	testutil.AssertOK(t, "GET /health", resp, status, err)

	var data string
	if err := json.Unmarshal(resp.Data, &data); err != nil || data != "ok" {
		t.Fatalf("health data want ok, got %q err=%v", string(resp.Data), err)
	}

	resp, status, err = c.GET(ctx, "/pub/stats", "")
	testutil.AssertOK(t, "GET /pub/stats", resp, status, err)

	resp, status, err = c.GET(ctx, "/user/authCode", "")
	testutil.AssertOK(t, "GET /user/authCode", resp, status, err)
}

func TestSmokeLoginAndUserInfo(t *testing.T) {
	c, err := testutil.NewClient("")
	if err != nil {
		t.Fatal(err)
	}
	testutil.RequireServer(t, c)
	ctx := context.Background()

	token, err := c.Login(ctx)
	token = testutil.SkipUnlessLogin(t, c, token, err)
	resp, status, err := c.GET(ctx, "/user/info", token)
	testutil.AssertOK(t, "GET /user/info", resp, status, err)
}

func TestSmokeWebSocket(t *testing.T) {
	c, err := testutil.NewClient("")
	if err != nil {
		t.Fatal(err)
	}
	testutil.RequireServer(t, c)
	ctx := context.Background()

	token, err := c.Login(ctx)
	token = testutil.SkipUnlessLogin(t, c, token, err)

	resp, status, err := c.GET(ctx, "/notification/since?seq=0", token)
	testutil.AssertOK(t, "GET /notification/since", resp, status, err)

	wsBase := strings.Replace(strings.Replace(c.Origin, "https://", "wss://", 1), "http://", "ws://", 1)
	wsURL := wsBase + "/realtime?token=" + token
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("ws dial: %v", err)
	}
	defer conn.Close()

	deadline := time.Now().Add(15 * time.Second)
	if err := conn.WriteJSON(map[string]string{"type": "ping"}); err != nil {
		t.Fatalf("ws ping: %v", err)
	}
	for time.Now().Before(deadline) {
		var msg map[string]any
		_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
		if err := conn.ReadJSON(&msg); err != nil {
			continue
		}
		if msg["type"] == "pong" {
			return
		}
	}
	t.Fatal("ws pong timeout")
}

func TestSmokeDevPushEndpoints(t *testing.T) {
	c, err := testutil.NewClient("")
	if err != nil {
		t.Fatal(err)
	}
	testutil.RequireServer(t, c)
	ctx := context.Background()

	token, err := c.Login(ctx)
	token = testutil.SkipUnlessLogin(t, c, token, err)

	for _, path := range []string{
		"/dev/ws-push?type=smokeTest",
		"/dev/ws-push-redis",
		"/dev/event-publish",
	} {
		resp, status, err := c.POST(ctx, path, nil, token)
		if err != nil || status != http.StatusOK || !testutil.IsOK(resp) {
			t.Fatalf("POST %s failed: status=%d code=%d err=%v", path, status, resp.Code, err)
		}
	}
}
