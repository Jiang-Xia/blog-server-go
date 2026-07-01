// Package repo 封装 blog 域 Ent 数据访问（仅 article/category/tag 表，禁止跨域 JOIN）。
package repo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/services/blog/ent"
	"github.com/Jiang-Xia/blog-server-go/services/blog/ent/article"
	"github.com/Jiang-Xia/blog-server-go/services/blog/ent/predicate"
)

// ArticleRepo 文章表读写。
type ArticleRepo struct {
	client   *ent.Client
	tagJunc  *articleTagJunction
}

// NewArticleRepo 构造 ArticleRepo。
func NewArticleRepo(client *ent.Client, cfg *config.Config) (*ArticleRepo, error) {
	junc, err := newArticleTagJunction(cfg)
	if err != nil {
		return nil, err
	}
	return &ArticleRepo{client: client, tagJunc: junc}, nil
}

// ListFilter 列表查询条件。
type ListFilter struct {
	Page              int
	PageSize          int
	CategoryID        string
	TagIDs            []string
	Title             string
	Description       string
	Content           string
	SortAsc           bool
	Client            bool
	AuthorUIDs        []int
	DeptIDs           []int
	DeptID            *int
	UID               *int
	OnlyNotDeleted    bool
	Status            string
	ExcludeID         int
}

// List 分页查询文章。
func (r *ArticleRepo) List(ctx context.Context, f ListFilter) ([]*ent.Article, int, error) {
	page, pageSize := f.Page, f.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	q := r.client.Article.Query()
	tagArticleIDs, err := r.articleIDsByTags(ctx, f.TagIDs)
	if err != nil {
		return nil, 0, err
	}
	q = applyArticleFilters(q, f, tagArticleIDs)

	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	order := ent.Desc(article.FieldTopping)
	if f.SortAsc {
		q = q.Order(order, ent.Asc(article.FieldCreateTime))
	} else {
		q = q.Order(order, ent.Desc(article.FieldCreateTime))
	}

	rows, err := q.Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func applyArticleFilters(q *ent.ArticleQuery, f ListFilter, tagArticleIDs []int) *ent.ArticleQuery {
	if f.OnlyNotDeleted || f.Client {
		q = q.Where(article.IsDeleteEQ(false))
	}
	if f.Client {
		q = q.Where(article.StatusEQ("publish"))
	}
	if f.Status != "" {
		q = q.Where(article.StatusEQ(f.Status))
	}
	if f.CategoryID != "" {
		q = q.Where(article.ArticlesEQ(f.CategoryID))
	}
	if len(f.TagIDs) > 0 {
		if len(tagArticleIDs) == 0 {
			q = q.Where(article.IDIn(-1))
		} else {
			q = q.Where(article.IDIn(tagArticleIDs...))
		}
	}
	if f.UID != nil {
		q = q.Where(article.UIDEQ(*f.UID))
	}
	if len(f.AuthorUIDs) > 0 {
		q = q.Where(article.UIDIn(f.AuthorUIDs...))
	}
	if f.DeptID != nil {
		q = q.Where(article.DeptIdEQ(*f.DeptID))
	} else if len(f.DeptIDs) > 0 {
		q = q.Where(article.DeptIdIn(f.DeptIDs...))
	}
	if f.ExcludeID > 0 {
		q = q.Where(article.IDNEQ(f.ExcludeID))
	}

	var orPreds []predicate.Article
	if f.Title != "" {
		orPreds = append(orPreds, article.TitleContains(f.Title))
	}
	if f.Description != "" {
		orPreds = append(orPreds, article.DescriptionContains(f.Description))
	}
	if f.Content != "" {
		orPreds = append(orPreds, article.ContentContains(f.Content))
	}
	if len(orPreds) > 0 {
		q = q.Where(article.Or(orPreds...))
	}
	return q
}

// articleIDsByTags 查询包含任一指定标签的文章 ID。
func (r *ArticleRepo) articleIDsByTags(ctx context.Context, tagIDs []string) ([]int, error) {
	return r.tagJunc.articleIDsByTagIDs(ctx, tagIDs)
}

// GetByID 按主键查询。
func (r *ArticleRepo) GetByID(ctx context.Context, id int) (*ent.Article, error) {
	return r.client.Article.Query().Where(article.IDEQ(id)).Only(ctx)
}

// FindByIDOrTitle 按 id 或 title 查询（Nest findById 兼容）。
func (r *ArticleRepo) FindByIDOrTitle(ctx context.Context, key string) (*ent.Article, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil, &ent.NotFoundError{}
	}
	if id, err := parseInt(key); err == nil {
		row, err := r.client.Article.Query().Where(article.IDEQ(id)).Only(ctx)
		if err == nil {
			return row, nil
		}
		if !ent.IsNotFound(err) {
			return nil, err
		}
	}
	return r.client.Article.Query().Where(article.TitleEQ(key)).Only(ctx)
}

