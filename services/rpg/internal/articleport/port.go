// Package articleport RPG 对 blog 文章表的跨域读（共享单库 raw SQL，无 blog Ent）。
package articleport

import (
	"context"
	"database/sql"
	"time"
)

// ArticleSummary 打赏/校验所需文章字段。
type ArticleSummary struct {
	ID       int
	UID      int
	IsDelete bool
	Status   string
}

// ArticleReader 文章只读端口。
type ArticleReader interface {
	GetByID(ctx context.Context, id int) (*ArticleSummary, error)
	ListPublishedByAuthor(ctx context.Context, uid int) ([]map[string]interface{}, error)
}

// SQLArticleReader 经共享 MySQL 查询 x_article。
type SQLArticleReader struct {
	db *sql.DB
}

// NewSQLArticleReader 构造 ArticleReader。
func NewSQLArticleReader(db *sql.DB) *SQLArticleReader {
	return &SQLArticleReader{db: db}
}

// GetByID 按主键查文章。
func (r *SQLArticleReader) GetByID(ctx context.Context, id int) (*ArticleSummary, error) {
	row := r.db.QueryRowContext(ctx,
		"SELECT id, uid, isDelete, status FROM x_article WHERE id = ? LIMIT 1", id)
	var a ArticleSummary
	var isDelete int
	if err := row.Scan(&a.ID, &a.UID, &isDelete, &a.Status); err != nil {
		return nil, err
	}
	a.IsDelete = isDelete != 0
	return &a, nil
}

// ListPublishedByAuthor 作者已发布文章列表（公开主页）。
func (r *SQLArticleReader) ListPublishedByAuthor(ctx context.Context, uid int) ([]map[string]interface{}, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, title, createTime, topping FROM x_article
		 WHERE uid = ? AND isDelete = 0 AND status = 'publish'
		 ORDER BY topping DESC, createTime DESC, id DESC`, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id int
		var title string
		var createTime time.Time
		var topping int
		if err := rows.Scan(&id, &title, &createTime, &topping); err != nil {
			return nil, err
		}
		out = append(out, map[string]interface{}{
			"id": id, "title": title, "createTime": createTime, "topping": topping,
		})
	}
	return out, nil
}
