package handler

import (
	"context"
	"strconv"

	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/pkg/response"
	blogrepo "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/blog/repo"
	blogsvc "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/blog/service"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/auth"
	"github.com/cloudwego/hertz/pkg/app"
)

// LikeHandler 点赞 HTTP 端点。
type LikeHandler struct {
	svc *blogsvc.LikeService
	jwt *auth.JWTService
}

// NewLikeHandler 构造 LikeHandler。
func NewLikeHandler(svc *blogsvc.LikeService, jwt *auth.JWTService) *LikeHandler {
	return &LikeHandler{svc: svc, jwt: jwt}
}

func (h *LikeHandler) Update(ctx context.Context, c *app.RequestContext) {
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
	uid := articleUID(ctx, c, h.jwt)
	ip := string(c.ClientIP())
	err = h.svc.UpdateLike(ctx, articleID, uid, ip, body["status"])
	handleAdminResult(ctx, c, nil, err)
}

func (h *LikeHandler) Check(ctx context.Context, c *app.RequestContext) {
	uid := articleUID(ctx, c, h.jwt)
	articleID, _ := strconv.Atoi(string(c.Query("articleId")))
	data, err := h.svc.CheckLiked(ctx, articleID, uid)
	handleAdminResult(ctx, c, data, err)
}

func (h *LikeHandler) MyIDs(ctx context.Context, c *app.RequestContext) {
	uid := articleUID(ctx, c, h.jwt)
	data, err := h.svc.GetLikedArticleIDs(ctx, uid)
	handleAdminResult(ctx, c, data, err)
}