// ExistsByTitle 检查标题是否已存在。
func (r *ArticleRepo) ExistsByTitle(ctx context.Context, title string, excludeID int) (bool, error) {
	q := r.client.Article.Query().Where(article.TitleEQ(title))
	if excludeID > 0 {
		q = q.Where(article.IDNEQ(excludeID))
	}
	return q.Exist(ctx)
}

// Create 创建文章。
func (r *ArticleRepo) Create(ctx context.Context, row *ent.Article) (*ent.Article, error) {
	b := r.client.Article.Create().
		SetUID(row.UID).
		SetTitle(row.Title).
		SetDescription(row.Description).
		SetContent(row.Content).
		SetContentHtml(row.ContentHtml).
		SetCover(row.Cover).
		SetStatus(row.Status).
		SetTopping(row.Topping).
		SetViews(row.Views).
		SetLikes(row.Likes).
		SetUTime(row.UTime)
	if row.DeptId != nil {
		b = b.SetDeptId(*row.DeptId)
	}
	if row.Articles != nil {
		b = b.SetArticles(*row.Articles)
	}
	if row.ScheduledPublishAt != nil {
		b = b.SetScheduledPublishAt(*row.ScheduledPublishAt)
	}
	return b.Save(ctx)
}

// Update 更新文章字段。
func (r *ArticleRepo) Update(ctx context.Context, id int, fn func(*ent.ArticleUpdateOne) *ent.ArticleUpdateOne) (*ent.Article, error) {
	up := r.client.Article.UpdateOneID(id)
	return fn(up).Save(ctx)
}

// UpdateFields 批量更新指定字段 map。
func (r *ArticleRepo) UpdateFields(ctx context.Context, id int, fields map[string]interface{}) (*ent.Article, error) {
	up := r.client.Article.UpdateOneID(id)
	for k, v := range fields {
		switch k {
		case "title":
			up = up.SetTitle(v.(string))
		case "description":
			up = up.SetDescription(v.(string))
		case "content":
			up = up.SetContent(v.(string))
		case "contentHtml":
			up = up.SetContentHtml(v.(string))
		case "cover":
			up = up.SetCover(v.(string))
		case "status":
			up = up.SetStatus(v.(string))
		case "isDelete":
			up = up.SetIsDelete(v.(bool))
		case "topping":
			up = up.SetTopping(v.(int))
		case "views":
			up = up.SetViews(v.(int))
		case "likes":
			up = up.SetLikes(v.(int))
		case "uTime":
			up = up.SetUTime(v.(string))
		case "articles":
			up = up.SetArticles(v.(string))
		case "scheduledPublishAt":
			if v == nil {
				up = up.ClearScheduledPublishAt()
			} else {
				up = up.SetScheduledPublishAt(v.(time.Time))
			}
		}
	}
	return up.Save(ctx)
}

// HardDelete 物理删除文章。
func (r *ArticleRepo) HardDelete(ctx context.Context, id int) error {
	return r.client.Article.DeleteOneID(id).Exec(ctx)
}

// IncrementViews 阅读量 +1。
func (r *ArticleRepo) IncrementViews(ctx context.Context, id int) (bool, error) {
	row, err := r.GetByID(ctx, id)
	if err != nil {
		return false, err
	}
	if row.IsDelete {
		return false, nil
	}
	_, err = r.client.Article.UpdateOneID(id).SetViews(row.Views + 1).Save(ctx)
	return err == nil, err
}

// ListPublishedByAuthor 查询作者已发布文章（导航用）。
func (r *ArticleRepo) ListPublishedByAuthor(ctx context.Context, uid int) ([]*ent.Article, error) {
	return r.client.Article.Query().
		Where(
			article.UIDEQ(uid),
			article.IsDeleteEQ(false),
			article.StatusEQ("publish"),
		).
		Order(ent.Desc(article.FieldTopping), ent.Desc(article.FieldCreateTime), ent.Desc(article.FieldID)).
		Select(article.FieldID, article.FieldTitle, article.FieldCreateTime, article.FieldTopping).
		All(ctx)
}

// ListPublishedAll 查询全部已发布未删除文章（归档/统计）。
func (r *ArticleRepo) ListPublishedAll(ctx context.Context) ([]*ent.Article, error) {
	return r.client.Article.Query().
		Where(article.IsDeleteEQ(false), article.StatusEQ("publish")).
		Order(ent.Desc(article.FieldCreateTime)).
		All(ctx)
}

// ListRelated 同作者已发布文章。
func (r *ArticleRepo) ListRelated(ctx context.Context, uid, excludeID, limit int) ([]*ent.Article, error) {
	if limit <= 0 {
		limit = 6
	}
	return r.client.Article.Query().
		Where(
			article.UIDEQ(uid),
			article.IsDeleteEQ(false),
			article.StatusEQ("publish"),
			article.IDNEQ(excludeID),
		).
		Order(ent.Desc(article.FieldCreateTime)).
		Limit(limit).
		All(ctx)
}

