// like_service 点赞业务逻辑。
package service

import (
	"context"

	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	blogrepo "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/blog/repo"
	blogevent "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/event"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent"
	"github.com/google/uuid"
)

// LikeService 点赞业务逻辑。
type LikeService struct {
	likes    *blogrepo.LikeRepo
	articles *blogrepo.ArticleRepo
	events   domainEventPublisher
}

// NewLikeService 构造 LikeService。
func NewLikeService(likes *blogrepo.LikeRepo, articles *blogrepo.ArticleRepo, publisher *blogevent.Publisher) *LikeService {
	return &LikeService{likes: likes, articles: articles, events: publisher}
}

// UpdateLike 点赞或取消点赞。
func (s *LikeService) UpdateLike(ctx context.Context, articleID, uid int, ip string, status interface{}) error {
	st := blogrepo.NormalizeLikeStatus(status)
	if st == "1" {
		count, err := s.likes.CountByArticleAndIP(ctx, articleID, ip)
		if err != nil {
			return err
		}
		if count > 20 {
			return errcode.WithMessage(errcode.NotFound, "该文章您点赞太过频繁了！")
		}
		likeUID := -999
		if uid > 0 {
			likeUID = uid
		}
		_, err = s.likes.Create(ctx, &ent.Like{
			ID:        uuid.NewString(),
			ArticleId: articleID,
			UID:       likeUID,
			IP:        ip,
			Status:    "1",
		})
		if err != nil {
			return err
		}
		if uid > 0 && s.events != nil {
			if art, err := s.articles.GetByID(ctx, articleID); err == nil && art.UID > 0 {
				s.events.Publish(ctx, blogevent.EventLikeCreated, blogevent.LikeCreatedPayload{
					UID: uid, ArticleID: articleID, AuthorUID: art.UID, DailyLimit: 10,
				})
			}
		}
	} else {
		row, err := s.likes.FindFirst(ctx, articleID, uid, ip)
		if err != nil {
			if ent.IsNotFound(err) {
				return nil
			}
			return err
		}
		if row != nil {
			return s.likes.DeleteByID(ctx, row.ID)
		}
	}
	return s.likes.SyncArticleLikes(ctx, articleID)
}

// CheckLiked 检查是否已点赞。
func (s *LikeService) CheckLiked(ctx context.Context, articleID, uid int) (map[string]interface{}, error) {
	if uid <= 0 {
		return map[string]interface{}{"liked": false}, nil
	}
	liked, err := s.likes.IsLiked(ctx, articleID, uid)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"liked": liked}, nil
}

// GetLikedArticleIDs 用户已点赞文章 ID 列表。
func (s *LikeService) GetLikedArticleIDs(ctx context.Context, uid int) (map[string]interface{}, error) {
	if uid <= 0 {
		return map[string]interface{}{"ids": []int{}}, nil
	}
	ids, err := s.likes.ListArticleIDsByUID(ctx, uid)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"ids": ids}, nil
}
