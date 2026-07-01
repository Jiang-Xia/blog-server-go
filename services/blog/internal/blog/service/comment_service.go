// comment_service 评论业务逻辑（限流、敏感词、站内通知）。
package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/pkg/redisutil"
	"github.com/Jiang-Xia/blog-server-go/services/blog/ent"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/blog/notification"
	blogrepo "github.com/Jiang-Xia/blog-server-go/services/blog/internal/blog/repo"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/blog/util"
	"github.com/Jiang-Xia/blog-server-go/pkg/usersvc"
	contentfilter "github.com/Jiang-Xia/blog-server-go/services/blog/internal/contentfilter"
	"github.com/Jiang-Xia/blog-server-go/pkg/pagination"
	"github.com/google/uuid"
)

const (
	commentRateWindowSec = 60
	commentRateMax        = 6
)

// CommentService 评论业务逻辑。
type CommentService struct {
	comments *blogrepo.CommentRepo
	replies  *ReplyService
	articles *blogrepo.ArticleRepo
	users    usersvc.UserService
	filter   contentfilter.FilterService
	redis    *redisutil.Store
	notify   *notification.Service
}

// NewCommentService 构造 CommentService。
func NewCommentService(
	comments *blogrepo.CommentRepo,
	replies *ReplyService,
	articles *blogrepo.ArticleRepo,
	users usersvc.UserService,
	filter contentfilter.FilterService,
	redis *redisutil.Store,
	notify *notification.Service,
) *CommentService {
	return &CommentService{
		comments: comments,
		replies:  replies,
		articles: articles,
		users:    users,
		filter:   filter,
		redis:    redis,
		notify:   notify,
	}
}

// Create 创建评论（限流 + 敏感词 + 通知）。
func (s *CommentService) Create(ctx context.Context, uid int, articleID int, content, ip string) (map[string]interface{}, error) {
	if err := s.assertRateLimit(ctx, uid, ip); err != nil {
		return nil, err
	}
	uidPtr := &uid
	ipPtr := &ip
	eval, err := s.filter.EvaluateContent(ctx, content)
	if err != nil {
		return nil, err
	}
	if len(eval.HitWords) > 0 && eval.Rejected {
		recordSensitiveHit(ctx, s.filter, "comment", "0", content, eval.HitWords, uidPtr, ipPtr)
		return nil, errcode.WithMessage(errcode.InvalidParam, "内容包含违规词汇，无法发布")
	}
	// 审核状态：默认通过；命中需人工复核词时 pending（对齐 Nest 敏感词等级）。
	status := "approved"
	if eval.NeedReview {
		status = "pending"
	}
	body := content
	if len(eval.HitWords) > 0 {
		// 低等级命中：使用过滤后正文入库，仍允许发布。
		body = eval.Content
	}
	article, err := s.articles.GetByID(ctx, articleID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errcode.WithMessage(errcode.NotFound, "文章不存在")
		}
		return nil, err
	}
	id := uuid.NewString()
	row, err := s.comments.Create(ctx, &ent.Comment{
		ID:        id,
		Content:   util.EscapeHTML(body),
		UID:       uid,
		UserId:    &uid,
		ArticleId: &articleID,
		Status:    status,
	})
	if err != nil {
		return nil, err
	}
	recordSensitiveHit(ctx, s.filter, "comment", id, content, eval.HitWords, uidPtr, ipPtr)
	if article.UID != uid && s.notify != nil {
		_, _ = s.notify.Create(ctx, article.UID, "comment_on_article", map[string]interface{}{
			"commentId":    id,
			"articleId":    article.ID,
			"articleTitle": article.Title,
			"fromUid":      uid,
			"status":       status,
		})
	}
	return map[string]interface{}{"id": row.ID, "status": row.Status}, nil
}

// Delete 删除评论及下属回复。
func (s *CommentService) Delete(ctx context.Context, id string) error {
	if err := s.comments.Delete(ctx, id); err != nil {
		return err
	}
	return s.replies.DeleteByParentID(ctx, id)
}