// ListLatestPublished 全站最新已发布文章。
func (r *ArticleRepo) ListLatestPublished(ctx context.Context, excludeID, limit int) ([]*ent.Article, error) {
	q := r.client.Article.Query().
		Where(article.IsDeleteEQ(false), article.StatusEQ("publish"))
	if excludeID > 0 {
		q = q.Where(article.IDNEQ(excludeID))
	}
	return q.Order(ent.Desc(article.FieldCreateTime)).Limit(limit).All(ctx)
}

// CountByAuthor 统计作者文章数。
func (r *ArticleRepo) CountByAuthor(ctx context.Context, uid int, status string, includeDeleted bool) (int, error) {
	q := r.client.Article.Query().Where(article.UIDEQ(uid))
	if status != "" {
		q = q.Where(article.StatusEQ(status))
	}
	if !includeDeleted {
		q = q.Where(article.IsDeleteEQ(false))
	}
	return q.Count(ctx)
}

// SumViewsByAuthor 汇总作者阅读量。
func (r *ArticleRepo) SumViewsByAuthor(ctx context.Context, uid int) (int, error) {
	var sum struct {
		Sum sql.NullInt64
	}
	err := r.client.Article.Query().
		Where(article.UIDEQ(uid), article.IsDeleteEQ(false)).
		Aggregate(ent.Sum(article.FieldViews)).
		Scan(ctx, &sum)
	if err != nil {
		return 0, err
	}
	if sum.Sum.Valid {
		return int(sum.Sum.Int64), nil
	}
	return 0, nil
}

// TopByAuthor 作者热门文章 Top N。
func (r *ArticleRepo) TopByAuthor(ctx context.Context, uid, limit int) ([]*ent.Article, error) {
	return r.client.Article.Query().
		Where(article.UIDEQ(uid), article.IsDeleteEQ(false)).
		Order(ent.Desc(article.FieldViews)).
		Limit(limit).
		Select(article.FieldID, article.FieldTitle, article.FieldViews, article.FieldLikes, article.FieldCreateTime).
		All(ctx)
}

// SetCategory 更新文章分类 FK。
func (r *ArticleRepo) SetCategory(ctx context.Context, id int, categoryID string) error {
	_, err := r.client.Article.UpdateOneID(id).SetArticles(categoryID).Save(ctx)
	return err
}

// --- article_tags_tag ---

// ListTagIDsByArticle 查询文章关联标签 ID。
func (r *ArticleRepo) ListTagIDsByArticle(ctx context.Context, articleID int) ([]string, error) {
	return r.tagJunc.tagIDsByArticleID(ctx, articleID)
}

// ListTagIDsByArticles 批量查询文章标签。
func (r *ArticleRepo) ListTagIDsByArticles(ctx context.Context, articleIDs []int) (map[int][]string, error) {
	return r.tagJunc.tagIDsByArticleIDs(ctx, articleIDs)
}

// ReplaceTags 替换文章标签关联。
func (r *ArticleRepo) ReplaceTags(ctx context.Context, articleID int, tagIDs []string) error {
	return r.tagJunc.replaceTags(ctx, articleID, tagIDs)
}

// CountByCategory 统计分类下文章数。
func (r *ArticleRepo) CountByCategory(ctx context.Context, categoryID string, publishedOnly bool) (int, error) {
	q := r.client.Article.Query().Where(article.ArticlesEQ(categoryID), article.IsDeleteEQ(false))
	if publishedOnly {
		q = q.Where(article.StatusEQ("publish"))
	}
	return q.Count(ctx)
}

// CountByTag 统计标签下文章数。
func (r *ArticleRepo) CountByTag(ctx context.Context, tagID string, publishedOnly bool) (int, error) {
	ids, err := r.tagJunc.articleIDsByTagID(ctx, tagID)
	if err != nil {
		return 0, err
	}
	if len(ids) == 0 {
		return 0, nil
	}
	q := r.client.Article.Query().Where(article.IDIn(ids...), article.IsDeleteEQ(false))
	if publishedOnly {
		q = q.Where(article.StatusEQ("publish"))
	}
	return q.Count(ctx)
}

// ListByTagID 查询标签关联文章（可选 status 过滤）。
func (r *ArticleRepo) ListByTagID(ctx context.Context, tagID, status string) ([]*ent.Article, error) {
	ids, err := r.tagJunc.articleIDsByTagID(ctx, tagID)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return []*ent.Article{}, nil
	}
	q := r.client.Article.Query().
		Where(article.IDIn(ids...)).
		Order(ent.Desc(article.FieldUpdateTime))
	if status != "" {
		q = q.Where(article.StatusEQ(status))
	}
	return q.All(ctx)
}

func parseInt(s string) (int, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}
