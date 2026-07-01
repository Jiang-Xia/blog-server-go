// dev_push_handler 开发环境 WebSocket 推送调试端点。
package handler

import (
	"context"
	"encoding/json"

	"github.com/Jiang-Xia/blog-server-go/pkg/ctxutil"
	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/pkg/response"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/blog/ws"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/event"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/auth"
	"github.com/cloudwego/hertz/pkg/app"
)

// DevPushHandler 开发/冒烟：经 Redis pub/sub 验证跨模块 WS 推送。
type DevPushHandler struct {
	pusher    *ws.RealtimePusher
	publisher *event.Publisher
	jwt       *auth.JWTService
}

// NewDevPushHandler 构造 DevPushHandler。
func NewDevPushHandler(pusher *ws.RealtimePusher, publisher *event.Publisher, jwt *auth.JWTService) *DevPushHandler {
	return &DevPushHandler{pusher: pusher, publisher: publisher, jwt: jwt}
}

// TestPush POST /dev/ws-push?type=ping body 可选 JSON data；推送给当前登录用户。
func (h *DevPushHandler) TestPush(ctx context.Context, c *app.RequestContext) {
	uid := ctxutil.UserID(ctx)
	if uid == 0 && h.jwt != nil {
		uid = parseJWTUID(c, h.jwt)
	}
	if uid == 0 {
		response.Error(ctx, c, errcode.Unauthorized)
		return
	}
	msgType := string(c.Query("type"))
	if msgType == "" {
		msgType = "test"
	}
	var data map[string]interface{}
	if len(c.Request.Body()) > 0 {
		_ = json.Unmarshal(c.Request.Body(), &data)
	}
	if data == nil {
		data = map[string]interface{}{"ok": true}
	}
	if err := h.pusher.PushToUser(ctx, uint64(uid), msgType, 0, data); err != nil {
		response.Error(ctx, c, errcode.InternalError)
		return
	}
	response.Success(ctx, c, map[string]interface{}{"pushed": true, "type": msgType})
}

func parseJWTUID(c *app.RequestContext, jwtSvc *auth.JWTService) int {
	authz := string(c.GetHeader("Authorization"))
	if authz == "" {
		return 0
	}
	token := authz
	if len(token) > 7 {
		token = token[7:]
	}
	claims, err := jwtSvc.Verify(token)
	if err != nil || claims == nil {
		return 0
	}
	return claims.ID
}

// TestPushRedis POST /dev/ws-push-redis 经 Redis pub/sub 推送（验证跨模块路径）。
func (h *DevPushHandler) TestPushRedis(ctx context.Context, c *app.RequestContext) {
	uid := ctxutil.UserID(ctx)
	if uid == 0 {
		uid = parseJWTUID(c, h.jwt)
	}
	if uid == 0 {
		response.Error(ctx, c, errcode.Unauthorized)
		return
	}
	msgType := string(c.Query("type"))
	if msgType == "" {
		msgType = "testRedis"
	}
	raw, _ := json.Marshal(map[string]interface{}{"via": "redis"})
	body, _ := json.Marshal(ws.Message{Type: msgType, Data: raw})
	if err := h.pusher.PublishRedis(ctx, uint64(uid), "", body); err != nil {
		response.Error(ctx, c, errcode.InternalError)
		return
	}
	response.Success(ctx, c, map[string]interface{}{"published": true})
}

// TestEvent POST /dev/event-publish 向 blog:events Stream 发布测试事件。
func (h *DevPushHandler) TestEvent(ctx context.Context, c *app.RequestContext) {
	if h.publisher == nil {
		response.Error(ctx, c, errcode.InternalError)
		return
	}
	eventType := string(c.Query("type"))
	if eventType == "" {
		eventType = event.EventArticlePublished
	}
	uid := ctxutil.UserID(ctx)
	if uid == 0 {
		uid = parseJWTUID(c, h.jwt)
	}
	payload := event.ArticlePublishedPayload{UID: uid, ArticleID: 0}
	if err := h.publisher.PublishTest(ctx, eventType, payload); err != nil {
		response.Error(ctx, c, errcode.InternalError)
		return
	}
	response.Success(ctx, c, map[string]interface{}{"published": true, "type": eventType})
}