// FindAll 文章下已审核评论（含回复）。
func (s *CommentService) FindAll(ctx context.Context, articleID, page, pageSize int, sort string) (map[string]interface{}, error) {
	if pageSize <= 0 {
		pageSize = 100
	}
	aid := articleID
	rows, total, err := s.comments.List(ctx, blogrepo.CommentFilter{
		ArticleID: &aid,
		Status:    "approved",
		Page:      page,
		PageSize:  pageSize,
		SortAsc:   strings.ToUpper(sort) == "ASC",
	})
	if err != nil {
		return nil, err
	}
	list, err := s.enrichComments(ctx, rows, true)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"list":       list,
		"pagination": pagination.CalcNestPagination(total, pageSize, page),
	}, nil
}

// FindAllAdmin 管理端评论列表。
func (s *CommentService) FindAllAdmin(ctx context.Context, q map[string]interface{}) (map[string]interface{}, error) {
	page := intField(q, "page", 1)
	pageSize := intField(q, "pageSize", 10)
	f := blogrepo.CommentFilter{
		Status:   strField(q, "status"),
		Content:  strField(q, "content"),
		Page:     page,
		PageSize: pageSize,
		SortAsc:  strings.ToUpper(strField(q, "sort")) == "ASC",
	}
	if aid := intField(q, "articleId", 0); aid > 0 {
		f.ArticleID = &aid
	}
	rows, total, err := s.comments.List(ctx, f)
	if err != nil {
		return nil, err
	}
	list, err := s.enrichComments(ctx, rows, true)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"list":       list,
		"pagination": pagination.CalcNestPagination(total, pageSize, page),
	}, nil
}

