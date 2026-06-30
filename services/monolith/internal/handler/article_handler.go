package handler

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/Jiang-Xia/blog-server-go/pkg/ctxutil"
	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/pkg/response"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/blog/domain"
	blogsvc "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/blog/service"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/auth"
	"github.com/cloudwego/hertz/pkg/app"
)

// ArticleHandler 文章 HTTP 端点，路径对齐 Nest ArticleController。
type ArticleHandler struct {
	svc *blogsvc.ArticleService
	jwt *auth.JWTService
}

// NewArticleHandler 构造 ArticleHandler。
func NewArticleHandler(svc *blogsvc.ArticleService, jwt *auth.JWTService) *ArticleHandler {
	return &ArticleHandler{svc: svc, jwt: jwt}
}

func (h *ArticleHandler) List(ctx context.Context, c *app.RequestContext) {
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	q := domain.ArticleListQuery{
		Page:        intField(body, "page"),
		PageSize:    intField(body, "pageSize"),
		Category:    strField(body, "category"),
		Title:       strField(body, "title"),
		Description: strField(body, "description"),
		Content:     strField(body, "content"),
		Sort:        strField(body, "sort"),
		Client:      boolField(body, "client"),
		Admin:       boolField(body, "admin"),
		CallerUID:   articleUID(ctx, c, h.jwt),
	}
	if tags, ok := body["tags"].([]interface{}); ok {
		for _, t := range tags {
			if s, ok := t.(string); ok {
				q.Tags = append(q.Tags, s)
			}
		}
	}
	if deptID, ok := body["deptId"]; ok {
		if n, ok := toInt(deptID); ok {
			q.DeptID = &n
		}
	}
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 {
		q.PageSize = 10
	}
	data, err := h.svc.List(ctx, q)
	handleAdminResult(ctx, c, data, err)
}

func (h *ArticleHandler) Info(ctx context.Context, c *app.RequestContext) {
	id := string(c.Query("id"))
	if id == "" {
		response.Error(ctx, c, errcode.WithMessage(errcode.InvalidParam, "请输入有效 id"))
		return
	}
	data, err := h.svc.Info(ctx, id)
	handleAdminResult(ctx, c, data, err)
}

func (h *ArticleHandler) Create(ctx context.Context, c *app.RequestContext) {
	uid := articleUID(ctx, c, h.jwt)
	if uid == 0 {
		response.Error(ctx, c, errcode.WithMessage(errcode.Unauthorized, "身份验证失败"))
		return
	}
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	in := parseCreateInput(body)
	data, err := h.svc.Create(ctx, uid, in)
	handleAdminResult(ctx, c, data, err)
}

func (h *ArticleHandler) Edit(ctx context.Context, c *app.RequestContext) {
	uid := articleUID(ctx, c, h.jwt)
	if uid == 0 {
		response.Error(ctx, c, errcode.WithMessage(errcode.Unauthorized, "身份验证失败"))
		return
	}
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	in := parseEditInput(body)
	data, err := h.svc.Edit(ctx, uid, in)
	handleAdminResult(ctx, c, data, err)
}

func (h *ArticleHandler) Delete(ctx context.Context, c *app.RequestContext) {
	uid := articleUID(ctx, c, h.jwt)
	if uid == 0 {
		response.Error(ctx, c, errcode.WithMessage(errcode.Unauthorized, "身份验证失败"))
		return
	}
	id, err := strconv.Atoi(string(c.Query("id")))
	if err != nil {
		response.Error(ctx, c, errcode.WithMessage(errcode.InvalidParam, "请输入有效 id"))
		return
	}
	data, err := h.svc.Delete(ctx, uid, id)
	handleAdminResult(ctx, c, data, err)
}

