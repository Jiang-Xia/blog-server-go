// like_repo 点赞表 Ent 读写。
package repo

import (
	"context"
	"strconv"

	"github.com/Jiang-Xia/blog-server-go/services/blog/ent"
	"github.com/Jiang-Xia/blog-server-go/services/blog/ent/like"
)

// LikeRepo 点赞表读写。
type LikeRepo struct {
	client *ent.Client
}

// NewLikeRepo 构造 LikeRepo。
func NewLikeRepo(client *ent.Client) *LikeRepo {
	return &LikeRepo{client: client}
}

// Create 写入点赞记录。
func (r *LikeRepo) Create(ctx context.Context, row *ent.Like) (*ent.Like, error) {
	b := r.client.Like.Create().
		SetID(row.ID).
		SetArticleId(row.ArticleId).
		SetUID(row.UID).
		SetIP(row.IP).
		SetStatus(row.Status)
	return b.Save(ctx)
}

// DeleteByID 删除点赞记录。
func (r *LikeRepo) DeleteByID(ctx context.Context, id string) error {
	_, err := r.client.Like.Delete().Where(like.IDEQ(id)).Exec(ctx)
	return err
}

// FindFirst 查找首条匹配记录（取消点赞用）。
func (r *LikeRepo) FindFirst(ctx context.Context, articleID, uid int, ip string) (*ent.Like, error) {
	q := r.client.Like.Query().Where(like.ArticleIdEQ(articleID))
	if uid > 0 {
		q = q.Where(like.UIDEQ(uid))
	} else if ip != "" {
		q = q.Where(like.IPEQ(ip))
	}
	return q.First(ctx)
}

// CountByArticleAndIP 统计 IP 对文章的点赞次数（防刷）。
func (r *LikeRepo) CountByArticleAndIP(ctx context.Context, articleID int, ip string) (int, error) {
	return r.client.Like.Query().
		Where(like.ArticleIdEQ(articleID), like.IPEQ(ip)).
		Count(ctx)
}

// CountByArticle 统计文章点赞总数。
func (r *LikeRepo) CountByArticle(ctx context.Context, articleID int) (int, error) {
	return r.client.Like.Query().
		Where(like.ArticleIdEQ(articleID), like.StatusEQ("1")).
		Count(ctx)
}

// IsLiked 登录用户是否已点赞。
func (r *LikeRepo) IsLiked(ctx context.Context, articleID, uid int) (bool, error) {
	n, err := r.client.Like.Query().
		Where(like.ArticleIdEQ(articleID), like.UIDEQ(uid), like.StatusEQ("1")).
		Count(ctx)
	return n > 0, err
}

// ListArticleIDsByUID 用户已点赞文章 ID 列表。
func (r *LikeRepo) ListArticleIDsByUID(ctx context.Context, uid int) ([]int, error) {
	rows, err := r.client.Like.Query().
		Where(like.UIDEQ(uid), like.StatusEQ("1")).
		Order(ent.Desc(like.FieldArticleId)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	seen := make(map[int]struct{})
	out := make([]int, 0, len(rows))
	for _, row := range rows {
		if _, ok := seen[row.ArticleId]; ok {
			continue
		}
		seen[row.ArticleId] = struct{}{}
		out = append(out, row.ArticleId)
	}
	return out, nil
}

// SyncArticleLikes 回写 article.likes 字段。
func (r *LikeRepo) SyncArticleLikes(ctx context.Context, articleID int) error {
	count, err := r.client.Like.Query().Where(like.ArticleIdEQ(articleID)).Count(ctx)
	if err != nil {
		return err
	}
	_, err = r.client.Article.UpdateOneID(articleID).SetLikes(count).Save(ctx)
	return err
}

// NormalizeLikeStatus 将前端 status 转为 1/0 字符串。
func NormalizeLikeStatus(status interface{}) string {
	switch v := status.(type) {
	case bool:
		if v {
			return "1"
		}
		return "0"
	case float64:
		if int(v) == 1 {
			return "1"
		}
		return "0"
	case int:
		if v == 1 {
			return "1"
		}
		return "0"
	case string:
		if v == "1" || v == "true" {
			return "1"
		}
		return "0"
	default:
		return "0"
	}
}

// ParseArticleIDInt 解析 articleId。
func ParseArticleIDInt(v interface{}) (int, error) {
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
