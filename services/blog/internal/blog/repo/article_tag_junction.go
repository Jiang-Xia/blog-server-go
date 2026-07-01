package repo

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	_ "github.com/go-sql-driver/mysql"
)

// articleTagJunction 直查 x_article_tags_tag（复合主键无 id 列，Ent 无法映射）。
type articleTagJunction struct {
	db    *sql.DB
	table string
}

func newArticleTagJunction(cfg *config.Config) (*articleTagJunction, error) {
	db, err := sql.Open("mysql", cfg.MySQL.FormatDSN())
	if err != nil {
		return nil, fmt.Errorf("open mysql for article_tags_tag: %w", err)
	}
	return &articleTagJunction{
		db:    db,
		table: cfg.MySQL.TablePrefixOrDefault() + "article_tags_tag",
	}, nil
}

func (j *articleTagJunction) articleIDsByTagIDs(ctx context.Context, tagIDs []string) ([]int, error) {
	if len(tagIDs) == 0 {
		return nil, nil
	}
	ph, args := inPlaceholdersStrings(tagIDs)
	q := fmt.Sprintf("SELECT DISTINCT articleId FROM %s WHERE tagId IN (%s)", j.table, ph)
	rows, err := j.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanIntColumn(rows)
}

func (j *articleTagJunction) articleIDsByTagID(ctx context.Context, tagID string) ([]int, error) {
	rows, err := j.db.QueryContext(ctx,
		fmt.Sprintf("SELECT articleId FROM %s WHERE tagId = ?", j.table), tagID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanIntColumn(rows)
}

func (j *articleTagJunction) tagIDsByArticleID(ctx context.Context, articleID int) ([]string, error) {
	rows, err := j.db.QueryContext(ctx,
		fmt.Sprintf("SELECT tagId FROM %s WHERE articleId = ?", j.table), articleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

func (j *articleTagJunction) tagIDsByArticleIDs(ctx context.Context, articleIDs []int) (map[int][]string, error) {
	if len(articleIDs) == 0 {
		return map[int][]string{}, nil
	}
	ph, args := inPlaceholders(articleIDs)
	q := fmt.Sprintf("SELECT articleId, tagId FROM %s WHERE articleId IN (%s)", j.table, ph)
	rows, err := j.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[int][]string, len(articleIDs))
	for rows.Next() {
		var articleID int
		var tagID string
		if err := rows.Scan(&articleID, &tagID); err != nil {
			return nil, err
		}
		out[articleID] = append(out[articleID], tagID)
	}
	return out, rows.Err()
}

func (j *articleTagJunction) replaceTags(ctx context.Context, articleID int, tagIDs []string) error {
	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx,
		fmt.Sprintf("DELETE FROM %s WHERE articleId = ?", j.table), articleID); err != nil {
		return err
	}
	ins := fmt.Sprintf("INSERT INTO %s (articleId, tagId) VALUES (?, ?)", j.table)
	for _, tagID := range tagIDs {
		if strings.TrimSpace(tagID) == "" {
			continue
		}
		if _, err := tx.ExecContext(ctx, ins, articleID, tagID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func inPlaceholders(ids []int) (string, []any) {
	ph := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		ph[i] = "?"
		args[i] = id
	}
	return strings.Join(ph, ","), args
}

func inPlaceholdersStrings(items []string) (string, []any) {
	ph := make([]string, len(items))
	args := make([]any, len(items))
	for i, v := range items {
		ph[i] = "?"
		args[i] = v
	}
	return strings.Join(ph, ","), args
}

func scanIntColumn(rows *sql.Rows) ([]int, error) {
	var out []int
	for rows.Next() {
		var n int
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}
