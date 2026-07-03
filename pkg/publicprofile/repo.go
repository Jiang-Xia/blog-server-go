// Package publicprofile 公开主页收藏/点赞文章列表（对齐 Nest profile.service）。
package publicprofile

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	_ "github.com/go-sql-driver/mysql"
)

const authorStatusActive = "active"

// ArticleRow 公开文章卡片字段（与 Nest mapPublicArticle 一致）。
type ArticleRow struct {
	ID             int
	Title          string
	Description    string
	Cover          string
	Views          int
	Likes          int
	ArticleLevel   int
	IsMasterpiece  int
	TipTotal       int
	CreateTime     time.Time
}

// Repo 收藏/点赞公开列表直查（共享 MySQL JOIN x_user 过滤锁定作者）。
type Repo struct {
	db     *sql.DB
	prefix string
}

// NewRepo 构造 Repo。
func NewRepo(cfg *config.Config) (*Repo, error) {
	db, err := sql.Open("mysql", cfg.MySQL.FormatDSN())
	if err != nil {
		return nil, fmt.Errorf("open mysql public profile: %w", err)
	}
	return &Repo{db: db, prefix: cfg.MySQL.TablePrefixOrDefault()}, nil
}

// Close 关闭连接池。
func (r *Repo) Close() {
	if r != nil && r.db != nil {
		_ = r.db.Close()
	}
}

// ListCollectArticles 用户公开收藏文章分页。
func (r *Repo) ListCollectArticles(ctx context.Context, uid, page, pageSize int) ([]ArticleRow, int, error) {
	page, pageSize = normalizePage(page, pageSize)
	base := fmt.Sprintf(`
FROM %scollect col
INNER JOIN %sarticle article ON article.id = col.articleId
INNER JOIN %suser author ON author.id = article.uid
WHERE col.uid = ?
  AND article.isDelete = 0
  AND article.status = 'publish'
  AND author.status = ?
  AND author.isDelete = 0`, r.prefix, r.prefix, r.prefix)

	total, err := r.count(ctx, "SELECT COUNT(*) "+base, uid, authorStatusActive)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.queryArticles(ctx,
		`SELECT article.id, article.title, article.description, article.cover,
		        article.views, article.likes, article.articleLevel, article.isMasterpiece,
		        article.tipTotal, article.createTime `+base+`
ORDER BY col.createTime DESC
LIMIT ? OFFSET ?`,
		uid, authorStatusActive, pageSize, (page-1)*pageSize)
	return rows, total, err
}

// ListLikeArticles 用户公开点赞文章分页（仅 status=1）。
func (r *Repo) ListLikeArticles(ctx context.Context, uid, page, pageSize int) ([]ArticleRow, int, error) {
	page, pageSize = normalizePage(page, pageSize)
	base := fmt.Sprintf(`
FROM %slike lk
INNER JOIN %sarticle article ON article.id = lk.articleId
INNER JOIN %suser author ON author.id = article.uid
WHERE lk.uid = ?
  AND lk.status = '1'
  AND article.isDelete = 0
  AND article.status = 'publish'
  AND author.status = ?
  AND author.isDelete = 0`, r.prefix, r.prefix, r.prefix)

	total, err := r.count(ctx, "SELECT COUNT(*) "+base, uid, authorStatusActive)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.queryArticles(ctx,
		`SELECT article.id, article.title, article.description, article.cover,
		        article.views, article.likes, article.articleLevel, article.isMasterpiece,
		        article.tipTotal, article.createTime `+base+`
ORDER BY lk.id DESC
LIMIT ? OFFSET ?`,
		uid, authorStatusActive, pageSize, (page-1)*pageSize)
	return rows, total, err
}

func (r *Repo) count(ctx context.Context, q string, args ...any) (int, error) {
	var n int
	if err := r.db.QueryRowContext(ctx, q, args...).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

func (r *Repo) queryArticles(ctx context.Context, q string, args ...any) ([]ArticleRow, error) {
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]ArticleRow, 0)
	for rows.Next() {
		var row ArticleRow
		if err := rows.Scan(
			&row.ID, &row.Title, &row.Description, &row.Cover,
			&row.Views, &row.Likes, &row.ArticleLevel, &row.IsMasterpiece,
			&row.TipTotal, &row.CreateTime,
		); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func normalizePage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	return page, pageSize
}