// FindMyComments 当前用户评论列表。
func (s *CommentService) FindMyComments(ctx context.Context, uid, page, pageSize int) (map[string]interface{}, error) {
	rows, total, err := s.comments.List(ctx, blogrepo.CommentFilter{
		UID:      &uid,
		Status:   "approved",
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		return nil, err
	}
	list := make([]map[string]interface{}, 0, len(rows))
	for _, c := range rows {
		item := map[string]interface{}{
			"id": c.ID, "content": c.Content, "createTime": c.CreateTime,
		}
		if c.ArticleId != nil {
			if art, err := s.articles.GetByID(ctx, *c.ArticleId); err == nil {
				item["articleId"] = art.ID
				item["articleTitle"] = art.Title
			}
		}
		list = append(list, item)
	}
	return map[string]interface{}{
		"list":       list,
		"pagination": pagination.CalcNestPagination(total, pageSize, page),
	}, nil
}

// FindOnMyArticles 作者收到的评论。
func (s *CommentService) FindOnMyArticles(ctx context.Context, authorUID, page, pageSize int) (map[string]interface{}, error) {
	arts, err := s.articles.ListPublishedByAuthor(ctx, authorUID)
	if err != nil {
		return nil, err
	}
	ids := make([]int, 0, len(arts))
	for _, a := range arts {
		ids = append(ids, a.ID)
	}
	if len(ids) == 0 {
		return map[string]interface{}{
			"list":       []interface{}{},
			"pagination": pagination.CalcNestPagination(0, pageSize, page),
		}, nil
	}
	rows, total, err := s.comments.List(ctx, blogrepo.CommentFilter{
		ArticleIDs: ids,
		Page:       page,
		PageSize:   pageSize,
	})
	if err != nil {
		return nil, err
	}
	list := make([]map[string]interface{}, 0, len(rows))
	uids := make([]uint64, 0, len(rows))
	for _, c := range rows {
		uids = append(uids, uint64(c.UID))
	}
	users, _ := s.users.GetUserBatch(ctx, uids)
	userMap := map[uint64]*usersvc.UserDTO{}
	for _, u := range users {
		if u != nil {
			userMap[u.ID] = u
		}
	}
	for _, c := range rows {
		item := map[string]interface{}{
			"id": c.ID, "content": c.Content, "status": c.Status, "createTime": c.CreateTime,
			"userInfo": util.UserInfoMap(userMap[uint64(c.UID)], "nickname", "id", "avatar"),
		}
		if c.ArticleId != nil {
			if art, err := s.articles.GetByID(ctx, *c.ArticleId); err == nil {
				item["articleId"] = art.ID
				item["articleTitle"] = art.Title
			}
		}
		list = append(list, item)
	}
	return map[string]interface{}{
		"list":       list,
		"pagination": pagination.CalcNestPagination(total, pageSize, page),
	}, nil
}

// CountByArticleIDs 批量统计评论数。
func (s *CommentService) CountByArticleIDs(ctx context.Context, articleIDs []int) (map[int]int, error) {
	return s.comments.CountByArticleIDs(ctx, articleIDs)
}

func (s *CommentService) enrichComments(ctx context.Context, rows []*ent.Comment, withReplies bool) ([]map[string]interface{}, error) {
	uids := make([]uint64, 0, len(rows))
	for _, c := range rows {
		uids = append(uids, uint64(c.UID))
	}
	users, _ := s.users.GetUserBatch(ctx, uids)
	userMap := map[uint64]*usersvc.UserDTO{}
	for _, u := range users {
		if u != nil {
			userMap[u.ID] = u
		}
	}
	list := make([]map[string]interface{}, 0, len(rows))
	for _, c := range rows {
		item := map[string]interface{}{
			"id": c.ID, "content": c.Content, "uid": c.UID, "status": c.Status,
			"createTime": c.CreateTime, "updateTime": c.UpdateTime,
			// userInfo 由 user gRPC 批量拉取后组装，禁止前端 dict 回显。
			"userInfo": util.UserInfoMap(userMap[uint64(c.UID)]),
		}
		if c.ArticleId != nil {
			item["articleId"] = *c.ArticleId
		}
		if withReplies && s.replies != nil {
			replyData, _ := s.replies.FindAll(ctx, c.ID, 1, 100, "DESC")
			item["replys"] = replyData["list"]
			if t, ok := replyData["total"].(int); ok {
				item["allReplyCount"] = t
			}
		}
		list = append(list, item)
	}
	return list, nil
}

func (s *CommentService) assertRateLimit(ctx context.Context, uid int, ip string) error {
	safeUID := sanitizeRateKey(fmt.Sprintf("%d", uid))
	safeIP := sanitizeRateKey(ip)
	key := fmt.Sprintf("comment:rate:%s:%s", safeUID, safeIP)
	count, err := s.redis.Incr(ctx, key)
	if err != nil {
		return err
	}
	if count == 1 {
		// 滑动窗口：commentRateWindowSec=60s，commentRateMax=6 次/窗口。
		_ = s.redis.Expire(ctx, key, commentRateWindowSec)
	}
	if count > commentRateMax {
		return errcode.WithMessage(errcode.TooManyRequests, "评论过于频繁，请稍后再试")
	}
	return nil
}

func sanitizeRateKey(s string) string {
	if s == "" {
		return "unknown"
	}
	var b strings.Builder
	for _, r := range s {
		if (r >= '0' && r <= '9') || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '.' || r == '_' || r == '-' {
			b.WriteRune(r)
		} else {
			b.WriteRune('_')
		}
	}
	return b.String()
}

func strField(m map[string]interface{}, key string) string {
	if m == nil {
		return ""
	}
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return strings.TrimSpace(s)
	}
	return fmt.Sprint(v)
}

func intField(m map[string]interface{}, key string, def int) int {
	if m == nil {
		return def
	}
	v, ok := m[key]
	if !ok || v == nil {
		return def
	}
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case string:
		var i int
		fmt.Sscanf(n, "%d", &i)
		if i > 0 {
			return i
		}
	}
	return def
}