func (h *ArticleHandler) Views(ctx context.Context, c *app.RequestContext) {
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	id, ok := toInt(body["id"])
	if !ok {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	ok2, err := h.svc.UpdateViews(ctx, id)
	handleAdminResult(ctx, c, ok2, err)
}

func (h *ArticleHandler) Likes(ctx context.Context, c *app.RequestContext) {
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	articleID, ok := toInt(body["articleId"])
	if !ok {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	status, _ := toInt(body["status"])
	err := h.svc.UpdateLikes(ctx, articleID, status)
	handleAdminResult(ctx, c, true, err)
}

func (h *ArticleHandler) Disabled(ctx context.Context, c *app.RequestContext) {
	h.patchField(ctx, c, "isDelete")
}

func (h *ArticleHandler) Topping(ctx context.Context, c *app.RequestContext) {
	h.patchField(ctx, c, "topping")
}

func (h *ArticleHandler) patchField(ctx context.Context, c *app.RequestContext, key string) {
	uid := articleUID(ctx, c, h.jwt)
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	id, ok := toInt(body["id"])
	if !ok {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	fields := map[string]interface{}{}
	if key == "isDelete" {
		fields["isDelete"] = boolField(body, "isDelete")
	} else if key == "topping" {
		if v, ok := toInt(body["topping"]); ok {
			fields["topping"] = v
		}
	}
	data, err := h.svc.UpdateField(ctx, uid, id, fields)
	handleAdminResult(ctx, c, data, err)
}

func (h *ArticleHandler) MyList(ctx context.Context, c *app.RequestContext) {
	uid := articleUID(ctx, c, h.jwt)
	if uid == 0 {
		response.Error(ctx, c, errcode.WithMessage(errcode.Unauthorized, "身份验证失败"))
		return
	}
	page, _ := strconv.Atoi(string(c.Query("page")))
	pageSize, _ := strconv.Atoi(string(c.Query("pageSize")))
	data, err := h.svc.MyList(ctx, uid, page, pageSize)
	handleAdminResult(ctx, c, data, err)
}

func (h *ArticleHandler) Archives(ctx context.Context, c *app.RequestContext) {
	data, err := h.svc.Archives(ctx)
	handleAdminResult(ctx, c, data, err)
}

func (h *ArticleHandler) Related(ctx context.Context, c *app.RequestContext) {
	id := string(c.Query("id"))
	limit, _ := strconv.Atoi(string(c.Query("limit")))
	data, err := h.svc.Related(ctx, id, limit)
	handleAdminResult(ctx, c, data, err)
}

func (h *ArticleHandler) AuthorStats(ctx context.Context, c *app.RequestContext) {
	uid := articleUID(ctx, c, h.jwt)
	if uid == 0 {
		response.Error(ctx, c, errcode.WithMessage(errcode.Unauthorized, "身份验证失败"))
		return
	}
	data, err := h.svc.AuthorStats(ctx, uid)
	handleAdminResult(ctx, c, data, err)
}

func (h *ArticleHandler) Statistics(ctx context.Context, c *app.RequestContext) {
	data, err := h.svc.Statistics(ctx)
	handleAdminResult(ctx, c, data, err)
}

func articleUID(ctx context.Context, c *app.RequestContext, jwt *auth.JWTService) int {
	if uid := ctxutil.UserID(ctx); uid != 0 {
		return uid
	}
	if jwt == nil {
		return 0
	}
	authz := strings.TrimSpace(string(c.GetHeader("Authorization")))
	if authz == "" {
		return 0
	}
	token := strings.TrimPrefix(authz, "Bearer ")
	claims, err := jwt.Verify(strings.TrimSpace(token))
	if err != nil || claims == nil {
		return 0
	}
	return claims.ID
}

func parseCreateInput(body map[string]interface{}) domain.CreateArticleInput {
	in := domain.CreateArticleInput{
		Title:       strField(body, "title"),
		Description: strField(body, "description"),
		Content:     strField(body, "content"),
		ContentHTML: strField(body, "contentHtml"),
		Cover:       strField(body, "cover"),
		Status:      strField(body, "status"),
	}
	if cat, ok := body["category"].(map[string]interface{}); ok {
		in.CategoryID = strField(cat, "id")
	} else {
		in.CategoryID = strField(body, "category")
	}
	in.TagIDs = parseTagIDs(body["tags"])
	if s := strField(body, "scheduledPublishAt"); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			in.ScheduledPublishAt = &t
		}
	}
	return in
}

func parseEditInput(body map[string]interface{}) domain.EditArticleInput {
	id, _ := toInt(body["id"])
	in := domain.EditArticleInput{ID: id}
	if v, ok := body["title"].(string); ok {
		in.Title = &v
	}
	if v, ok := body["description"].(string); ok {
		in.Description = &v
	}
	if v, ok := body["content"].(string); ok {
		in.Content = &v
	}
	if v, ok := body["contentHtml"].(string); ok {
		in.ContentHTML = &v
	}
	if v, ok := body["cover"].(string); ok {
		in.Cover = &v
	}
	if v, ok := body["status"].(string); ok {
		in.Status = &v
	}
	if v, ok := body["isDelete"].(bool); ok {
		in.IsDelete = &v
	}
	if cat, ok := body["category"].(map[string]interface{}); ok {
		cid := strField(cat, "id")
		in.CategoryID = &cid
	}
	in.TagIDs = parseTagIDs(body["tags"])
	return in
}

func parseTagIDs(raw interface{}) []string {
	switch v := raw.(type) {
	case string:
		if v == "" {
			return nil
		}
		return strings.Split(v, ",")
	case []interface{}:
		out := make([]string, 0, len(v))
		for _, item := range v {
			switch t := item.(type) {
			case string:
				out = append(out, t)
			case map[string]interface{}:
				out = append(out, strField(t, "id"))
			}
		}
		return out
	default:
		return nil
	}
}

func boolField(m map[string]interface{}, key string) bool {
	switch v := m[key].(type) {
	case bool:
		return v
	case float64:
		return v != 0
	case string:
		return v == "true" || v == "1"
	default:
		return false
	}
}
