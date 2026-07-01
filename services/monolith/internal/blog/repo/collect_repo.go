// collect_repo 收藏表 Ent 读写。
package repo

import (
	"context"
	"strconv"

	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent/collect"
)

// CollectRepo 收藏表读写。
type CollectRepo struct {
	client *ent.Client
}

// NewCollectRepo 构造 CollectRepo。
func NewCollectRepo(client *ent.Client) *CollectRepo {
	return &CollectRepo{client: client}
}

// FindByArticleAndUID 查找收藏记录。
func (r *CollectRepo) FindByArticleAndUID(ctx context.Context, articleID, uid int) (*ent.Collect, error) {
	return r.client.Collect.Query().
		Where(collect.ArticleIdEQ(articleID), collect.UIDEQ(uid)).
		First(ctx)
}

// Create 写入收藏。
func (r *CollectRepo) Create(ctx context.Context, row *ent.Collect) (*ent.Collect, error) {
	return r.client.Collect.Create().
		SetID(row.ID).
		SetArticleId(row.ArticleId).
		SetUID(row.UID).
		Save(ctx)
}

// DeleteByID 按收藏记录 id 删除。
func (r *CollectRepo) DeleteByID(ctx context.Context, id string) (int, error) {
	return r.client.Collect.Delete().Where(collect.IDEQ(id)).Exec(ctx)
}

// ListByUID 用户收藏分页列表。
func (r *CollectRepo) ListByUID(ctx context.Context, uid, page, pageSize int) ([]*ent.Collect, int, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	q := r.client.Collect.Query().Where(collect.UIDEQ(uid))
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	rows, err := q.Order(ent.Desc(collect.FieldCreateTime)).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		All(ctx)
	return rows, total, err
}

// CountByArticle 文章收藏总数。
func (r *CollectRepo) CountByArticle(ctx context.Context, articleID int) (int, error) {
	return r.client.Collect.Query().Where(collect.ArticleIdEQ(articleID)).Count(ctx)
}

// IsCollected 是否已收藏。
func (r *CollectRepo) IsCollected(ctx context.Context, articleID, uid int) (bool, error) {
	n, err := r.client.Collect.Query().
		Where(collect.ArticleIdEQ(articleID), collect.UIDEQ(uid)).
		Count(ctx)
	return n > 0, err
}

// ParseCollectArticleID 解析收藏 articleId（Nest 为数字字符串）。
func ParseCollectArticleID(v interface{}) (int, error) {
	switch n := v.(type) {
	case float64:
		return int(n), nil
	case int:
		return n, nil
	case string:
		return strconv.Atoi(n)
	default:
		return 0, strconv.ErrSyntax
	}
}
