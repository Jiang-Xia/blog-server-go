// Package constants RPG 经济系统默认常量，对齐 Nest modules/rpg/constants/economy.ts。
package constants

// Economy RPG 经济与互动默认数值。
var Economy = struct {
	RegisterBonus int

	ArticleViewExp        int
	ArticleLikeExp        int
	ArticleCommentExp     int
	ArticleCollectExp     int
	ArticlePublishBaseExp int

	ArticleViewReputation int
	ReputationLike        int
	ReputationComment     int
	ReputationCollect     int

	MasterpieceLevel int
	MasterpieceExp   int

	CheerHP          int
	CheerDailyLimit  int
	EggCost          int
	EggHP            int
	EggDailyLimit    int
	FlowerCost       int
	FlowerReputation int
	FlowerDailyLimit int

	TipMin int

	LotteryEpicPityThreshold      int
	LotteryLegendaryPityThreshold int
	LotteryCurrencyCost           int

	RechargeRate    int
	RechargeMinYuan float64
	RechargeMaxYuan float64
}{
	RegisterBonus: 200,

	ArticleViewExp:        1,
	ArticleLikeExp:        2,
	ArticleCommentExp:     3,
	ArticleCollectExp:     5,
	ArticlePublishBaseExp: 10,

	ArticleViewReputation: 1,
	ReputationLike:        2,
	ReputationComment:     3,
	ReputationCollect:     5,

	MasterpieceLevel: 10,
	MasterpieceExp:   1000,

	CheerHP:          10,
	CheerDailyLimit:  3,
	EggCost:          15,
	EggHP:            -5,
	EggDailyLimit:    3,
	FlowerCost:       10,
	FlowerReputation: 3,
	FlowerDailyLimit: 5,

	TipMin: 1,

	LotteryEpicPityThreshold:      90,
	LotteryLegendaryPityThreshold: 180,
	LotteryCurrencyCost:           10,

	RechargeRate:    100,
	RechargeMinYuan: 0.01,
	RechargeMaxYuan: 200,
}

// RechargeAmounts 快捷充值面额（元）。
var RechargeAmounts = []float64{1, 5, 10, 50}

// SocialAction 用户间社交互动类型。
type SocialAction string

const (
	SocialActionCheer  SocialAction = "cheer"
	SocialActionEgg    SocialAction = "egg"
	SocialActionFlower SocialAction = "flower"
)
