// Package level 用户与文章等级服务（单体副本，逻辑对齐 rpg-service）。
package level

import (
	"context"
	"fmt"

	"github.com/Jiang-Xia/blog-server-go/pkg/blogsvc"
	rpgconst "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/constants"
)

// ReputationGrant 给作者增加的声望。
type ReputationGrant struct {
	Amount int
	Reason string
}

// ArticleLevelResult AddArticleExp 副作用摘要。
type ArticleLevelResult struct {
	LeveledUp     bool
	IsMasterpiece bool
}

// AchievementTracker 成就进度追踪。
type AchievementTracker interface {
	TrackProgress(ctx context.Context, uid int, event string) error
}

// ReputationAdder 作者声望增减。
type ReputationAdder interface {
	AddReputation(ctx context.Context, uid, amount int, reason string) (int, error)
}

// ArticleLevelService 文章经验/等级/神作。
type ArticleLevelService struct {
	articles    blogsvc.ArticleRPGStore
	reputation  ReputationAdder
	achievement AchievementTracker
}

// NewArticleLevelService 构造 ArticleLevelService。
func NewArticleLevelService(
	articles blogsvc.ArticleRPGStore,
	reputation ReputationAdder,
	achievement AchievementTracker,
) *ArticleLevelService {
	return &ArticleLevelService{
		articles:    articles,
		reputation:  reputation,
		achievement: achievement,
	}
}

func articleLevelThreshold(level int) int {
	if level <= 1 {
		return 0
	}
	return level * (level - 1) * 20
}

func checkMasterpiece(level, exp int) bool {
	return level >= rpgconst.Economy.MasterpieceLevel || exp >= rpgconst.Economy.MasterpieceExp
}

// AddArticleExp 累加文章经验；reputationSkip 为 true 时不给作者加声望。
func (s *ArticleLevelService) AddArticleExp(
	ctx context.Context,
	articleID, amount, authorUID int,
	reputation *ReputationGrant,
	reputationSkip bool,
) (ArticleLevelResult, error) {
	var out ArticleLevelResult
	if s.articles == nil || articleID <= 0 || amount <= 0 {
		return out, nil
	}
	snap, err := s.articles.GetArticleRPGFields(ctx, articleID)
	if err != nil || snap == nil {
		return out, err
	}

	wasMasterpiece := snap.IsMasterpiece == 1
	snap.ArticleExp += amount
	for snap.ArticleExp >= articleLevelThreshold(snap.ArticleLevel+1) {
		snap.ArticleLevel++
		out.LeveledUp = true
	}

	if authorUID > 0 {
		snap.ReputationGained += amount
		if !reputationSkip {
			grant := reputation
			if grant == nil {
				grant = &ReputationGrant{
					Amount: rpgconst.Economy.ArticleViewReputation,
					Reason: "article_view",
				}
			}
			if grant.Amount > 0 && s.reputation != nil {
				_, _ = s.reputation.AddReputation(ctx, authorUID, grant.Amount, grant.Reason)
			}
		}
	}

	if checkMasterpiece(snap.ArticleLevel, snap.ArticleExp) {
		snap.IsMasterpiece = 1
		out.IsMasterpiece = true
	}

	if err := s.articles.UpdateArticleRPGFields(ctx, articleID, snap.ArticleExp, snap.ArticleLevel, snap.ReputationGained, snap.IsMasterpiece); err != nil {
		return out, err
	}

	authorID := authorUID
	if authorID <= 0 {
		authorID = snap.AuthorUID
	}
	if s.achievement != nil && authorID > 0 {
		if out.LeveledUp {
			_ = s.achievement.TrackProgress(ctx, authorID, "article_level_up")
		}
		if out.IsMasterpiece && !wasMasterpiece {
			_ = s.achievement.TrackProgress(ctx, authorID, "masterpiece")
		}
	}
	return out, nil
}

// AddTipTotal 累加打赏总额。
func (s *ArticleLevelService) AddTipTotal(ctx context.Context, articleID, amount int) error {
	if s.articles == nil || articleID <= 0 || amount <= 0 {
		return nil
	}
	return s.articles.AddArticleTipTotal(ctx, articleID, amount)
}

// ArticleTitle 返回文章标题。
func (s *ArticleLevelService) ArticleTitle(ctx context.Context, articleID int) string {
	if s.articles == nil || articleID <= 0 {
		return fmt.Sprintf("文章 #%d", articleID)
	}
	snap, err := s.articles.GetArticleRPGFields(ctx, articleID)
	if err != nil || snap == nil || snap.Title == "" {
		return fmt.Sprintf("文章 #%d", articleID)
	}
	return snap.Title
}
