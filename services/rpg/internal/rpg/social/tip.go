// Package social 文章打赏。
package social

import (
	"context"

	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/services/rpg/internal/articleport"
	"github.com/Jiang-Xia/blog-server-go/services/rpg/internal/event"
	"github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg/constants"
	"github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg/inventory"
	rpgnotify "github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg/notify"
	rpgrepo "github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg/repo"
	"github.com/Jiang-Xia/blog-server-go/services/rpg/ent"
)

// TipService 打赏业务。
type TipService struct {
	articles   articleport.ArticleReader
	repo       *rpgrepo.RpgRepo
	inventory  *inventory.Service
	reputation *ReputationService
	publisher  *event.Publisher
	notify     *rpgnotify.RpgNotifyService
}

// NewTipService 构造 TipService。
func NewTipService(
	articles articleport.ArticleReader,
	repo *rpgrepo.RpgRepo,
	inventory *inventory.Service,
	reputation *ReputationService,
	publisher *event.Publisher,
	notify *rpgnotify.RpgNotifyService,
) *TipService {
	return &TipService{
		articles:   articles,
		repo:       repo,
		inventory:  inventory,
		reputation: reputation,
		publisher:  publisher,
		notify:     notify,
	}
}

// TipArticle 打赏文章：扣款/入账同事务。
func (s *TipService) TipArticle(ctx context.Context, fromUID, articleID, amount int) (map[string]interface{}, error) {
	if amount < constants.Economy.TipMin {
		return nil, errcode.WithMessage(errcode.InvalidParam, "打赏金额过低")
	}
	article, err := s.articles.GetByID(ctx, articleID)
	if err != nil || article == nil || article.IsDelete {
		return nil, errcode.WithMessage(errcode.NotFound, "文章不存在")
	}
	if article.UID == fromUID {
		return nil, errcode.WithMessage(errcode.InvalidParam, "不能打赏自己的文章")
	}

	var authorBalance int
	err = s.repo.WithTx(ctx, func(tx *ent.Tx) error {
		if _, err := s.inventory.AdjustCurrency(ctx, fromUID, -amount, "tip"); err != nil {
			return err
		}
		bal, err := s.inventory.AdjustCurrency(ctx, article.UID, amount, "tip_received")
		if err != nil {
			return err
		}
		authorBalance = bal
		_, err = tx.RpgArticleTip.Create().
			SetUID(fromUID).
			SetArticleId(articleID).
			SetAuthorUid(article.UID).
			SetAmount(amount).
			Save(ctx)
		return err
	})
	if err != nil {
		return nil, err
	}

	repAmount := (amount + 1) / 2
	_, _ = s.reputation.AddReputation(ctx, article.UID, repAmount, "tip")

	if s.publisher != nil {
		s.publisher.Publish(ctx, event.EventArticleTipped, map[string]interface{}{
			"uid":       fromUID,
			"articleId": articleID,
			"authorUid": article.UID,
			"amount":    amount,
		})
	}
	if s.notify != nil {
		_ = s.notify.NotifyTipReceived(ctx, article.UID, rpgnotify.TipReceivedPayload{
			FromUID:   fromUID,
			ArticleID: articleID,
			Amount:    amount,
		})
	}
	return map[string]interface{}{
		"success":   true,
		"amount":    amount,
		"authorUid": article.UID,
		"balance":   authorBalance,
	}, nil
}
