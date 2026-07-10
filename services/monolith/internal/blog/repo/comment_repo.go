// comment_repo 评论表 Ent 读写。
package repo

import (
	"context"

	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent/comment"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent/reply"
)

// CommentRepo 评论表读写。
type CommentRepo struct {
	client *ent.Client
}

// NewCommentRepo 构造 CommentRepo。
func NewCommentRepo(client *ent.Client) *CommentRepo {
	return &CommentRepo{client: client}
}

// CommentFilter 评论列表筛选。
type CommentFilter struct {
	ArticleID *int
	UID       *int
	ArticleIDs []int
	Status     string
	Content   string
	Page      int
	PageSize  int
	SortAsc   bool
}

// Create 写入评论。
func (r *CommentRepo) Create(ctx context.Context, row *ent.Comment) (*ent.Comment, error) {
	b := r.client.Comment.Create().
		SetID(row.ID).
		SetContent(row.Content).
		SetUID(row.UID).
		SetStatus(row.Status)
	if row.UserId != nil {
		b.SetUserId(*row.UserId)
	}
	if row.ArticleId != nil {
		b.SetArticleId(*row.ArticleId)
	}
	return b.Save(ctx)
}

// Delete 物理删除评论。
func (r *CommentRepo) Delete(ctx context.Context, id string) error {
	_, err := r.client.Comment.Delete().Where(comment.IDEQ(id)).Exec(ctx)
	return err
}

// GetByID 按主键查询。
func (r *CommentRepo) GetByID(ctx context.Context, id string) (*ent.Comment, error) {
	return r.client.Comment.Query().Where(comment.IDEQ(id)).Only(ctx)
}

// UpdateStatus 更新审核状态。
func (r *CommentRepo) UpdateStatus(ctx context.Context, id, status string) error {
	_, err := r.client.Comment.UpdateOneID(id).SetStatus(status).Save(ctx)
	return err
}

// List 分页查询评论，支持按文章/用户/状态筛选。
func (r *CommentRepo) List(ctx context.Context, f CommentFilter) ([]*ent.Comment, int, error) {
	page, pageSize := f.Page, f.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	q := r.client.Comment.Query()
	if f.ArticleID != nil {
		q = q.Where(comment.ArticleIdEQ(*f.ArticleID))
	}
	if f.UID != nil {
		q = q.Where(comment.UIDEQ(*f.UID))
	}
	if f.Status != "" {
		q = q.Where(comment.StatusEQ(f.Status))
	}
	if f.Content != "" {
		q = q.Where(comment.ContentContains(f.Content))
	}
	if len(f.ArticleIDs) > 0 {
		q = q.Where(comment.ArticleIdIn(f.ArticleIDs...))
	}
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	if f.SortAsc {
		q = q.Order(ent.Asc(comment.FieldCreateTime))
	} else {
		q = q.Order(ent.Desc(comment.FieldCreateTime))
	}
	rows, err := q.Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	return rows, total, err
}

// CountByArticleIDs 批量统计文章已审核评论数。
func (r *CommentRepo) CountByArticleIDs(ctx context.Context, articleIDs []int) (map[int]int, error) {
	out := make(map[int]int, len(articleIDs))
	if len(articleIDs) == 0 {
		return out, nil
	}
	type row struct {
		ArticleID int `json:"articleId"`
		Count     int `json:"count"`
	}
	var rows []row
	err := r.client.Comment.Query().
		Where(
			comment.ArticleIdIn(articleIDs...),
			comment.StatusEQ("approved"),
		).
		GroupBy(comment.FieldArticleId).
		Aggregate(ent.Count()).
		Scan(ctx, &rows)
	if err != nil {
		return nil, err
	}
	for _, item := range rows {
		out[item.ArticleID] = item.Count
	}
	return out, nil
}

// CountDiscussionTotal 统计文章集合下已通过审核的评论与回复总数（对齐 Nest findComment 汇总）。
func (r *CommentRepo) CountDiscussionTotal(ctx context.Context, articleIDs []int) (int, error) {
	if len(articleIDs) == 0 {
		return 0, nil
	}
	commentCount, err := r.client.Comment.Query().
		Where(comment.ArticleIdIn(articleIDs...), comment.StatusEQ("approved")).
		Count(ctx)
	if err != nil {
		return 0, err
	}
	parentIDs, err := r.client.Comment.Query().
		Where(comment.ArticleIdIn(articleIDs...), comment.StatusEQ("approved")).
		IDs(ctx)
	if err != nil {
		return 0, err
	}
	if len(parentIDs) == 0 {
		return commentCount, nil
	}
	replyCount, err := r.client.Reply.Query().
		Where(reply.StatusEQ("approved"), reply.ParentIdIn(parentIDs...)).
		Count(ctx)
	if err != nil {
		return 0, err
	}
	return commentCount + replyCount, nil
}

