// comment_handler 评论 HTTP 端点。
package handler

import (
	"context"
	"strconv"

	"github.com/Jiang-Xia/blog-server-go/pkg/ctxutil"
	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/pkg/response"
	"github.com/Jiang-Xia/blog-server-go/pkg/rpgsvc"
	blogrepo "github.com/Jiang-Xia/blog-server-go/services/blog/internal/blog/repo"
	blogsvc "github.com/Jiang-Xia/blog-server-go/services/blog/internal/blog/service"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/auth"
	"github.com/cloudwego/hertz/pkg/app"
)

// CommentHandler 评论 HTTP 端点。
type CommentHandler struct {
	svc  *blogsvc.CommentService
	jwt  *auth.JWTService
	ban  rpgsvc.BanChecker
}

// NewCommentHandler 构造 CommentHandler。
func NewCommentHandler(svc *blogsvc.CommentService, jwt *auth.JWTService, ban rpgsvc.BanChecker) *CommentHandler {
	return &CommentHandler{svc: svc, jwt: jwt, ban: ban}
}

func (h *CommentHandler) Create(ctx context.Context, c *app.RequestContext) {
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
	articleID, err := blogrepo.ParseArticleIDInt(body["articleId"])
	if err != nil || articleID <= 0 {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	ip := string(c.ClientIP())
	data, err := h.svc.Create(ctx, uid, articleID, strField(body, "content"), ip)
	handleAdminResult(ctx, c, data, err)
}

func (h *CommentHandler) Delete(ctx context.Context, c *app.RequestContext) {
	id := string(c.Query("id"))
	if id == "" {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	handleAdminResult(ctx, c, nil, h.svc.Delete(ctx, id))
}

func (h *CommentHandler) FindAll(ctx context.Context, c *app.RequestContext) {
	articleID, _ := strconv.Atoi(string(c.Query("articleId")))
	page, _ := strconv.Atoi(string(c.Query("page")))
	pageSize, _ := strconv.Atoi(string(c.Query("pageSize")))
	sort := string(c.Query("sort"))
	data, err := h.svc.FindAll(ctx, articleID, page, pageSize, sort)
	handleAdminResult(ctx, c, data, err)
}

func (h *CommentHandler) Admin(ctx context.Context, c *app.RequestContext) {
	q := map[string]interface{}{}
	_ = c.Bind(&q)
	if q["articleId"] == nil {
		q["articleId"] = string(c.Query("articleId"))
	}
	if q["status"] == nil {
		q["status"] = string(c.Query("status"))
	}
	if q["content"] == nil {
		q["content"] = string(c.Query("content"))
	}
	if q["page"] == nil {
		q["page"] = string(c.Query("page"))
	}
	if q["pageSize"] == nil {
		q["pageSize"] = string(c.Query("pageSize"))
	}
	data, err := h.svc.FindAllAdmin(ctx, q)
	handleAdminResult(ctx, c, data, err)
}

func (h *CommentHandler) MyList(ctx context.Context, c *app.RequestContext) {
	uid := articleUID(ctx, c, h.jwt)
	if uid == 0 {
		response.Error(ctx, c, errcode.WithMessage(errcode.Unauthorized, "请先登录！"))
		return
	}
	page, _ := strconv.Atoi(string(c.Query("page")))
	pageSize, _ := strconv.Atoi(string(c.Query("pageSize")))
	data, err := h.svc.FindMyComments(ctx, uid, page, pageSize)
	handleAdminResult(ctx, c, data, err)
}

func (h *CommentHandler) OnMyArticles(ctx context.Context, c *app.RequestContext) {
	uid := ctxutil.UserID(ctx)
	if uid == 0 {
		uid = articleUID(ctx, c, h.jwt)
	}
	if uid == 0 {
		response.Error(ctx, c, errcode.WithMessage(errcode.Unauthorized, "请先登录！"))
		return
	}
	page, _ := strconv.Atoi(string(c.Query("page")))
	pageSize, _ := strconv.Atoi(string(c.Query("pageSize")))
	data, err := h.svc.FindOnMyArticles(ctx, uid, page, pageSize)
	handleAdminResult(ctx, c, data, err)
}
