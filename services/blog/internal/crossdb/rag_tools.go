package crossdb

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// RagArticleItem Tool 用文章列表项。
type RagArticleItem struct {
	ArticleID     int
	Title         string
	URL           string
	Views         int
	Likes         int
	ArticleLevel  int
	IsMasterpiece int
	CreateTime    time.Time
	Comments      *int
	Collects      *int
}

const publishedArticleJoin = `
FROM x_article a
INNER JOIN x_user u ON u.id = a.uid
WHERE a.isDelete = 0 AND a.status = 'publish'
  AND u.status = 'active' AND u.isDelete = 0
`

// RagArticleRanking 全站文章排行。
func (c *CrossDB) RagArticleRanking(ctx context.Context, metric string, limit int) ([]RagArticleItem, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 20 {
		limit = 20
	}
	var q string
	switch metric {
	case "comments":
		q = fmt.Sprintf(`SELECT a.id, a.title, a.views, a.likes,
			(SELECT COUNT(*) FROM x_comment cm WHERE cm.articleId = a.id) AS extra
			%s ORDER BY extra DESC LIMIT ?`, publishedArticleJoin)
	case "collects":
		q = fmt.Sprintf(`SELECT a.id, a.title, a.views, a.likes,
			(SELECT COUNT(*) FROM x_collect col WHERE col.articleId = a.id) AS extra
			%s ORDER BY extra DESC LIMIT ?`, publishedArticleJoin)
	case "likes":
		q = fmt.Sprintf(`SELECT a.id, a.title, a.views, a.likes, 0 AS extra
			%s ORDER BY a.likes DESC LIMIT ?`, publishedArticleJoin)
	default:
		metric = "views"
		q = fmt.Sprintf(`SELECT a.id, a.title, a.views, a.likes, 0 AS extra
			%s ORDER BY a.views DESC LIMIT ?`, publishedArticleJoin)
	}
	rows, err := c.db.QueryContext(ctx, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []RagArticleItem
	for rows.Next() {
		var item RagArticleItem
		var extra int
		if err := rows.Scan(&item.ArticleID, &item.Title, &item.Views, &item.Likes, &extra); err != nil {
			return nil, err
		}
		item.URL = fmt.Sprintf("/detail/%d", item.ArticleID)
		if metric == "comments" {
			item.Comments = intPtr(extra)
		}
		if metric == "collects" {
			item.Collects = intPtr(extra)
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

// RagAuthorRow 作者统计行。
type RagAuthorRow struct {
	UID          int
	Nickname     string
	ArticleCount int
}

// RagListAuthors 有已发布文章的作者列表。
func (c *CrossDB) RagListAuthors(ctx context.Context, page, pageSize int) ([]RagAuthorRow, int, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 50 {
		pageSize = 50
	}
	countQ := fmt.Sprintf(`SELECT COUNT(*) FROM (
		SELECT u.id %s GROUP BY u.id, u.nickname
	) t`, publishedArticleJoin)
	var total int
	if err := c.db.QueryRowContext(ctx, countQ).Scan(&total); err != nil {
		return nil, 0, err
	}
	q := fmt.Sprintf(`SELECT u.id, u.nickname, COUNT(a.id) AS cnt
		%s GROUP BY u.id, u.nickname ORDER BY cnt DESC LIMIT ? OFFSET ?`, publishedArticleJoin)
	rows, err := c.db.QueryContext(ctx, q, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []RagAuthorRow
	for rows.Next() {
		var row RagAuthorRow
		if err := rows.Scan(&row.UID, &row.Nickname, &row.ArticleCount); err != nil {
			return nil, 0, err
		}
		out = append(out, row)
	}
	return out, total, rows.Err()
}

// RagAuthorStats 作者发文汇总。
func (c *CrossDB) RagAuthorStats(ctx context.Context, uid int, nickname string) (map[string]interface{}, error) {
	var authorUID int
	var authorNick string
	if uid > 0 {
		err := c.db.QueryRowContext(ctx,
			`SELECT id, nickname FROM x_user WHERE id = ? AND isDelete = 0 AND status = 'active'`, uid).
			Scan(&authorUID, &authorNick)
		if err == sql.ErrNoRows {
			return map[string]interface{}{"error": "未找到作者"}, nil
		}
		if err != nil {
			return nil, err
		}
	} else if strings.TrimSpace(nickname) != "" {
		err := c.db.QueryRowContext(ctx,
			`SELECT id, nickname FROM x_user WHERE nickname LIKE ? AND isDelete = 0 AND status = 'active' ORDER BY id ASC LIMIT 1`,
			"%"+strings.TrimSpace(nickname)+"%").Scan(&authorUID, &authorNick)
		if err == sql.ErrNoRows {
			return map[string]interface{}{"error": "未找到作者"}, nil
		}
		if err != nil {
			return nil, err
		}
	} else {
		return map[string]interface{}{"error": "未找到作者"}, nil
	}

	q := fmt.Sprintf(`SELECT a.id, a.views, a.likes %s AND a.uid = ?`, publishedArticleJoin)
	rows, err := c.db.QueryContext(ctx, q, authorUID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var articleIDs []int
	totalViews, totalLikes := 0, 0
	for rows.Next() {
		var id, views, likes int
		if err := rows.Scan(&id, &views, &likes); err != nil {
			return nil, err
		}
		articleIDs = append(articleIDs, id)
		totalViews += views
		totalLikes += likes
	}
	totalComments, totalCollects := 0, 0
	if len(articleIDs) > 0 {
		ph, args := ragInPlaceholdersInts(articleIDs)
		_ = c.db.QueryRowContext(ctx,
			fmt.Sprintf(`SELECT COUNT(*) FROM x_comment WHERE articleId IN (%s)`, ph), args...).
			Scan(&totalComments)
		_ = c.db.QueryRowContext(ctx,
			fmt.Sprintf(`SELECT COUNT(*) FROM x_collect WHERE articleId IN (%s)`, ph), args...).
			Scan(&totalCollects)
	}
	return map[string]interface{}{
		"uid": authorUID, "nickname": authorNick, "profileUrl": fmt.Sprintf("/user/%d", authorUID),
		"articleCount": len(articleIDs), "totalViews": totalViews, "totalLikes": totalLikes,
		"totalComments": totalComments, "totalCollects": totalCollects,
	}, nil
}

// RagCategoryStat 分类统计。
type RagCategoryStat struct {
	Category     string
	ArticleCount int
}

// RagCategoryStats 各分类文章数。
func (c *CrossDB) RagCategoryStats(ctx context.Context, limit int) ([]RagCategoryStat, error) {
	if limit <= 0 {
		limit = 20
	}
	q := `SELECT cat.label, COUNT(a.id) AS cnt
		FROM x_article a
		INNER JOIN x_user u ON u.id = a.uid
		LEFT JOIN x_category cat ON cat.id = a.articles
		WHERE a.isDelete = 0 AND a.status = 'publish'
		  AND u.status = 'active' AND u.isDelete = 0
		  AND cat.label IS NOT NULL
		GROUP BY cat.id, cat.label
		ORDER BY cnt DESC LIMIT ?`
	rows, err := c.db.QueryContext(ctx, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []RagCategoryStat
	for rows.Next() {
		var row RagCategoryStat
		if err := rows.Scan(&row.Category, &row.ArticleCount); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

// RagTagStat 标签统计。
type RagTagStat struct {
	Tag          string
	ArticleCount int
}

// RagTagCloud 热门标签。
func (c *CrossDB) RagTagCloud(ctx context.Context, limit int) ([]RagTagStat, error) {
	if limit <= 0 {
		limit = 20
	}
	q := `SELECT t.label, COUNT(DISTINCT a.id) AS cnt
		FROM x_article a
		INNER JOIN x_user u ON u.id = a.uid
		INNER JOIN x_article_tags_tag j ON j.articleId = a.id
		INNER JOIN x_tag t ON t.id = j.tagId
		WHERE a.isDelete = 0 AND a.status = 'publish'
		  AND u.status = 'active' AND u.isDelete = 0
		GROUP BY t.id, t.label
		ORDER BY cnt DESC LIMIT ?`
	rows, err := c.db.QueryContext(ctx, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []RagTagStat
	for rows.Next() {
		var row RagTagStat
		if err := rows.Scan(&row.Tag, &row.ArticleCount); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

// RagArchiveStat 归档统计。
type RagArchiveStat struct {
	Year         int
	ArticleCount int
}

// RagArchiveStats 按年统计发文。
func (c *CrossDB) RagArchiveStats(ctx context.Context, year *int) ([]RagArchiveStat, error) {
	q := fmt.Sprintf(`SELECT YEAR(a.createTime) AS y, COUNT(a.id) AS cnt %s`, publishedArticleJoin)
	var args []interface{}
	if year != nil {
		q += ` AND YEAR(a.createTime) = ?`
		args = append(args, *year)
	}
	q += ` GROUP BY YEAR(a.createTime) ORDER BY y DESC`
	rows, err := c.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []RagArchiveStat
	for rows.Next() {
		var row RagArchiveStat
		if err := rows.Scan(&row.Year, &row.ArticleCount); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

// RagRecentArticles 最近发布文章。
func (c *CrossDB) RagRecentArticles(ctx context.Context, limit int) ([]RagArticleItem, error) {
	if limit <= 0 {
		limit = 10
	}
	q := fmt.Sprintf(`SELECT a.id, a.title, a.views, a.likes, a.articleLevel, a.isMasterpiece, a.createTime
		%s ORDER BY a.createTime DESC LIMIT ?`, publishedArticleJoin)
	rows, err := c.db.QueryContext(ctx, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRagArticleRows(rows)
}

// RagMasterpieceArticles 神作列表。
func (c *CrossDB) RagMasterpieceArticles(ctx context.Context, limit int) ([]RagArticleItem, error) {
	if limit <= 0 {
		limit = 10
	}
	q := fmt.Sprintf(`SELECT a.id, a.title, a.views, a.likes, a.articleLevel, a.isMasterpiece, a.createTime
		%s AND (a.isMasterpiece = 1 OR a.articleLevel >= 10)
		ORDER BY a.articleLevel DESC, a.views DESC LIMIT ?`, publishedArticleJoin)
	rows, err := c.db.QueryContext(ctx, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRagArticleRows(rows)
}

func scanRagArticleRows(rows *sql.Rows) ([]RagArticleItem, error) {
	var out []RagArticleItem
	for rows.Next() {
		var item RagArticleItem
		if err := rows.Scan(&item.ArticleID, &item.Title, &item.Views, &item.Likes,
			&item.ArticleLevel, &item.IsMasterpiece, &item.CreateTime); err != nil {
			return nil, err
		}
		item.URL = fmt.Sprintf("/detail/%d", item.ArticleID)
		out = append(out, item)
	}
	return out, rows.Err()
}

func intPtr(n int) *int { return &n }

func ragInPlaceholdersInts(ids []int) (string, []interface{}) {
	ph := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		ph[i] = "?"
		args[i] = id
	}
	return strings.Join(ph, ","), args
}
