// reply_handler 回复 HTTP 端点。
package handler

import (
	"context"
	"strconv"

	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/pkg/response"
	"github.com/Jiang-Xia/blog-server-go/pkg/rpgsvc"
	blogrepo "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/blog/repo"
	blogsvc "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/blog/service"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/auth"
	"github.com/cloudwego/hertz/pkg/app"
)

// ReplyHandler 回复 HTTP 端点。
type ReplyHandler struct {
	svc *blogsvc.ReplyService
	jwt *auth.JWTService
	ban rpgsvc.BanChecker
}

// NewReplyHandler 构造 ReplyHandler。
func NewReplyHandler(svc *blogsvc.ReplyService, jwt *auth.JWTService, ban rpgsvc.BanChecker) *ReplyHandler {
	return &ReplyHandler{svc: svc, jwt: jwt, ban: ban}
}

func (h *ReplyHandler) Create(ctx context.Context, c *app.RequestContext) {
	uid := articleUID(ctx, c, h.jwt)
	if uid == 0 {
		response.Error(ctx, c, errcode.WithMessage(errcode.Unauthorized, "请先登录！"))
		return
	}
	if h.ban != nil {
		if err := h.ban.AssertNotBanned(ctx, uid); err != nil {
			handleAdminResult(ctx, c, nil, err)
			return
		}
	}
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	data, err := h.svc.Create(ctx, uid, strField(body, "parentId"), strField(body, "replyUid"), strField(body, "content"))
	handleAdminResult(ctx, c, data, err)
}

func (h *ReplyHandler) Delete(ctx context.Context, c *app.RequestContext) {
	id := string(c.Query("id"))
	if id == "" {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	handleAdminResult(ctx, c, nil, h.svc.Delete(ctx, id))
}

func (h *ReplyHandler) FindAll(ctx context.Context, c *app.RequestContext) {
	parentID := string(c.Query("articleId"))
	if parentID == "" {
		parentID = string(c.Query("parentId"))
	}
	page, _ := strconv.Atoi(string(c.Query("page")))
	pageSize, _ := strconv.Atoi(string(c.Query("pageSize")))
	sort := string(c.Query("sort"))
	data, err := h.svc.FindAll(ctx, parentID, page, pageSize, sort)
	handleAdminResult(ctx, c, data, err)
}

func (h *ReplyHandler) MyList(ctx context.Context, c *app.RequestContext) {
	uid := articleUID(ctx, c, h.jwt)
	if uid == 0 {
		response.Error(ctx, c, errcode.WithMessage(errcode.Unauthorized, "请先登录！"))
		return
	}
	page, _ := strconv.Atoi(string(c.Query("page")))
	pageSize, _ := strconv.Atoi(string(c.Query("pageSize")))
	data, err := h.svc.FindMyReplies(ctx, uid, page, pageSize)
	handleAdminResult(ctx, c, data, err)
}

// ParseArticleID 解析 articleId 供 handler 使用。
func ParseArticleID(v interface{}) (int, error) {
	return blogrepo.ParseArticleIDInt(v)
}
