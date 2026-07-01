package service

import (
	"context"
	"strconv"
	"strings"

	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent"
	blogrepo "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/blog/repo"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/blog/util"
	"github.com/Jiang-Xia/blog-server-go/pkg/usersvc"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/sensitive"
	userrepo "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/repo"
	"github.com/google/uuid"
)

// ReplyService 回复业务逻辑。
type ReplyService struct {
	replies  *blogrepo.ReplyRepo
	comments *blogrepo.CommentRepo
	articles *blogrepo.ArticleRepo
	users    usersvc.UserService
	filter   sensitive.FilterService
}

// NewReplyService 构造 ReplyService。
func NewReplyService(
	replies *blogrepo.ReplyRepo,
	comments *blogrepo.CommentRepo,
	articles *blogrepo.ArticleRepo,
	users usersvc.UserService,
	filter sensitive.FilterService,
) *ReplyService {
	return &ReplyService{
		replies:  replies,
		comments: comments,
		articles: articles,
		users:    users,
		filter:   filter,
	}
}

// Create 创建回复。
func (s *ReplyService) Create(ctx context.Context, uid int, parentID, replyUID, content string) (*ent.Reply, error) {
	uidPtr := &uid
	eval, err := s.filter.EvaluateContent(ctx, content)
	if err != nil {
		return nil, err
	}
	if len(eval.HitWords) > 0 && eval.Rejected {
		recordSensitiveHit(ctx, s.filter, "reply", "0", content, eval.HitWords, uidPtr, nil)
		return nil, errcode.WithMessage(errcode.InvalidParam, "内容包含违规词汇，无法发布")
	}
	status := "approved"
	if eval.NeedReview {
		status = "pending"
	}
	body := content
	if len(eval.HitWords) > 0 {
		body = eval.Content
	}
	id := uuid.NewString()
	row, err := s.replies.Create(ctx, &ent.Reply{
		ID:       id,
		ParentId: parentID,
		ReplyUid: replyUID,
		Content:  util.EscapeHTML(body),
		UID:      uid,
		Status:   status,
	})
	if err != nil {
		return nil, err
	}
	recordSensitiveHit(ctx, s.filter, "reply", id, content, eval.HitWords, uidPtr, nil)
	return row, nil
}

// Delete 删除回复。
func (s *ReplyService) Delete(ctx context.Context, id string) error {
	return s.replies.Delete(ctx, id)
}

// DeleteByParentID 删除评论下全部回复。
func (s *ReplyService) DeleteByParentID(ctx context.Context, parentID string) error {
	return s.replies.DeleteByParentID(ctx, parentID)
}

// FindAll 评论下已审核回复列表。
func (s *ReplyService) FindAll(ctx context.Context, parentID string, page, pageSize int, sort string) (map[string]interface{}, error) {
	rows, total, err := s.replies.List(ctx, blogrepo.ReplyFilter{
		ParentID: parentID,
		Status:   "approved",
		Page:     page,
		PageSize: pageSize,
		SortAsc:  strings.ToUpper(sort) == "ASC",
	})
	if err != nil {
		return nil, err
	}
	list, err := s.enrichReplies(ctx, rows)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"list":       list,
		"total":      total,
		"pagination": userrepo.CalcNestPagination(total, pageSize, page),
	}, nil
}

// FindMyReplies 当前用户回复列表。
func (s *ReplyService) FindMyReplies(ctx context.Context, uid, page, pageSize int) (map[string]interface{}, error) {
	rows, total, err := s.replies.List(ctx, blogrepo.ReplyFilter{
		UID:      &uid,
		Status:   "approved",
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		return nil, err
	}
	list := make([]map[string]interface{}, 0, len(rows))
	for _, r := range rows {
		item := map[string]interface{}{
			"id": r.ID, "content": r.Content, "createTime": r.CreateTime,
			"parentId": r.ParentId,
		}
		if parent, err := s.comments.GetByID(ctx, r.ParentId); err == nil {
			item["parentCommentContent"] = parent.Content
			if parent.ArticleId != nil {
				item["articleId"] = *parent.ArticleId
				if art, err := s.articles.GetByID(ctx, *parent.ArticleId); err == nil {
					item["articleTitle"] = art.Title
				}
			}
		}
		list = append(list, item)
	}
	return map[string]interface{}{
		"list":       list,
		"pagination": userrepo.CalcNestPagination(total, pageSize, page),
	}, nil
}

func (s *ReplyService) enrichReplies(ctx context.Context, rows []*ent.Reply) ([]map[string]interface{}, error) {
	uids := make([]uint64, 0, len(rows)*2)
	for _, r := range rows {
		uids = append(uids, uint64(r.UID))
		if rid, err := strconv.Atoi(r.ReplyUid); err == nil && rid > 0 {
			uids = append(uids, uint64(rid))
		}
	}
	users, _ := s.users.GetUserBatch(ctx, uids)
	userMap := map[uint64]*usersvc.UserDTO{}
	for _, u := range users {
		if u != nil {
			userMap[u.ID] = u
		}
	}
	list := make([]map[string]interface{}, 0, len(rows))
	for _, r := range rows {
		item := map[string]interface{}{
			"id": r.ID, "parentId": r.ParentId, "replyUid": r.ReplyUid,
			"content": r.Content, "uid": r.UID, "status": r.Status,
			"createTime": r.CreateTime, "updateTime": r.UpdateTime,
			"userInfo":   util.UserInfoMap(userMap[uint64(r.UID)]),
		}
		if rid, err := strconv.Atoi(r.ReplyUid); err == nil && rid > 0 {
			item["tUserInfo"] = util.UserInfoMap(userMap[uint64(rid)])
		} else {
			item["tUserInfo"] = map[string]interface{}{}
		}
		list = append(list, item)
	}
	return list, nil
}
