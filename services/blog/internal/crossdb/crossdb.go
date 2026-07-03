// Package crossdb 共享 MySQL 跨域表直查（user 域 operation_log、sensitive_word_hit 等）。
package crossdb

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	_ "github.com/go-sql-driver/mysql"
)

// CrossDB 共享库跨表直查。
type CrossDB struct {
	db *sql.DB
}

// New 构造 CrossDB。
func New(cfg *config.Config) (*CrossDB, error) {
	db, err := sql.Open("mysql", cfg.MySQL.FormatDSN())
	if err != nil {
		return nil, fmt.Errorf("open mysql cross db: %w", err)
	}
	return &CrossDB{db: db}, nil
}

// Close 关闭连接池。
func (c *CrossDB) Close() {
	if c != nil && c.db != nil {
		_ = c.db.Close()
	}
}

// DeleteOldOperationLogs 删除过期操作日志。
func (c *CrossDB) DeleteOldOperationLogs(ctx context.Context, before interface{}) (int, error) {
	res, err := c.db.ExecContext(ctx, "DELETE FROM x_operation_log WHERE createTime < ?", before)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return int(n), nil
}

// QueryPendingSensitiveHits 查询 pending 敏感词命中。
func (c *CrossDB) QueryPendingSensitiveHits(ctx context.Context) ([]struct {
	HitWords, SourceType string
}, error) {
	rows, err := c.db.QueryContext(ctx,
		"SELECT hitWords, sourceType FROM x_sensitive_word_hit WHERE status = ? AND isDelete = 0", "pending")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []struct {
		HitWords, SourceType string
	}
	for rows.Next() {
		var h struct {
			HitWords, SourceType string
		}
		if err := rows.Scan(&h.HitWords, &h.SourceType); err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, rows.Err()
}

// CountPendingSensitiveHits 统计 pending 命中数。
func (c *CrossDB) CountPendingSensitiveHits(ctx context.Context) (int, error) {
	var n int
	err := c.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM x_sensitive_word_hit WHERE status = ? AND isDelete = 0", "pending").Scan(&n)
	return n, err
}

// DeleteOldTaskLogs 删除过期任务执行日志。
func (c *CrossDB) DeleteOldTaskLogs(ctx context.Context, before interface{}) (int, error) {
	res, err := c.db.ExecContext(ctx, "DELETE FROM x_scheduled_task_log WHERE createTime < ?", before)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return int(n), nil
}
